// Copyright (c) 2020-2023, The OTNS Authors.
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

package main

import (
	"bufio"
	"context"
	"os"
	"time"

	"github.com/openthread/ot-ns/logger"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/prototext"

	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
)

var (
	unmarshalOptions = prototext.UnmarshalOptions{}
)

type grpcService struct {
	replayFile string
}

func (gs *grpcService) Visualize(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_VisualizeServer) error {
	defer logger.Infof("Visualize finished.")

	heartbeatEvent := &pb.VisualizeEvent{
		Type: &pb.VisualizeEvent_Heartbeat{Heartbeat: &pb.HeartbeatEvent{}},
	}

	visualizeDone := make(chan struct{})
	go gs.visualizeStream(stream, visualizeDone)

	heartbeatTicker := time.NewTicker(time.Second)
	defer heartbeatTicker.Stop()

waitloop:
	for {
		select {
		case <-heartbeatTicker.C:
			_ = stream.Send(heartbeatEvent)
		case <-visualizeDone:
			break waitloop
		case <-stream.Context().Done():
			break waitloop
		}
	}

	return nil
}

func (gs *grpcService) EnergyReport(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_EnergyReportServer) error {
	//TODO: implement energy report for replay, if it fits.
	var err error
	return err
}

func (gs *grpcService) Command(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error) {
	// TODO: implement some commands for replay (e.g. speed)
	return nil, errors.Errorf("can not run command on replay")
}

func (gs *grpcService) visualizeStream(stream pb.VisualizeGrpcService_VisualizeServer, visualizeDone chan struct{}) {
	defer func() {
		close(visualizeDone)

		err := recover()
		if err != nil && stream.Context().Err() == nil {
			logger.Errorf("visualization error: %v", err)
		}
	}()

	replay, err := os.Open(gs.replayFile)
	logger.PanicIfError(err)

	scanner := bufio.NewScanner(bufio.NewReader(replay))
	scanner.Split(bufio.ScanLines)

	startTime := time.Now()

	for scanner.Scan() {
		line := scanner.Text()

		logger.Infof("visualize: %#v", line)

		var entry pb.ReplayEntry
		err = unmarshalOptions.Unmarshal([]byte(line), &entry)
		logger.PanicIfError(err)

		playTime := startTime.Add(time.Duration(entry.Timestamp) * time.Microsecond)
		time.Sleep(time.Until(playTime))

		err = stream.Send(entry.Event)
		logger.PanicIfError(err)
	}
}
