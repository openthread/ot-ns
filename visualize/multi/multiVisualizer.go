// Copyright (c) 2022-2023, The OTNS Authors.
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
	"github.com/openthread/ot-ns/visualize/grpc/pb"
)

type multiVisualizer struct {
	vs []visualize.Visualizer
}

// NewMultiVisualizer creates a new Visualizer that multiplexes to multiple Visualizers.
func NewMultiVisualizer(vs ...visualize.Visualizer) visualize.Visualizer {
	return &multiVisualizer{vs: vs}
}

func (mv *multiVisualizer) SetNetworkInfo(networkInfo visualize.NetworkInfo) {
	for _, v := range mv.vs {
		v.SetNetworkInfo(networkInfo)
	}
}

func (mv *multiVisualizer) OnExtAddrChange(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.OnExtAddrChange(id, extaddr)
	}
}

func (mv *multiVisualizer) SetSpeed(speed float64) {
	for _, v := range mv.vs {
		v.SetSpeed(speed)
	}
}

func (mv *multiVisualizer) Init() {
	for _, v := range mv.vs {
		v.Init()
	}
}

func (mv *multiVisualizer) Run() {
	for i := 1; i < len(mv.vs); i++ {
		go mv.vs[i].Run()
	}
	mv.vs[0].Run()
}

func (mv *multiVisualizer) Stop() {
	for _, v := range mv.vs {
		v.Stop()
	}
}

func (mv *multiVisualizer) AddNode(nodeid NodeId, x int, y int, radioRange int) {
	for _, v := range mv.vs {
		v.AddNode(nodeid, x, y, radioRange)
	}
}

func (mv *multiVisualizer) SetNodeRloc16(nodeid NodeId, rloc16 uint16) {
	for _, v := range mv.vs {
		v.SetNodeRloc16(nodeid, rloc16)
	}
}

func (mv *multiVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {
	for _, v := range mv.vs {
		v.SetNodeRole(nodeid, role)
	}
}

func (mv *multiVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {
	for _, v := range mv.vs {
		v.SetNodeMode(nodeid, mode)
	}
}

func (mv *multiVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *visualize.MsgVisualizeInfo) {
	for _, v := range mv.vs {
		v.Send(srcid, dstid, mvinfo)
	}
}

func (mv *multiVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {
	for _, v := range mv.vs {
		v.SetNodePartitionId(nodeid, parid)
	}
}

func (mv *multiVisualizer) AdvanceTime(ts uint64, speed float64) {
	for _, v := range mv.vs {
		v.AdvanceTime(ts, speed)
	}
}

func (mv *multiVisualizer) OnNodeFail(nodeid NodeId) {
	for _, v := range mv.vs {
		v.OnNodeFail(nodeid)
	}
}

func (mv *multiVisualizer) OnNodeRecover(nodeid NodeId) {
	for _, v := range mv.vs {
		v.OnNodeRecover(nodeid)
	}
}

func (mv *multiVisualizer) SetController(ctrl visualize.SimulationController) {
	for _, v := range mv.vs {
		v.SetController(ctrl)
	}
}

func (mv *multiVisualizer) SetNodePos(nodeid NodeId, x, y int) {
	for _, v := range mv.vs {
		v.SetNodePos(nodeid, x, y)
	}
}

func (mv *multiVisualizer) DeleteNode(id NodeId) {
	for _, v := range mv.vs {
		v.DeleteNode(id)
	}
}

func (mv *multiVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.AddRouterTable(id, extaddr)
	}
}

func (mv *multiVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.RemoveRouterTable(id, extaddr)
	}
}

func (mv *multiVisualizer) AddChildTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.AddChildTable(id, extaddr)
	}
}

func (mv *multiVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.RemoveChildTable(id, extaddr)
	}
}

func (mv *multiVisualizer) ShowDemoLegend(x int, y int, title string) {
	for _, v := range mv.vs {
		v.ShowDemoLegend(x, y, title)
	}
}

func (mv *multiVisualizer) CountDown(duration time.Duration, text string) {
	for _, v := range mv.vs {
		v.CountDown(duration, text)
	}
}

func (mv *multiVisualizer) SetParent(id NodeId, extaddr uint64) {
	for _, v := range mv.vs {
		v.SetParent(id, extaddr)
	}
}

func (mv *multiVisualizer) SetTitle(titleInfo visualize.TitleInfo) {
	for _, v := range mv.vs {
		v.SetTitle(titleInfo)
	}
}

func (mv *multiVisualizer) UpdateNodesEnergy(node []*pb.NodeEnergy, timestamp uint64, updateView bool) {
	for _, v := range mv.vs {
		v.UpdateNodesEnergy(node, timestamp, updateView)
	}
}

func (mv *multiVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {
	for _, v := range mv.vs {
		v.SetEnergyAnalyser(ea)
	}
}
