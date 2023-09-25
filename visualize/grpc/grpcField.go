// Copyright (c) 2020-2022, The OTNS Authors.
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
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type grpcField struct {
	nodes       map[NodeId]*grpcNode
	curTime     uint64
	curSpeed    float64
	speed       float64
	titleInfo   visualize.TitleInfo
	networkInfo visualize.NetworkInfo
}

func (f *grpcField) addNode(id NodeId, x int, y int, radioRange int) *grpcNode {
	logger.AssertNil(f.nodes[id])
	gn := newGprcNode(id, x, y, radioRange)
	f.nodes[id] = gn
	return gn
}

func (f *grpcField) setNodeRloc16(id NodeId, rloc16 uint16) {
	f.nodes[id].rloc16 = rloc16
}

func (f *grpcField) setNodeRole(id NodeId, role OtDeviceRole) {
	f.nodes[id].role = role
}

func (f *grpcField) setNodeMode(id NodeId, mode NodeMode) {
	f.nodes[id].mode = mode
}

func (f *grpcField) setNodePartitionId(id NodeId, parid uint32) {
	f.nodes[id].partitionId = parid
}

func (f *grpcField) advanceTime(ts uint64, speed float64) bool {
	hasChanged := f.curTime != ts || f.curSpeed != speed
	f.curTime = ts
	f.curSpeed = speed
	return hasChanged
}

func (f *grpcField) onNodeFail(nodeid NodeId) {
	f.nodes[nodeid].failed = true
}

func (f *grpcField) onNodeRecover(id NodeId) {
	f.nodes[id].failed = false
}

func (f *grpcField) setNodePos(id NodeId, x int, y int) {
	node := f.nodes[id]
	node.x = x
	node.y = y
}

func (f *grpcField) deleteNode(id NodeId) {
	delete(f.nodes, id)
}

func (f *grpcField) setParent(id NodeId, extaddr uint64) {
	f.nodes[id].parent = extaddr
}

func (f *grpcField) addRouterTable(id NodeId, extaddr uint64) {
	f.nodes[id].routerTable[extaddr] = struct{}{}
}

func (f *grpcField) removeRouterTable(id NodeId, extaddr uint64) {
	delete(f.nodes[id].routerTable, extaddr)
}

func (f *grpcField) addChildTable(id NodeId, extaddr uint64) {
	f.nodes[id].childTable[extaddr] = struct{}{}
}

func (f *grpcField) removeChildTable(id NodeId, extaddr uint64) {
	delete(f.nodes[id].childTable, extaddr)
}

func (f *grpcField) setSpeed(speed float64) {
	f.speed = speed
}

func (f *grpcField) onExtAddrChange(id NodeId, extaddr uint64) {
	f.nodes[id].extaddr = extaddr
}

func (f *grpcField) setTitleInfo(info visualize.TitleInfo) {
	f.titleInfo = info
}

func newGrpcField() *grpcField {
	gf := &grpcField{
		nodes:       map[NodeId]*grpcNode{},
		curSpeed:    1,
		speed:       1,
		networkInfo: visualize.DefaultNetworkInfo(),
	}
	return gf
}
