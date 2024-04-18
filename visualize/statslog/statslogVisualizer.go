// Copyright (c) 2023-2024, The OTNS Authors.
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

package visualize_statslog

import (
	"fmt"
	"os"
	"time"

	"github.com/openthread/ot-ns/energy"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
	. "github.com/openthread/ot-ns/visualize"
	"github.com/openthread/ot-ns/visualize/grpc/pb"
)

type statslogVisualizer struct {
	logFile        *os.File
	logFileName    string
	isFileEnabled  bool
	changed        bool   // flag to track if some node stats changed
	timestampUs    uint64 // simulation current timestamp
	logTimestampUs uint64 // last log entry timestamp
	stats          nodeStats
	oldStats       nodeStats

	nodeRoles      map[NodeId]OtDeviceRole
	nodeModes      map[NodeId]NodeMode
	nodePartitions map[NodeId]uint32
	nodesFailed    map[NodeId]struct{}
}

type nodeStats struct {
	numNodes      int
	numLeaders    int
	numPartitions int
	numRouters    int
	numEndDevices int
	numDetached   int
	numDisabled   int
	numSleepy     int
	numFailed     int
}

// NewStatslogVisualizer creates a new Visualizer that writes a log of network stats to file.
func NewStatslogVisualizer(outputDir string, simulationId int) Visualizer {
	return &statslogVisualizer{
		logFileName:    getStatsLogFileName(outputDir, simulationId),
		isFileEnabled:  true,
		changed:        true,
		nodeRoles:      make(map[NodeId]OtDeviceRole, 64),
		nodeModes:      make(map[NodeId]NodeMode, 64),
		nodePartitions: make(map[NodeId]uint32, 64),
		nodesFailed:    make(map[NodeId]struct{}),
	}
}

func (sv *statslogVisualizer) SetNetworkInfo(NetworkInfo) {
}

func (sv *statslogVisualizer) OnExtAddrChange(NodeId, uint64) {
}

func (sv *statslogVisualizer) SetSpeed(float64) {
}

func (sv *statslogVisualizer) SetParent(NodeId, uint64) {
}

func (sv *statslogVisualizer) CountDown(time.Duration, string) {
}

func (sv *statslogVisualizer) ShowDemoLegend(int, int, string) {
}

func (sv *statslogVisualizer) AddRouterTable(NodeId, uint64) {
}

func (sv *statslogVisualizer) RemoveRouterTable(NodeId, uint64) {
}

func (sv *statslogVisualizer) AddChildTable(NodeId, uint64) {
}

func (sv *statslogVisualizer) RemoveChildTable(NodeId, uint64) {
}

func (sv *statslogVisualizer) DeleteNode(id NodeId) {
	sv.changed = true
	delete(sv.nodeRoles, id)
	delete(sv.nodeModes, id)
	delete(sv.nodePartitions, id)
	delete(sv.nodesFailed, id)
}

func (sv *statslogVisualizer) SetNodePos(NodeId, int, int, int) {
}

func (sv *statslogVisualizer) SetController(SimulationController) {
}

func (sv *statslogVisualizer) Init() {
	sv.createLogFile()
}

func (sv *statslogVisualizer) Run() {
	// no goroutine
}

func (sv *statslogVisualizer) Stop() {
	// add a final entry with final status
	sv.writeLogEntry(sv.timestampUs, sv.calcStats())
	sv.close()
	logger.Debugf("statslogVisualizer stopped and CSV log file closed.")
}

func (sv *statslogVisualizer) AddNode(nodeid NodeId, cfg *NodeConfig) {
	sv.changed = true
	sv.nodeRoles[nodeid] = OtDeviceRoleDisabled
	sv.nodeModes[nodeid] = NodeMode{
		RxOnWhenIdle:     !cfg.RxOffWhenIdle,
		FullThreadDevice: !cfg.IsMtd,
		FullNetworkData:  !cfg.IsMtd,
	}
}

func (sv *statslogVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *MsgVisualizeInfo) {
}

func (sv *statslogVisualizer) SetNodeRloc16(id NodeId, rloc16 uint16) {
}

