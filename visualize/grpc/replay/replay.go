package replay

import (
	"bufio"
	"os"
	"time"

	"github.com/openthread/ot-ns/logger"
	visualize_grpc_pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"google.golang.org/protobuf/encoding/prototext"
)

var (
	marshalOptions = prototext.MarshalOptions{
		Multiline: false,
	}
)

type Replay struct {
	f              *os.File
	fileWriter     *bufio.Writer
	pendingChan    chan *visualize_grpc_pb.ReplayEntry
	fileWriterDone chan struct{}
	beginTime      time.Time
}

func (rep *Replay) Append(event *visualize_grpc_pb.VisualizeEvent, trivial bool) {
	timestamp := time.Since(rep.beginTime) / time.Microsecond
	entry := &visualize_grpc_pb.ReplayEntry{
		Event:     event,
		Timestamp: uint64(timestamp),
	}

	if !trivial {
		rep.pendingChan <- entry
	} else {
		select {
		case rep.pendingChan <- entry:
			break
		default:
			logger.Warnf("replay generation routine is busy, dropping trivial events ...")
			break
		}
	}
}

func (rep *Replay) Close() {
	close(rep.pendingChan)
	<-rep.fileWriterDone
}

func (rep *Replay) fileWriterRoutine() {
	var err error

	defer func() {
		close(rep.fileWriterDone)

		if err != nil {
			logger.Errorf("replay write routine quit unexpectedly: %v", err)
		}
	}()

	defer rep.f.Close()

	for e := range rep.pendingChan {
		var data []byte

		if data, err = marshalOptions.Marshal(e); err != nil {
			break
		}

		if _, err = rep.fileWriter.Write(data); err != nil {
			break
		}

		if _, err = rep.fileWriter.Write([]byte{'\n'}); err != nil {
			break
		}
	}

	err = rep.fileWriter.Flush()
}

func NewReplay(filename string) *Replay {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	logger.PanicIfError(err)

	rep := &Replay{
		f:              f,
		fileWriter:     bufio.NewWriterSize(f, 8192),
		pendingChan:    make(chan *visualize_grpc_pb.ReplayEntry, 10000),
		fileWriterDone: make(chan struct{}),
		beginTime:      time.Now(),
	}

	go rep.fileWriterRoutine()

	return rep
}
