// Copyright (c) 2020-2026, The OTNS Authors.
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
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

// grpcField tracks simulation topology and node properties. This information is used to recreate the GUI
// after a reload event, or to selectively render GUI elements based on user interaction where the rendering
// logic runs on the server (Go) side.
type grpcField struct {
	nodes         map[NodeId]*grpcNode
	extAddrMap    map[uint64]NodeId
	curTime       uint64
	curSpeed      float64
	speed         float64
	titleInfo     visualize.TitleInfo
	networkInfo   visualize.NetworkInfo
	nodeStatsInfo visualize.NodeStatsInfo
}

func (f *grpcField) addNode(id NodeId, cfg *NodeConfig) *grpcNode {
	logger.AssertNil(f.nodes[id])
	gn := newGprcNode(id, cfg)
	logger.AssertNotNil(gn.extaddr)
	logger.AssertTrue(gn.extaddr > 0)
	f.nodes[id] = gn
	f.extAddrMap[gn.extaddr] = id
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

func (f *grpcField) send(srcid NodeId, dstid NodeId, mvInfo *visualize.MsgVisualizeInfo) {
	// remember properties of (ongoing) frame transmitted by source, for later use
	f.nodes[srcid].curTxPower = mvInfo.PowerDbm
}

func (f *grpcField) setNodePartitionId(id NodeId, parid uint32) {
	f.nodes[id].partitionId = parid
}

func (f *grpcField) setNodeVersion(id NodeId, version string) {
	f.nodes[id].version = version
}

func (f *grpcField) setNodeCommit(id NodeId, commit string) {
	f.nodes[id].commit = commit
}

func (f *grpcField) setNodeThreadVersion(id NodeId, version uint16) {
	f.nodes[id].threadVersion = version
}

func (f *grpcField) advanceTime(ts uint64, speed float64) bool {
	hasChanged := f.curTime != ts || f.curSpeed != speed
	f.curTime = ts
	f.curSpeed = speed
	return hasChanged
}

func (f *grpcField) onNodeFail(id NodeId) {
	f.nodes[id].failed = true
}

func (f *grpcField) onNodeRecover(id NodeId) {
	f.nodes[id].failed = false
}

func (f *grpcField) setNodePos(id NodeId, x, y, z int) {
	node := f.nodes[id]
	node.x = x
	node.y = y
	node.z = z
}

func (f *grpcField) deleteNode(id NodeId) {
	extaddr := f.nodes[id].extaddr
	for _, node := range f.nodes {
		delete(node.neighborInfo, extaddr)
	}
	delete(f.extAddrMap, extaddr)
	delete(f.nodes, id)
}

func (f *grpcField) setParent(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	node.parent = extaddr
	node.setLinked(extaddr, true)
}

func (f *grpcField) addRouterTable(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	node.routerTable[extaddr] = struct{}{}
	node.setLinked(extaddr, true)
}

func (f *grpcField) removeRouterTable(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	delete(node.routerTable, extaddr)
	node.setLinked(extaddr, false)
}

func (f *grpcField) addChildTable(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	node.childTable[extaddr] = struct{}{}
	node.setLinked(extaddr, true)
}

func (f *grpcField) removeChildTable(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	delete(node.childTable, extaddr)
	node.setLinked(extaddr, false)
}

func (f *grpcField) setLinkStats(id NodeId, opt visualize.LinkStatsOptions) {
	node := f.nodes[id]
	node.linkStatsOpt = opt
}

func (f *grpcField) checkLinkStatUpdates(node *grpcNode, peerExtAddr uint64) {
	nbStats := node.getNeighborInfo(peerExtAddr)
	if nbStats.isLinked && node.linkStatsOpt.Visible {
		//
	}
}

func (f *grpcField) getNodeLinkedPeers(id NodeId) []uint64 {
	node := f.nodes[id]
	peerExtAddrs := make([]uint64, 0, len(node.neighborInfo)+5)
	for peerExtAddr, nbInfo := range node.neighborInfo {
		if nbInfo.isLinked {
			peerExtAddrs = append(peerExtAddrs, peerExtAddr)
		} else {
			peerNodeId := f.extAddrMap[peerExtAddr]
			if peerNbInfo, ok := f.nodes[peerNodeId].neighborInfo[node.extaddr]; ok && peerNbInfo.isLinked {
				peerExtAddrs = append(peerExtAddrs, peerExtAddr)
			}
		}
	}
	return peerExtAddrs
}

func (f *grpcField) setSpeed(speed float64) {
	f.speed = speed
}

func (f *grpcField) onExtAddrChange(id NodeId, extaddr uint64) {
	node := f.nodes[id]
	oldExtaddr := node.extaddr
	if oldExtaddr != InvalidExtAddr {
		for _, pnode := range f.nodes {
			delete(pnode.neighborInfo, oldExtaddr)
		}
	}
	delete(f.extAddrMap, oldExtaddr)
	node.extaddr = extaddr
	f.extAddrMap[extaddr] = id
}

// onRadioFrameDispatch updates Tx/Rx frame statistics that can be potentially visualized.
// isSrcChange/isDstChange return true if respectively a link property at the source or destination node visibly
// changed.
func (f *grpcField) onRadioFrameDispatch(srcid NodeId, dstid NodeId, evt *event.Event) (isSrcChange bool, isDstChange bool) {
	if evt.Type == event.EventTypeRadioRxDone {
		// keep track of stats for successfully dispatched frames to/from neighbor node
		src := f.nodes[srcid]
		dst := f.nodes[dstid]

		nbInfo := src.getNeighborInfo(dst.extaddr)
		dstNbInfo := dst.getNeighborInfo(src.extaddr)
		isLinked := nbInfo.isLinked || dstNbInfo.isLinked

		if nbInfo.lastTxPower != src.curTxPower {
			nbInfo.lastTxPower = src.curTxPower
			isSrcChange = src.linkStatsOpt.Visible && src.linkStatsOpt.TxPower && isLinked
		}

		newRssi := evt.RadioCommData.PowerDbm
		if dstNbInfo.lastRssi != newRssi {
			dstNbInfo.lastRssi = newRssi
			isDstChange = dst.linkStatsOpt.Visible && dst.linkStatsOpt.RxRssi && isLinked
		}
	}
	return
}

func (f *grpcField) setTitleInfo(info visualize.TitleInfo) {
	f.titleInfo = info
}

func (f *grpcField) setNodeStatsInfo(info visualize.NodeStatsInfo) {
	f.nodeStatsInfo = info
}

func newGrpcField() *grpcField {
	gf := &grpcField{
		nodes:         map[NodeId]*grpcNode{},
		extAddrMap:    map[uint64]NodeId{},
		curTime:       0,
		curSpeed:      1.0,
		speed:         1.0,
		titleInfo:     visualize.DefaultTitleInfo(),
		networkInfo:   visualize.DefaultNetworkInfo(),
		nodeStatsInfo: visualize.DefaultNodeStatsInfo(),
	}
	return gf
}
