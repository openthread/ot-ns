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
	"strings"

	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type StatsType int

const (
	NodeStatsType StatsType = iota
	TxBytesStatsType
	ChanSampleCountStatsType
)

type statslogVisualizer struct {
	visualize.NopVisualizer

	statsType     StatsType
	simController visualize.SimulationController
	logFile       *os.File
	logFileName   string
	isFileEnabled bool
	changed       bool   // flag to track if some node stats changed
	timestampUs   uint64 // last node stats timestamp (= last log entry)
	stats         NodeStats
	oldStats      NodeStats
	lastIdMax     NodeId
}

// NewStatslogVisualizer creates a new Visualizer that writes a log of network stats to CSV file.
func NewStatslogVisualizer(outputDir string, simulationId int, csvDataType StatsType) visualize.Visualizer {
	return &statslogVisualizer{
		statsType:     csvDataType,
		logFileName:   getStatsLogFileName(csvDataType, outputDir, simulationId),
		isFileEnabled: true,
		changed:       true,
	}
}

func (sv *statslogVisualizer) Init() {
	sv.createLogFile()
}

func (sv *statslogVisualizer) Stop() {
	if sv.statsType == NodeStatsType {
		// add a final entry with final status
		sv.writeNodeStatsLogEntry(sv.timestampUs, sv.stats)
	}
	sv.close()
	logger.Debugf("statslogVisualizer stopped and CSV log file closed.")
}

func (sv *statslogVisualizer) UpdateNodeStats(info *visualize.NodeStatsInfo) {
	if sv.statsType == NodeStatsType {
		sv.oldStats = sv.stats
		sv.stats = info.Stats
		sv.timestampUs = info.TimeUs
		sv.writeNodeStatsLogEntry(sv.timestampUs, sv.stats)
	}
}

func (sv *statslogVisualizer) UpdateTimeWindowStats(info *visualize.TimeWindowStatsInfo) {
	ts := info.WinStartUs + info.WinWidthUs
	switch sv.statsType {
	case TxBytesStatsType:
		sv.writePhyStatsLogEntry(ts, info.PhyTxBytes)
	case ChanSampleCountStatsType:
		sv.writePhyStatsLogEntry(ts, info.ChanSampleCount)
	}
}

func (sv *statslogVisualizer) SetController(simController visualize.SimulationController) {
	sv.simController = simController
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
	var header string

	switch sv.statsType {
	case NodeStatsType:
		// RFC 4180 CSV file: no leading or trailing spaces in header field names
		header = "timeSec,nNodes,nPartitions,nLeaders,nRouters,nChildren,nDetached,nDisabled,nSleepy,nFailed"
		_ = sv.writeToLogFile(header)

	default:
		break
	}
}

func (sv *statslogVisualizer) writeNodeStatsLogEntry(ts uint64, stats NodeStats) {
	timeSec := float64(ts) / 1e6
	entry := fmt.Sprintf("%12.6f, %3d,%3d,%3d,%3d,%3d,%3d,%3d,%3d,%3d", timeSec, stats.NumNodes, stats.NumPartitions,
		stats.NumLeaders, stats.NumRouters, stats.NumEndDevices, stats.NumDetached, stats.NumDisabled,
		stats.NumSleepy, stats.NumFailed)
	_ = sv.writeToLogFile(entry)
	logger.Debugf("statslog entry added: %s", entry)
}

func (sv *statslogVisualizer) writePhyStatsLogEntry(ts uint64, stats map[NodeId]uint64) {
	var sb strings.Builder
	timeSec := float64(ts) / 1.0e6
	sb.WriteString(fmt.Sprintf("%12.6f, ", timeSec))

	idMin := 1
	_, idMax := calcMinMaxNodeId(stats)
	if sv.lastIdMax > idMax {
		idMax = sv.lastIdMax
	} else {
		sv.lastIdMax = idMax
	}
	for id := idMin; id <= idMax; id++ {
		if value, ok := stats[id]; ok {
			sb.WriteString(fmt.Sprintf("%5d, ", value))
		} else {
			sb.WriteString("    0, ")
		}
	}
	entry := sb.String()
	_ = sv.writeToLogFile(entry[:len(entry)-2])
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

func getStatsName(tp StatsType) string {
	switch tp {
	case NodeStatsType:
		return "stats"
	case TxBytesStatsType:
		return "txbytes"
	case ChanSampleCountStatsType:
		return "chansamples"
	default:
		return "INVALID"
	}
}

func getStatsLogFileName(tp StatsType, outputDir string, simId int) string {
	return fmt.Sprintf("%s/%d_%s.csv", outputDir, simId, getStatsName(tp))
}

func calcMinMaxNodeId(m map[NodeId]uint64) (NodeId, NodeId) {
	var idMin, idMax NodeId
	for id := range m {
		if idMin == 0 || id < idMin {
			idMin = id
		}
		if idMax == 0 || id > idMax {
			idMax = id
		}
	}
	return idMin, idMax
}
