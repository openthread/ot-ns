// Copyright (c) 2020-2024, The OTNS Authors.
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

package replay

import (
	"bufio"
	"os"
	"time"

	"google.golang.org/protobuf/encoding/prototext"

	"github.com/openthread/ot-ns/logger"
	visualize_grpc_pb "github.com/openthread/ot-ns/visualize/grpc/pb"
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

func (rep *Replay) Append(event *visualize_grpc_pb.VisualizeEvent) {
	timestamp := time.Since(rep.beginTime) / time.Microsecond
	entry := &visualize_grpc_pb.ReplayEntry{
		Event:     event,
		Timestamp: uint64(timestamp),
	}
	rep.pendingChan <- entry
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
