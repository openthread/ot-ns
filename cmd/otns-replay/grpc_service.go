package main

import (
	"bufio"
	"context"
	"os"
	"time"

	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	unmarshalOptions = prototext.UnmarshalOptions{}
)

type grpcService struct {
	replayFile string
}

func (gs *grpcService) Visualize(req *pb.VisualizeRequest, stream pb.VisualizeGrpcService_VisualizeServer) error {
	defer simplelogger.Infof("Visualize finished.")

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

func (gs *grpcService) Command(context.Context, *pb.CommandRequest) (*pb.CommandResponse, error) {
	// TODO: implement some commands for replay (e.g. speed)
	return nil, errors.Errorf("can not run command on replay")
}

func (gs *grpcService) visualizeStream(stream pb.VisualizeGrpcService_VisualizeServer, visualizeDone chan struct{}) {
	defer func() {
		close(visualizeDone)

		err := recover()
		if err != nil && stream.Context().Err() == nil {
			simplelogger.Errorf("visualization error: %v", err)
		}
	}()

	replay, err := os.Open(gs.replayFile)
	simplelogger.PanicIfError(err)

	scanner := bufio.NewScanner(bufio.NewReader(replay))
	scanner.Split(bufio.ScanLines)

	startTime := time.Now()

	for scanner.Scan() {
		line := scanner.Text()

		simplelogger.Infof("visualize: %#v", line)

		var entry pb.ReplayEntry
		err = unmarshalOptions.Unmarshal([]byte(line), &entry)
		simplelogger.PanicIfError(err)

		playTime := startTime.Add(time.Duration(entry.Timestamp) * time.Microsecond)
		time.Sleep(time.Until(playTime))

		err = stream.Send(entry.Event)
		simplelogger.PanicIfError(err)
	}
}
