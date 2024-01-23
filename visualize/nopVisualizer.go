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

package visualize

import (
	"time"

	"github.com/openthread/ot-ns/energy"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize/grpc/pb"
)

type nopVisualizer struct{}

// NewNopVisualizer creates a new Visualizer that does nothing.
func NewNopVisualizer() Visualizer {
	return nopVisualizer{}
}

func (nv nopVisualizer) SetNetworkInfo(networkInfo NetworkInfo) {
}

func (nv nopVisualizer) OnExtAddrChange(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) SetSpeed(speed float64) {
}

func (nv nopVisualizer) SetParent(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) CountDown(duration time.Duration, text string) {
}

func (nv nopVisualizer) ShowDemoLegend(x int, y int, title string) {
}

func (nv nopVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) AddChildTable(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
}

func (nv nopVisualizer) DeleteNode(id NodeId) {
}

func (nv nopVisualizer) SetNodePos(nodeid NodeId, x, y, z int) {
}

func (nv nopVisualizer) SetController(ctrl SimulationController) {
}

func (nv nopVisualizer) Init() {

}

func (nv nopVisualizer) Run() {
	for {
		time.Sleep(time.Hour)
	}
}

func (nv nopVisualizer) Stop() {

}

func (nv nopVisualizer) AddNode(nodeid NodeId, cfg *NodeConfig) {

}

func (nv nopVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *MsgVisualizeInfo) {

}

func (nv nopVisualizer) SetNodeRloc16(id NodeId, rloc16 uint16) {

}

func (nv nopVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {

}

func (nv nopVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {

}

func (nv nopVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {

}

func (nv nopVisualizer) AdvanceTime(ts uint64, speed float64) {

}

func (nv nopVisualizer) OnNodeFail(NodeId) {

}

func (nv nopVisualizer) OnNodeRecover(NodeId) {

}

func (nv nopVisualizer) SetTitle(titleInfo TitleInfo) {

}

func (nv nopVisualizer) UpdateNodesEnergy(node []*pb.NodeEnergy, timestamp uint64, updateView bool) {

}

func (nv nopVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {

}
