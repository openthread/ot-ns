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
	"github.com/openthread/ot-ns/visualize/grpc/pb"
)

type grpcStream struct {
	vizType visualizeStreamType
	pb.VisualizeGrpcService_VisualizeServer
}

type grpcEnergyStream struct {
	pb.VisualizeGrpcService_EnergyServer
}

// acceptsEvent determines if the stream will accept the given event, based on the viz-type of the stream.
func (gst *grpcStream) acceptsEvent(event *pb.VisualizeEvent) bool {
	if gst.vizType == meshTopologyVizType {
		return event.GetNodeStatsInfo() == nil
	} else if event.GetAdvanceTime() != nil ||
		event.GetNodeStatsInfo() != nil ||
		event.GetHeartbeat() != nil ||
		event.GetSetTitle() != nil {
		return true
	}
	return false
}

func (gst *grpcStream) close() {
}

func (gst *grpcEnergyStream) close() {
}

func newGrpcStream(vizType visualizeStreamType, stream pb.VisualizeGrpcService_VisualizeServer) *grpcStream {
	gst := &grpcStream{
		vizType:                              vizType,
		VisualizeGrpcService_VisualizeServer: stream,
	}
	return gst
}

func newGrpcEnergyStream(stream pb.VisualizeGrpcService_EnergyServer) *grpcEnergyStream {
	gst := &grpcEnergyStream{
		VisualizeGrpcService_EnergyServer: stream,
	}
	return gst
}
