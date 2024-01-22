// Copyright (c) 2022-2024, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package visualize_grpc

import (
	"context"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/visualize/grpc/pb"
)

type grpcServer struct {
	vis                      *grpcVisualizer
	server                   *grpc.Server
	address                  string
	visualizingStreams       map[*grpcStream]struct{}
	visualizingEnergyStreams map[*grpcEnergyStream]struct{}
	grpcClientAdded          chan string
}

func (gs *grpcServer) Visualize(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_VisualizeServer) error {
	var err error
	contextDone := stream.Context().Done()
	heartbeatEvent := &pb.VisualizeEvent{
		Type: &pb.VisualizeEvent_Heartbeat{Heartbeat: &pb.HeartbeatEvent{}},
	}
	var heartbeatTicker *time.Ticker

	gstream := newGrpcStream(stream)
	logger.Debugf("New gRPC visualize request received.")

	gs.vis.Lock()
	err = gs.prepareStream(gstream)
	if err != nil {
		gs.vis.Unlock()
		goto exit
	}
	gs.visualizingStreams[gstream] = struct{}{}
	// if web.OpenWeb goroutine is waiting for a new client, then notify it.
	select {
	case gs.grpcClientAdded <- req.String():
		break
	default:
		break
	}
	gs.vis.Unlock()

	defer gs.disposeStream(gstream)

	heartbeatTicker = time.NewTicker(time.Second)
	defer heartbeatTicker.Stop()

	for {
		select {
		case <-heartbeatTicker.C:
			err = stream.Send(heartbeatEvent)
			if err != nil {
				goto exit
			}
		case <-contextDone:
			err = stream.Context().Err()
			goto exit
		}
	}

exit:
	logger.Debugf("Visualize stream exit: %v", err)
	return err
}

func (gs *grpcServer) EnergyReport(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_EnergyReportServer) error {
	var err error
	contextDone := stream.Context().Done()

	//TODO: do we need a heartbeat and a idle checker here too?

	gstream := newGrpcEnergyStream(stream)
	logger.Debugf("New energy report request got.")

	gs.visualizingEnergyStreams[gstream] = struct{}{}
	defer gs.disposeEnergyStream(gstream)

	energyHist := gs.vis.energyAnalyser.GetNetworkEnergyHistory()
	energyHistByNodes := gs.vis.energyAnalyser.GetEnergyHistoryByNodes()
	for i := 0; i < len(energyHistByNodes); i++ {
		gs.vis.UpdateNodesEnergy(energyHistByNodes[i], energyHist[i].Timestamp, ((i + 1) == len(energyHistByNodes)))
	}

	//Wait for the first event
	<-contextDone
	err = stream.Context().Err()

	logger.Debugf("energy report stream exit: %v", err)
	return err
}

func (gs *grpcServer) Command(ctx context.Context, req *pb.CommandRequest) (*pb.CommandResponse, error) {
	output, err := gs.vis.simctrl.Command(req.Command)
	return &pb.CommandResponse{
		Output: output,
	}, err
}

func (gs *grpcServer) Run() error {
	lis, err := net.Listen("tcp", gs.address)
	if err != nil {
		return err
	}
	logger.Infof("gRPC visualizer server serving on %s ...", lis.Addr())
	return gs.server.Serve(lis)
}

func (gs *grpcServer) SendEvent(event *pb.VisualizeEvent, trivial bool) {
	for stream := range gs.visualizingStreams {
		_ = stream.Send(event)
	}
}

func (gs *grpcServer) SendEnergyEvent(event *pb.NetworkEnergyEvent) {
	for stream := range gs.visualizingEnergyStreams {
		_ = stream.Send(event)
	}
}

func (gs *grpcServer) stop() {
	for stream := range gs.visualizingStreams {
		stream.close()
	}
	gs.server.Stop()
}

func (gs *grpcServer) disposeStream(stream *grpcStream) {
	gs.vis.Lock()
	delete(gs.visualizingStreams, stream)
	gs.vis.Unlock()
	stream.close()
}

func (gs *grpcServer) disposeEnergyStream(stream *grpcEnergyStream) {
	gs.vis.Lock()
	delete(gs.visualizingEnergyStreams, stream)
	gs.vis.Unlock()
	stream.close()
}

func (gs *grpcServer) prepareStream(stream *grpcStream) error {
	return gs.vis.prepareStream(stream)
}

func newGrpcServer(vis *grpcVisualizer, address string, chanNewClientNotifier chan string) *grpcServer {
	server := grpc.NewServer(grpc.ReadBufferSize(1024*8), grpc.WriteBufferSize(1024*1024*1))
	gs := &grpcServer{
		vis:                      vis,
		server:                   server,
		address:                  address,
		visualizingStreams:       map[*grpcStream]struct{}{},
		visualizingEnergyStreams: map[*grpcEnergyStream]struct{}{},
		grpcClientAdded:          chanNewClientNotifier,
	}
	pb.RegisterVisualizeGrpcServiceServer(server, gs)
	return gs
}
