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
)

type NopVisualizer struct {
}

// NewNopVisualizer creates a new Visualizer that does nothing. It can be used as a base class to build
// an own Visualizer that implements a useful subset of the methods.
func NewNopVisualizer() Visualizer {
	return &NopVisualizer{}
}

func (nv *NopVisualizer) SetNetworkInfo(networkInfo NetworkInfo) {
}

func (nv *NopVisualizer) OnExtAddrChange(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) SetSpeed(speed float64) {
}

func (nv *NopVisualizer) SetParent(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) CountDown(duration time.Duration, text string) {
}

func (nv *NopVisualizer) ShowDemoLegend(x int, y int, title string) {
}

func (nv *NopVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) AddChildTable(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
}

func (nv *NopVisualizer) DeleteNode(id NodeId) {
}

func (nv *NopVisualizer) SetNodePos(nodeid NodeId, x, y, z int) {
}

func (nv *NopVisualizer) SetController(simController SimulationController) {

}

func (nv *NopVisualizer) Init() {

}

func (nv *NopVisualizer) Run() {

}

func (nv *NopVisualizer) Stop() {

}

func (nv *NopVisualizer) AddNode(nodeid NodeId, cfg *NodeConfig) {

}

func (nv *NopVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *MsgVisualizeInfo) {

}

func (nv *NopVisualizer) SetNodeRloc16(id NodeId, rloc16 uint16) {

}

func (nv *NopVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {

}

func (nv *NopVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {

}

func (nv *NopVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {

}

func (nv *NopVisualizer) AdvanceTime(ts uint64, speed float64) {

}

func (nv *NopVisualizer) OnNodeFail(NodeId) {

}

func (nv *NopVisualizer) OnNodeRecover(NodeId) {

}

func (nv *NopVisualizer) SetTitle(titleInfo TitleInfo) {

}

func (nv *NopVisualizer) UpdateNodesEnergy(node []*energy.NodeEnergy, timestamp uint64, updateView bool) {

}

func (nv *NopVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {

}

func (nv *NopVisualizer) UpdateNodeStats(nodeStatsInfo *NodeStatsInfo) {

}

func (nv *NopVisualizer) UpdateTimeWindowStats(timeWinStatsInfo *TimeWindowStatsInfo) {

}