func (sv *statslogVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {
	sv.changed = true
	sv.nodeRoles[nodeid] = role
}

func (sv *statslogVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {
	sv.changed = true
	sv.nodeModes[nodeid] = mode
}

func (sv *statslogVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {
	logger.AssertTrue(parid > 0, "Partition ID cannot be 0")
	sv.changed = true
	sv.nodePartitions[nodeid] = parid
}

func (sv *statslogVisualizer) AdvanceTime(ts uint64, speed float64) {
	if sv.changed && sv.checkLogEntryChange() {
		sv.writeLogEntry(sv.timestampUs, sv.stats)
		sv.logTimestampUs = sv.timestampUs
		sv.oldStats = sv.stats
	}
	sv.changed = false // this is kept to avoid sv.calcStats() call every time.
	sv.timestampUs = ts
}

func (sv *statslogVisualizer) OnNodeFail(nodeid NodeId) {
	sv.changed = true
	sv.nodesFailed[nodeid] = struct{}{}
}

func (sv *statslogVisualizer) OnNodeRecover(nodeid NodeId) {
	sv.changed = true
	delete(sv.nodesFailed, nodeid)
}

func (sv *statslogVisualizer) SetTitle(TitleInfo) {
}

func (sv *statslogVisualizer) UpdateNodesEnergy([]*pb.NodeEnergy, uint64, bool) {
}

func (sv *statslogVisualizer) SetEnergyAnalyser(*energy.EnergyAnalyser) {
}

func (sv *statslogVisualizer) createLogFile() {
	logger.AssertNil(sv.logFile)

	var err error
	_ = os.Remove(sv.logFileName)

	sv.logFile, err = os.OpenFile(sv.logFileName, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		logger.Errorf("creating new stats log file %s failed: %+v", sv.logFileName, err)
		sv.isFileEnabled = false
		return
	}
	sv.writeLogFileHeader()
	logger.Debugf("Stats log file '%s' created.", sv.logFileName)
}

func (sv *statslogVisualizer) writeLogFileHeader() {
	// RFC 4180 CSV file: no leading or trailing spaces in header field names
	header := "timeSec,nNodes,nPartitions,nLeaders,nRouters,nChildren,nDetached,nDisabled,nSleepy,nFailed"
	_ = sv.writeToLogFile(header)
}

func (sv *statslogVisualizer) calcStats() nodeStats {
	s := nodeStats{
		numNodes:      len(sv.nodeRoles),
		numLeaders:    countRole(&sv.nodeRoles, OtDeviceRoleLeader),
		numPartitions: countUniquePts(&sv.nodePartitions),
		numRouters:    countRole(&sv.nodeRoles, OtDeviceRoleRouter),
		numEndDevices: countRole(&sv.nodeRoles, OtDeviceRoleChild),
		numDetached:   countRole(&sv.nodeRoles, OtDeviceRoleDetached),
		numDisabled:   countRole(&sv.nodeRoles, OtDeviceRoleDisabled),
		numSleepy:     countSleepy(&sv.nodeModes),
		numFailed:     len(sv.nodesFailed),
	}
	return s
}

func (sv *statslogVisualizer) checkLogEntryChange() bool {
	sv.stats = sv.calcStats()
	return sv.stats != sv.oldStats
}

func (sv *statslogVisualizer) writeLogEntry(ts uint64, stats nodeStats) {
	timeSec := float64(ts) / 1e6
	entry := fmt.Sprintf("%12.6f, %3d,%3d,%3d,%3d,%3d,%3d,%3d,%3d,%3d", timeSec, stats.numNodes, stats.numPartitions,
		stats.numLeaders, stats.numRouters, stats.numEndDevices, stats.numDetached, stats.numDisabled,
		stats.numSleepy, stats.numFailed)
	_ = sv.writeToLogFile(entry)
	logger.Debugf("statslog entry added: %s", entry)
}

func (sv *statslogVisualizer) writeToLogFile(line string) error {
	if !sv.isFileEnabled {
		return nil
	}
	_, err := sv.logFile.WriteString(line + "\n")
	if err != nil {
		sv.close()
		sv.isFileEnabled = false
		logger.Errorf("couldn't write to node log file (%s), closing it", sv.logFileName)
	}
	return err
}

func (sv *statslogVisualizer) close() {
	if sv.logFile != nil {
		_ = sv.logFile.Close()
		sv.logFile = nil
		sv.isFileEnabled = false
	}
}

func getStatsLogFileName(outputDir string, simId int) string {
	return fmt.Sprintf("%s/%d_stats.csv", outputDir, simId)
}

func countRole(nodeRoles *map[NodeId]OtDeviceRole, role OtDeviceRole) int {
	c := 0
	for _, r := range *nodeRoles {
		if r == role {
			c++
		}
	}
	return c
}

func countUniquePts(nodePts *map[NodeId]uint32) int {
	pts := make(map[uint32]struct{})
	for _, part := range *nodePts {
		pts[part] = struct{}{}
	}
	return len(pts)
}

func countSleepy(nodeModes *map[NodeId]NodeMode) int {
	c := 0
	for _, m := range *nodeModes {
		if !m.RxOnWhenIdle {
			c++
		}
	}
	return c
}
