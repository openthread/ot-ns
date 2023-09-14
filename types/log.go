// Copyright (c) 2020-2023, The OTNS Authors.
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

package types

import (
	"fmt"
	"os"

	"github.com/openthread/ot-ns/cli/runcli"

	"github.com/simonlingoogle/go-simplelogger"
)

// WatchLogLevel is the log-level for watching what happens in the simulation as a whole, or to watch an
// individual node. Values inherit OT logging.h values and extend these with OT-NS specific items.
type WatchLogLevel int

const (
	WatchMicroLevel   WatchLogLevel = 7
	WatchTraceLevel   WatchLogLevel = 6
	WatchDebugLevel   WatchLogLevel = 5
	WatchInfoLevel    WatchLogLevel = 4
	WatchNoteLevel    WatchLogLevel = 3
	WatchWarnLevel    WatchLogLevel = 2
	WatchCritLevel    WatchLogLevel = 1
	WatchOffLevel     WatchLogLevel = 0
	WatchDefaultLevel               = WatchInfoLevel
)

const (
	WatchOffLevelString     = "off"
	WatchNoneLevelString    = "none"
	WatchDefaultLevelString = "default"
)

type LogEntry struct {
	NodeId  NodeId
	Level   WatchLogLevel
	Msg     string
	IsWatch bool
}

var (
	isLogToTerminal = false
)

func init() {
	o, _ := os.Stdout.Stat()
	if (o.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		isLogToTerminal = true
	}
}

func GetWatchLogLevelString(level WatchLogLevel) string {
	switch level {
	case WatchMicroLevel:
		return "micro"
	case WatchTraceLevel:
		return "trace"
	case WatchDebugLevel:
		return "debug"
	case WatchInfoLevel:
		return "info"
	case WatchNoteLevel:
		return "note"
	case WatchWarnLevel:
		return "warn"
	case WatchCritLevel:
		return "crit"
	case WatchOffLevel:
		return "off"
	default:
		simplelogger.Panicf("Unknown WatchLogLevel: %d", level)
		return ""
	}
}

func ParseWatchLogLevel(level string) WatchLogLevel {
	switch level {
	case "micro":
		return WatchMicroLevel
	case "trace", "T":
		return WatchTraceLevel
	case "debug", "D":
		return WatchDebugLevel
	case "info", "I":
		return WatchInfoLevel
	case "note", "N":
		return WatchNoteLevel
	case "warn", "warning", "W":
		return WatchWarnLevel
	case "crit", "critical", "error", "err", "C", "E":
		return WatchCritLevel
	case "off", "none":
		return WatchOffLevel
	case "default", "def":
		fallthrough
	default:
		return WatchDefaultLevel
	}
}

// GetSimpleloggerLevel converts WatchLogLevel to simplelogger.Level
func GetSimpleloggerLevel(lev WatchLogLevel) simplelogger.Level {
	switch lev {
	case WatchTraceLevel:
		fallthrough
	case WatchDebugLevel:
		return simplelogger.DebugLevel
	case WatchInfoLevel:
		return simplelogger.InfoLevel
	case WatchNoteLevel:
		return simplelogger.InfoLevel
	case WatchWarnLevel:
		return simplelogger.WarnLevel
	case WatchCritLevel:
		return simplelogger.ErrorLevel
	case WatchOffLevel:
		return simplelogger.PanicLevel
	default:
		simplelogger.Panicf("Unknown WatchLogLevel: %d", lev)
		return simplelogger.PanicLevel
	}
}

// PrintConsole prints a message for the user at the current console/CLI.
func PrintConsole(msg string) {
	if isLogToTerminal {
		fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	fmt.Fprint(os.Stdout, msg+"\n")
	if isLogToTerminal {
		runcli.RestorePrompt()
	}
}

// PrintLog prints the log msg at specified level using simplelogger.
func PrintLog(lev WatchLogLevel, msg string) {
	if isLogToTerminal {
		fmt.Fprint(os.Stdout, "\033[2K\r") // ANSI sequence to clear the CLI line
	}
	switch GetSimpleloggerLevel(lev) {
	case simplelogger.DebugLevel:
		simplelogger.Debugf("%s", msg)
	case simplelogger.InfoLevel:
		simplelogger.Infof("%s", msg)
	case simplelogger.WarnLevel:
		simplelogger.Warnf("%s", msg)
	case simplelogger.ErrorLevel:
		simplelogger.Errorf("%s", msg)
	case simplelogger.PanicLevel:
		simplelogger.Panicf("%s", msg)
	default:
		simplelogger.Panicf("%s", msg)
	}
	if isLogToTerminal {
		runcli.RestorePrompt()
	}
}
