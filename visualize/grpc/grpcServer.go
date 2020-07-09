// Copyright (c) 2020, The OTNS Authors.
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
	"sync"
	"time"

	"github.com/openthread/ot-ns/types"
	. "github.com/openthread/ot-ns/types"

	"github.com/simonlingoogle/go-simplelogger"

	pb "github.com/openthread/ot-ns/visualize/grpc/pb"

	"google.golang.org/grpc"
)

type grpcServer struct {
	vis                    *grpcVisualizer
	server                 *grpc.Server
	address                string
	visualizingStreamsLock sync.Mutex
	visualizingStreams     map[*grpcStream]struct{}
}

func (gs *grpcServer) Visualize(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_VisualizeServer) error {
	var err error
	contextDone := stream.Context().Done()
	heartbeatEvent := &pb.VisualizeEvent{
		Type: &pb.VisualizeEvent_Heartbeat{Heartbeat: &pb.HeartbeatEvent{}},
	}
	var heartbeatTicker *time.Ticker

	gstream := newGrpcStream(stream)
	simplelogger.Infof("New visualize request got.")
	err = gs.prepareStream(gstream)
	simplelogger.Infof("Visualize stream prepared: error=%v", err)
	if err != nil {
		goto exit
	}

	gs.visualizingStreamsLock.Lock()
	gs.visualizingStreams[gstream] = struct{}{}
	gs.visualizingStreamsLock.Unlock()

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
	simplelogger.Infof("Visualize stream exit: %v", err)
	return err
}

func (gs *grpcServer) CtrlAddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.Empty, error) {
	err := gs.vis.simctrl.CtrlAddNode(int(req.X), int(req.Y), req.IsRouter, types.NodeMode{
		RxOnWhenIdle:       req.Mode.RxOnWhenIdle,
		SecureDataRequests: req.Mode.SecureDataRequests,
		FullThreadDevice:   req.Mode.FullThreadDevice,
		FullNetworkData:    req.Mode.FullNetworkData,
	}, NodeId(req.NodeId))
	return &pb.Empty{}, err
}

func (gs *grpcServer) CtrlDeleteNode(ctx context.Context, req *pb.DeleteNodeRequest) (*pb.Empty, error) {
	err := gs.vis.simctrl.CtrlDeleteNode(NodeId(req.NodeId))
	return &pb.Empty{}, err
}

func (gs *grpcServer) CtrlMoveNodeTo(ctx context.Context, req *pb.MoveNodeToRequest) (*pb.Empty, error) {
	err := gs.vis.simctrl.CtrlMoveNodeTo(NodeId(req.NodeId), int(req.X), int(req.Y))
	return &pb.Empty{}, err
}

func (gs *grpcServer) CtrlSetNodeFailed(ctx context.Context, req *pb.SetNodeFailedRequest) (*pb.Empty, error) {
	err := gs.vis.simctrl.CtrlSetNodeFailed(NodeId(req.NodeId), req.Failed)
	return &pb.Empty{}, err
}

func (gs *grpcServer) CtrlSetSpeed(ctx context.Context, req *pb.SetSpeedRequest) (*pb.Empty, error) {
	err := gs.vis.simctrl.CtrlSetSpeed(req.Speed)
	return &pb.Empty{}, err
}

func (gs *grpcServer) Run() error {
	lis, err := net.Listen("tcp", gs.address)
	simplelogger.PanicIfError(err)
	simplelogger.Infof("gRPC visualizer serving on %s ...", lis.Addr())
	return gs.server.Serve(lis)
}

func (gs *grpcServer) SendEvent(event *pb.VisualizeEvent, trivial bool) {
	streams := gs.getAllStreams()
	for _, stream := range streams {
		_ = stream.Send(event)
	}
}

func (gs *grpcServer) getAllStreams() []*grpcStream {
	gs.visualizingStreamsLock.Lock()
	defer gs.visualizingStreamsLock.Unlock()

	streams := make([]*grpcStream, 0, len(gs.visualizingStreams))
	for stream := range gs.visualizingStreams {
		streams = append(streams, stream)
	}
	return streams
}

func (gs *grpcServer) stop() {
	streams := gs.getAllStreams()

	for _, stream := range streams {
		stream.close()
	}

	gs.server.Stop()
}

func (gs *grpcServer) disposeStream(stream *grpcStream) {
	gs.visualizingStreamsLock.Lock()
	delete(gs.visualizingStreams, stream)
	gs.visualizingStreamsLock.Unlock()
	stream.close()
}

func (gs *grpcServer) prepareStream(stream *grpcStream) error {
	return gs.vis.prepareStream(stream)
}

func newGrpcServer(vis *grpcVisualizer, address string) *grpcServer {
	server := grpc.NewServer(grpc.ReadBufferSize(1024*8), grpc.WriteBufferSize(1024*1024*1))
	gs := &grpcServer{
		vis:                vis,
		server:             server,
		address:            address,
		visualizingStreams: map[*grpcStream]struct{}{},
	}
	pb.RegisterVisualizeGrpcServiceServer(server, gs)
	return gs
}
