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

package simulation

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/radiomodel"
	. "github.com/openthread/ot-ns/types"
)

type KpiManager struct {
	sim           *Simulation
	data          *Kpi
	startCounters NodeCountersStore
	curCounters   NodeCountersStore
	curRadioStats RadioStatsStore
	isRunning     bool
}

type NodeCountersStore map[NodeId]NodeCounters
type RadioStatsStore map[ChannelId]KpiChannel

// NewKpiManager creates a new KPI manager/bookkeeper for a particular simulation.
func NewKpiManager() *KpiManager {
	km := &KpiManager{}
	return km
}

// Init inits the KPI manager for the given simulation.
func (km *KpiManager) Init(sim *Simulation) {
	logger.AssertNil(km.sim)
	logger.AssertFalse(km.isRunning)
	km.sim = sim
	km.data = &Kpi{Status: "ok"}
	km.startCounters = NodeCountersStore{}
	km.curCounters = NodeCountersStore{}
}

func (km *KpiManager) Start() {
	logger.AssertNotNil(km.sim)
	km.startCounters = km.retrieveNodeCounters()
	km.data.TimeUs.StartTimeUs = km.sim.Dispatcher().CurTime
	rm := km.sim.Dispatcher().GetRadioModel()
	for ch := radiomodel.MinChannelNumber; ch <= radiomodel.MaxChannelNumber; ch++ {
		rm.ResetChannelStats(ch)
	}
	km.isRunning = true
	km.SaveFile(km.getDefaultSaveFileName())
}

func (km *KpiManager) Stop() {
	if km.isRunning {
		km.curCounters = km.retrieveNodeCounters()
		km.curRadioStats = km.retrieveRadioModelStats()
		km.isRunning = false
		km.calculateKpis()
		km.SaveDefaultFile()
	}
}

func (km *KpiManager) IsRunning() bool {
	return km.isRunning
}

func (km *KpiManager) SaveDefaultFile() {
	km.SaveFile(km.getDefaultSaveFileName())
}

func (km *KpiManager) SaveFile(fn string) {
	logger.AssertNotNil(km.sim)
	if km.isRunning {
		km.curCounters = km.retrieveNodeCounters()
		km.curRadioStats = km.retrieveRadioModelStats()
		km.calculateKpis()
	}

	km.data.FileTime = time.Now().Format(time.RFC3339)
	json, err := json.MarshalIndent(km.data, "", "    ")
	if err != nil {
		logger.Fatalf("Could not marshal KPI JSON data: %v", err)
		return
	}

	err = os.WriteFile(fn, json, 0644)
	if err != nil {
		logger.Errorf("Could not write  KPI JSON file %s: %v", fn, err)
		return
	}
}

func (km *KpiManager) stopNode(nodeid NodeId) {
	// deleted nodes during a KPI period won't be used anymore in final node-specific KPI calculations.
	delete(km.startCounters, nodeid)
	delete(km.curCounters, nodeid)
}

func (km *KpiManager) retrieveNodeCounters() NodeCountersStore {
	if km.sim.IsStopping() {
		return nil
	}
	nodes := km.sim.GetNodes()
	nodesMap := make(NodeCountersStore, len(nodes))
	for _, nid := range nodes {
		counters1 := km.sim.nodes[nid].GetCounters("mac", "mac.")
		counters2 := km.sim.nodes[nid].GetCounters("mle", "mle.")
		counters3 := km.sim.nodes[nid].GetCounters("ip", "ip.")
		nodesMap[nid] = mergeNodeCounters(counters1, counters2, counters3)
		km.sim.nodes[nid].DisplayPendingLogEntries()
		km.sim.nodes[nid].DisplayPendingLines()
	}
	return nodesMap
}

func (km *KpiManager) retrieveRadioModelStats() RadioStatsStore {
	ret := make(RadioStatsStore)
	curTime := km.sim.Dispatcher().CurTime
	passedTime := curTime - km.data.TimeUs.StartTimeUs

	if passedTime > 0 {
		for ch := radiomodel.MinChannelNumber; ch <= radiomodel.MaxChannelNumber; ch++ {
			stats := km.sim.Dispatcher().GetRadioModel().GetChannelStats(ch, curTime)
			if stats != nil {
				chanKpi := KpiChannel{
					TxTimeUs:     stats.TxTimeUs,
					TxPercentage: 100.0 * float64(stats.TxTimeUs) / float64(passedTime),
					NumFrames:    stats.NumFrames,
					AvgFps:       1.0e6 * float64(stats.NumFrames) / float64(passedTime),
				}
				ret[ch] = chanKpi
			}
		}
	}

	return ret
}

func getCountersDiff(curCtr NodeCounters, startCtr NodeCounters) NodeCounters {
	ret := NodeCounters{}
	for k, v := range curCtr {
		startVal := 0 // if node wasn't known at start, it was created during - use 0 for a counter's start value.
		if sv, ok := startCtr[k]; ok {
			startVal = sv
		}
		ret[k] = v - startVal
	}
	return ret
}

func (km *KpiManager) calculateKpis() {
	// time
	km.data.TimeUs.EndTimeUs = km.sim.Dispatcher().CurTime
	km.data.TimeUs.PeriodUs = km.data.TimeUs.EndTimeUs - km.data.TimeUs.StartTimeUs
	km.data.TimeSec.StartTimeSec = float64(km.data.TimeUs.StartTimeUs) / 1e6
	km.data.TimeSec.EndTimeSec = float64(km.data.TimeUs.EndTimeUs) / 1e6
	km.data.TimeSec.PeriodSec = float64(km.data.TimeUs.PeriodUs) / 1e6

	// channels
	km.data.Channels = km.curRadioStats

	// counters
	km.data.Mac.NoAckPercentage = make(map[NodeId]float64)
	km.data.Counters = make(map[NodeId]NodeCounters)
	if km.curCounters == nil {
		km.data.Status = "'counters' and 'mac' not included due to interrupted simulation"
	} else {
		for nid, ctr := range km.curCounters {
			counters := getCountersDiff(ctr, km.startCounters[nid])
			noAckPercent := 100.0 - 100.0*float64(counters["mac.TxAcked"])/float64(counters["mac.TxAckRequested"])
			if math.IsNaN(noAckPercent) {
				noAckPercent = 0.0
			}
			km.data.Mac.NoAckPercentage[nid] = noAckPercent
			km.data.Counters[nid] = counters
		}
	}
}

func (km *KpiManager) getDefaultSaveFileName() string {
	return fmt.Sprintf("%s/%d_kpi.json", km.sim.cfg.OutputDir, km.sim.cfg.Id)
}
