package main

import (
	"context"
	"flag"
	"net"
	"os"

	"github.com/openthread/ot-ns/progctx"
	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/openthread/ot-ns/web"
	webSite "github.com/openthread/ot-ns/web/site"
	"github.com/simonlingoogle/go-simplelogger"
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
	simplelogger.SetLevel(simplelogger.InfoLevel)

	ctx := progctx.New(context.Background())

	server := grpc.NewServer(grpc.ReadBufferSize(1024*8), grpc.WriteBufferSize(1024*1024*1))
	gs := &grpcService{replayFile: args.ReplayFile}
	pb.RegisterVisualizeGrpcServiceServer(server, gs)

	lis, err := net.Listen("tcp", ":8999")
	simplelogger.PanicIfError(err)

	go func() {
		siteAddr := ":8997"
		err := webSite.Serve(siteAddr)
		simplelogger.PanicIfError(err)
	}()

	go func() {
		web.ConfigWeb("", 8998, 8999, 8997)
		_ = web.OpenWeb(ctx)
	}()

	err = server.Serve(lis)
	simplelogger.Errorf("server quit: %v", err)
}

func checkReplayFile(filename string) {
	f, err := os.Open(filename)
	simplelogger.PanicIfError(err)

	defer f.Close()
	fs, err := f.Stat()
	simplelogger.PanicIfError(err)

	if fs.IsDir() {
		simplelogger.Panicf("%s is not a valid replay", filename)
	}
}
