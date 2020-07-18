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

	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/simonlingoogle/go-simplelogger"
)

type grpcSimCtrl struct {
	ctx    context.Context
	client pb.VisualizeGrpcServiceClient
}

func (gsc *grpcSimCtrl) CtrlSetTitle(titleInfo visualize.TitleInfo) error {
	_, err := gsc.client.CtrlSetTitle(gsc.ctx, &pb.SetTitleEvent{
		Title:    titleInfo.Title,
		X:        int32(titleInfo.X),
		Y:        int32(titleInfo.Y),
		FontSize: int32(titleInfo.FontSize),
	})
	return err
}

func (gsc *grpcSimCtrl) CtrlSetSpeed(speed float64) error {
	_, err := gsc.client.CtrlSetSpeed(gsc.ctx, &pb.SetSpeedRequest{
		Speed: speed,
	})
	return err
}

func (gsc *grpcSimCtrl) CtrlSetNodeFailed(nodeid NodeId, failed bool) error {
	_, err := gsc.client.CtrlSetNodeFailed(gsc.ctx, &pb.SetNodeFailedRequest{
		NodeId: int32(nodeid),
		Failed: failed,
	})
	return err
}

func (gsc *grpcSimCtrl) CtrlAddNode(x, y int, router bool, mode NodeMode, nodeid NodeId) error {
	simplelogger.AssertTrue(nodeid == InvalidNodeId || nodeid > 0)

	_, err := gsc.client.CtrlAddNode(gsc.ctx, &pb.AddNodeRequest{
		X:        int32(x),
		Y:        int32(y),
		IsRouter: router,
		Mode: &pb.NodeMode{
			RxOnWhenIdle:       mode.RxOnWhenIdle,
			SecureDataRequests: mode.SecureDataRequests,
			FullThreadDevice:   mode.FullThreadDevice,
			FullNetworkData:    mode.FullNetworkData,
		},
		NodeId: uint32(nodeid),
	})
	return err
}

func (gsc *grpcSimCtrl) CtrlMoveNodeTo(nodeid NodeId, x, y int) error {
	_, err := gsc.client.CtrlMoveNodeTo(gsc.ctx, &pb.MoveNodeToRequest{
		NodeId: int32(nodeid),
		X:      int32(x),
		Y:      int32(y),
	})
	return err
}

func (gsc *grpcSimCtrl) CtrlDeleteNode(nodeid NodeId) error {
	_, err := gsc.client.CtrlDeleteNode(gsc.ctx, &pb.DeleteNodeRequest{
		NodeId: int32(nodeid),
	})
	return err
}

func NewGrpcSimulationController(ctx context.Context, client pb.VisualizeGrpcServiceClient) visualize.SimulationController {
	return &grpcSimCtrl{
		ctx:    ctx,
		client: client,
	}
}
