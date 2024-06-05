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

package visualize_multi

import (
	"time"

	"github.com/openthread/ot-ns/energy"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type MultiVisualizer struct {
	vs []visualize.Visualizer
}

// NewMultiVisualizer creates a new Visualizer that multiplexes to multiple Visualizers.
func NewMultiVisualizer(vs ...visualize.Visualizer) *MultiVisualizer {
	return &MultiVisualizer{vs: vs}
}

func (mv *MultiVisualizer) AddVisualizer(vs ...visualize.Visualizer) {
	mv.vs = append(mv.vs, vs...)
}

func (mv *MultiVisualizer) SetNetworkInfo(networkInfo visualize.NetworkInfo) {
	for _, v := range mv.vs {
		v.SetNetworkInfo(networkInfo)
	}
}

func (mv *MultiVisualizer) OnExtAddrChange(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.OnExtAddrChange(id, extaddr)
	}
}

func (mv *MultiVisualizer) SetSpeed(speed float64) {
	for _, v := range mv.vs {
		v.SetSpeed(speed)
	}
}

func (mv *MultiVisualizer) Init() {
	for _, v := range mv.vs {
		v.Init()
	}
}

func (mv *MultiVisualizer) Run() {
	for i := 1; i < len(mv.vs); i++ {
		go mv.vs[i].Run()
	}
	mv.vs[0].Run()
}

func (mv *MultiVisualizer) Stop() {
	for _, v := range mv.vs {
		v.Stop()
	}
}

func (mv *MultiVisualizer) AddNode(nodeid NodeId, cfg *NodeConfig) {
	for _, v := range mv.vs {
		v.AddNode(nodeid, cfg)
	}
}

func (mv *MultiVisualizer) SetNodeRloc16(nodeid NodeId, rloc16 uint16) {
	for _, v := range mv.vs {
		v.SetNodeRloc16(nodeid, rloc16)
	}
}

func (mv *MultiVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {
	for _, v := range mv.vs {
		v.SetNodeRole(nodeid, role)
	}
}

func (mv *MultiVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {
	for _, v := range mv.vs {
		v.SetNodeMode(nodeid, mode)
	}
}

func (mv *MultiVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *visualize.MsgVisualizeInfo) {
	for _, v := range mv.vs {
		v.Send(srcid, dstid, mvinfo)
	}
}

func (mv *MultiVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {
	for _, v := range mv.vs {
		v.SetNodePartitionId(nodeid, parid)
	}
}

func (mv *MultiVisualizer) AdvanceTime(ts uint64, speed float64) {
	for _, v := range mv.vs {
		v.AdvanceTime(ts, speed)
	}
}

func (mv *MultiVisualizer) OnNodeFail(nodeid NodeId) {
	for _, v := range mv.vs {
		v.OnNodeFail(nodeid)
	}
}

func (mv *MultiVisualizer) OnNodeRecover(nodeid NodeId) {
	for _, v := range mv.vs {
		v.OnNodeRecover(nodeid)
	}
}

func (mv *MultiVisualizer) SetController(ctrl visualize.SimulationController) {
	for _, v := range mv.vs {
		v.SetController(ctrl)
	}
}

func (mv *MultiVisualizer) SetNodePos(nodeid NodeId, x, y, z int) {
	for _, v := range mv.vs {
		v.SetNodePos(nodeid, x, y, z)
	}
}

func (mv *MultiVisualizer) DeleteNode(id NodeId) {
	for _, v := range mv.vs {
		v.DeleteNode(id)
	}
}

func (mv *MultiVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.AddRouterTable(id, extaddr)
	}
}

func (mv *MultiVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.RemoveRouterTable(id, extaddr)
	}
}

func (mv *MultiVisualizer) AddChildTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.AddChildTable(id, extaddr)
	}
}

func (mv *MultiVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.RemoveChildTable(id, extaddr)
	}
}

func (mv *MultiVisualizer) ShowDemoLegend(x int, y int, title string) {
	for _, v := range mv.vs {
		v.ShowDemoLegend(x, y, title)
	}
}

func (mv *MultiVisualizer) CountDown(duration time.Duration, text string) {
	for _, v := range mv.vs {
		v.CountDown(duration, text)
	}
}

func (mv *MultiVisualizer) SetParent(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.SetParent(id, extaddr)
	}
}

func (mv *MultiVisualizer) SetTitle(titleInfo visualize.TitleInfo) {
	for _, v := range mv.vs {
		v.SetTitle(titleInfo)
	}
}

func (mv *MultiVisualizer) UpdateNodesEnergy(node []*energy.NodeEnergy, timestamp uint64, updateView bool) {
	for _, v := range mv.vs {
		v.UpdateNodesEnergy(node, timestamp, updateView)
	}
}

func (mv *MultiVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {
	for _, v := range mv.vs {
		v.SetEnergyAnalyser(ea)
	}
}

func (mv *MultiVisualizer) UpdateNodeStats(nodeStatsInfo *visualize.NodeStatsInfo) {
	for _, v := range mv.vs {
		v.UpdateNodeStats(nodeStatsInfo)
	}
}

func (mv *MultiVisualizer) UpdateTimeWindowStats(txRateStatsInfo *visualize.TimeWindowStatsInfo) {
	for _, v := range mv.vs {
		v.UpdateTimeWindowStats(txRateStatsInfo)
	}
}
