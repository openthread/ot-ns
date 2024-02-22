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

package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/openthread/ot-ns/types"
)

// NodeLogger is a node-specific log object. Levels and output file can be set per individual node.
type NodeLogger struct {
	Id           NodeId
	fileLevel    Level
	displayLevel Level

	logFile       *os.File
	logFileName   string
	isFileEnabled bool
	entries       chan logEntry
	timestampUs   uint64
}

var (
	nodeLogs = make(map[NodeId]*NodeLogger, 10)
	mutex    = sync.Mutex{}
)

// GetNodeLogger gets the NodeLogger instance for the given ( simulation ID, node config ) and configures it.
func GetNodeLogger(outputDir string, simulationId int, cfg *NodeConfig) *NodeLogger {
	mutex.Lock()
	defer mutex.Unlock()

	var nl *NodeLogger
	nodeid := cfg.ID
	nl, ok := nodeLogs[nodeid]
	if !ok {
		nl = &NodeLogger{
			Id:            nodeid,
			fileLevel:     ErrorLevel,
			displayLevel:  ErrorLevel,
			entries:       make(chan logEntry, 1000),
			logFileName:   getLogFileName(outputDir, simulationId, nodeid),
			isFileEnabled: cfg.NodeLogFile,
		}
		nodeLogs[nodeid] = nl
		if nl.isFileEnabled {
			nl.createLogFile()
		}
	} else {
		// if logger already exists, adjust the configuration to latest provided and open file if needed.
		nl.isFileEnabled = cfg.NodeLogFile
		if nl.isFileEnabled && nl.logFile == nil {
			nl.openLogFile()
		}
	}
	return nl
}

func getLogFileName(outputPath string, simId int, nodeId NodeId) string {
	return fmt.Sprintf("%s/%d_%d.log", outputPath, simId, nodeId)
}

func (nl *NodeLogger) createLogFile() {
	var err error
	nl.logFile, err = os.OpenFile(nl.logFileName, os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		nl.Errorf("creating node log file %s failed: %+v", nl.logFileName, err)
		nl.isFileEnabled = false
		return
	}

	nl.writeLogFileHeader()
	nl.Debugf("Node log file '%s' created.", nl.logFileName)
}

func (nl *NodeLogger) openLogFile() {
	AssertTrue(nl.logFile == nil)

	var err error
	nl.logFile, err = os.OpenFile(nl.logFileName, os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		nl.Errorf("opening node log file %s failed: %+v", nl.logFileName, err)
		nl.isFileEnabled = false
		return
	}

	nl.writeLogFileHeader()
	nl.Debugf("Node log file '%s' opened.", nl.logFileName)
}

func (nl *NodeLogger) writeLogFileHeader() {
	header := fmt.Sprintf("#\n# OpenThread node log for %s Created %s\n", GetNodeName(nl.Id),
		time.Now().Format(time.RFC3339)) +
		"# SimTimeUs NodeTime     Lev LogModule       Message"
	_ = nl.writeToLogFile(header)
}

// NodeLogf logs a formatted log message for the specific nodeid; correct NodeLogger object will be auto-found.
func NodeLogf(nodeid NodeId, level Level, format string, args ...interface{}) {
	nl := nodeLogs[nodeid]
	if level > nl.fileLevel && level > nl.displayLevel {
		return
	}
	msg := getMessage(format, args)
	entry := logEntry{
		NodeId: nodeid,
		Level:  level,
		Msg:    msg,
	}
	select {
	case nl.entries <- entry:
		break
	default:
		nl.DisplayPendingLogEntries(nl.timestampUs)
		nl.entries <- entry
	}
}

func (nl *NodeLogger) SetFileLevel(level Level) {
	nl.fileLevel = level
}

func (nl *NodeLogger) SetDisplayLevel(level Level) {
	nl.displayLevel = level
}

func (nl *NodeLogger) Log(level Level, msg string) {
	NodeLogf(nl.Id, level, msg)
}

func (nl *NodeLogger) Logf(level Level, format string, args []interface{}) {
	NodeLogf(nl.Id, level, format, args)
}

func (nl *NodeLogger) Tracef(format string, args ...interface{}) {
	NodeLogf(nl.Id, TraceLevel, format, args...)
}

func (nl *NodeLogger) Debugf(format string, args ...interface{}) {
	NodeLogf(nl.Id, DebugLevel, format, args...)
}

func (nl *NodeLogger) Infof(format string, args ...interface{}) {
	NodeLogf(nl.Id, InfoLevel, format, args...)
}

func (nl *NodeLogger) Info(format string) {
	NodeLogf(nl.Id, InfoLevel, format)
}

func (nl *NodeLogger) Warnf(format string, args ...interface{}) {
	NodeLogf(nl.Id, WarnLevel, format, args...)
}

func (nl *NodeLogger) Warn(format string) {
	NodeLogf(nl.Id, WarnLevel, format)
}

func (nl *NodeLogger) Errorf(format string, args ...interface{}) {
	NodeLogf(nl.Id, ErrorLevel, format, args...)
}

func (nl *NodeLogger) Error(err error) {
	if err == nil {
		return
	}
	NodeLogf(nl.Id, ErrorLevel, "%v", err)
}

func (nl *NodeLogger) Panicf(format string, args ...interface{}) {
	NodeLogf(nl.Id, PanicLevel, format, args...)
}

func (nl *NodeLogger) writeToLogFile(line string) error {
	_, err := nl.logFile.WriteString(line + "\n")
	if err != nil {
		nl.Close()
		nl.isFileEnabled = false
		nl.Errorf("couldn't write to node log file (%s), closing it", nl.logFileName)
	}
	return err
}

// DisplayPendingLogEntries displays all pending log entries for the node, using given simulation time ts.
// This includes writing any pending entries to the node log file.
func (nl *NodeLogger) DisplayPendingLogEntries(ts uint64) {
	nl.timestampUs = ts
	tsStr := fmt.Sprintf("%11d ", ts)
	nodeStr := GetNodeName(nl.Id)
	for {
		select {
		case entry := <-nl.entries:
			isSaveEntry := nl.fileLevel >= entry.Level
			isDisplayEntry := nl.displayLevel >= entry.Level
			logStr := tsStr + entry.Msg
			// whatever is displayed (watch), will also be logged to file.
			if (isDisplayEntry || isSaveEntry) && nl.isFileEnabled {
				if logStr[len(logStr)-1:] == "\n" { // remove duplicate newline chars
					_ = nl.writeToLogFile(logStr[:len(logStr)-1])
				} else {
					_ = nl.writeToLogFile(logStr)
				}
			}
			if isDisplayEntry {
				logAlways(entry.Level, nodeStr+logStr)
			}
			break
		default:
			return
		}
	}
}

// IsFileEnabled returns true if logging to file is currently enabled, false if not.
func (nl *NodeLogger) IsFileEnabled() bool {
	return nl.isFileEnabled
}

// Close closes the node log file and also saves/displays any pending entries.
func (nl *NodeLogger) Close() {
	if nl.logFile != nil {
		nl.Debugf("Closing log file.")
		nl.DisplayPendingLogEntries(nl.timestampUs)
		_ = nl.logFile.Close()
		nl.logFile = nil
	}
}
