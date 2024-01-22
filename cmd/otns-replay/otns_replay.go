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

package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	"os"

	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/openthread/ot-ns/web"
	webSite "github.com/openthread/ot-ns/web/site"
	"google.golang.org/grpc"
)

var args struct {
	ReplayFile string
}

func parseArgs() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	args.ReplayFile = flag.Arg(0)
}

func main() {
	parseArgs()
	checkReplayFile(args.ReplayFile)
	logger.SetLevel(logger.InfoLevel)

	ctx := progctx.New(context.Background())

	server := grpc.NewServer(grpc.ReadBufferSize(1024*8), grpc.WriteBufferSize(1024*1024*1))
	gs := &grpcService{replayFile: args.ReplayFile}
	pb.RegisterVisualizeGrpcServiceServer(server, gs)

	lis, err := net.Listen("tcp", ":8999")
	logger.PanicIfError(err)

	go func() {
		siteAddr := ":8997"
		err := webSite.Serve(siteAddr)
		if err != http.ErrServerClosed {
			logger.PanicIfError(err)
		}
	}()

	go func() {
		web.ConfigWeb("", 8998, 8999, 8997)
		_ = web.OpenWeb(ctx, web.MainTab)
	}()

	err = server.Serve(lis)
	logger.Errorf("server quit: %v", err)
}

func checkReplayFile(filename string) {
	f, err := os.Open(filename)
	logger.PanicIfError(err)

	defer f.Close()
	fs, err := f.Stat()
	logger.PanicIfError(err)

	if fs.IsDir() {
		logger.Panicf("%s is not a valid replay", filename)
	}
}
