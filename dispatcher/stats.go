// Copyright (c) 2024, The OTNS Authors.
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

package dispatcher

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

// updateNodeStats calculates fresh node statistics and sends it to the Visualizers.
func (d *Dispatcher) updateNodeStats() {
	s := d.calcStats()
	if s != d.oldStats {
		nodeStatsInfo := &visualize.NodeStatsInfo{
			TimeUs: d.CurTime,
			Stats:  s,
		}
		d.vis.UpdateNodeStats(nodeStatsInfo)
	}
}

func (d *Dispatcher) updateTimeWindowStats() {
	winEndTime := d.timeWinStats.WinStartUs + d.timeWinStats.WinWidthUs
	// conclude last time window, and move ahead 1 or more time windows
	if d.CurTime > winEndTime {
		statsEnd := d.radioModel.GetPhyStats()
		d.timeWinStats.PhyStats = calcPhyStatsDiff(d.timeWinStats.statsWinStart, statsEnd)
		d.visSendTimeWindowStats(&d.timeWinStats)

		d.timeWinStats.statsWinStart = statsEnd // reset for next round
		d.timeWinStats.WinStartUs += d.timeWinStats.WinWidthUs

		for d.CurTime > d.timeWinStats.WinStartUs+d.timeWinStats.WinWidthUs {
			d.timeWinStats.PhyStats = clearMapValuesPhyStats(d.timeWinStats.PhyStats)
			d.visSendTimeWindowStats(&d.timeWinStats) // send empty time window stats when no event happened.
			d.timeWinStats.WinStartUs += d.timeWinStats.WinWidthUs
		}
	}
}

func (d *Dispatcher) finalizeTimeWindowStats() {
	statsEnd := d.radioModel.GetPhyStats()
	d.timeWinStats.PhyStats = calcPhyStatsDiff(d.timeWinStats.statsWinStart, statsEnd)
	d.visSendTimeWindowStats(&d.timeWinStats)
}

func (d *Dispatcher) visSendTimeWindowStats(stats *TimeWindowStats) {
	statsInfo := &visualize.TimeWindowStatsInfo{
		WinStartUs:      stats.WinStartUs,
		WinWidthUs:      stats.WinWidthUs,
		NodePhyStats:    stats.PhyStats,
		PhyTxBytes:      make(map[NodeId]uint64),
		ChanSampleCount: make(map[NodeId]uint64),
	}
	for id, st := range stats.PhyStats {
		statsInfo.PhyTxBytes[id] = st.TxBytes
		statsInfo.ChanSampleCount[id] = st.ChanSampleCount
	}

	d.vis.UpdateTimeWindowStats(statsInfo)
}

func (d *Dispatcher) calcStats() NodeStats {
	s := NodeStats{
		NumNodes:      len(d.nodes),
		NumLeaders:    countRole(d.nodes, OtDeviceRoleLeader),
		NumPartitions: countUniquePts(d.nodes),
		NumRouters:    countRole(d.nodes, OtDeviceRoleRouter),
		NumEndDevices: countRole(d.nodes, OtDeviceRoleChild),
		NumDetached:   countRole(d.nodes, OtDeviceRoleDetached),
		NumDisabled:   countRole(d.nodes, OtDeviceRoleDisabled),
		NumSleepy:     countSleepy(d.nodes),
		NumFailed:     countFailed(d.nodes),
	}
	return s
}

func clearMapValuesPhyStats(m map[NodeId]PhyStats) map[NodeId]PhyStats {
	mNew := make(map[NodeId]PhyStats)
	for id := range m {
		mNew[id] = PhyStats{}
	}
	return mNew
}

func calcPhyStatsDiff(statsStart, statsEnd map[NodeId]PhyStats) map[NodeId]PhyStats {
	var st1 PhyStats
	var ok bool

	res := make(map[NodeId]PhyStats)
	for id, st2 := range statsEnd {
		if st1, ok = statsStart[id]; ok {
			res[id] = st2.Minus(st1)
		} else {
			res[id] = st2
		}
	}
	return res
}

func countRole(nodes map[NodeId]*Node, role OtDeviceRole) int {
	c := 0
	for _, n := range nodes {
		if n.Role == role {
			c++
		}
	}
	return c
}

func countUniquePts(nodes map[NodeId]*Node) int {
	pts := make(map[uint32]struct{})
	for _, n := range nodes {
		if n.PartitionId > 0 {
			pts[n.PartitionId] = struct{}{}
		}
	}
	return len(pts)
}

func countSleepy(nodeModes map[NodeId]*Node) int {
	c := 0
	for _, n := range nodeModes {
		if !n.Mode.RxOnWhenIdle {
			c++
		}
	}
	return c
}

func countFailed(nodes map[NodeId]*Node) int {
	c := 0
	for _, n := range nodes {
		if n.isFailed {
			c++
		}
	}
	return c
}
