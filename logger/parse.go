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

package logger

import (
	"fmt"
	"regexp"
)

const (
	OffLevelString     = "off"
	NoneLevelString    = "none"
	DefaultLevelString = "default"
)

var (
	logPattern = regexp.MustCompile(`\[(-|C|W|N|I|D|CRIT|WARN|NOTE|INFO|DEBG)]`)
)

func ParseLevelString(level string) (Level, error) {
	switch level {
	case "micro":
		return MicroLevel, nil
	case "trace", "T":
		return TraceLevel, nil
	case "debug", "D":
		return DebugLevel, nil
	case "info", "I":
		return InfoLevel, nil
	case "note", "N":
		return NoteLevel, nil
	case "warn", "warning", "W":
		return WarnLevel, nil
	case "crit", "critical", "error", "err", "C", "E":
		return ErrorLevel, nil
	case "off", "none":
		return OffLevel, nil
	case "default", "def":
		return DefaultLevel, nil
	default:
		return DefaultLevel, fmt.Errorf("invalid log level string: %s", level)
	}
}

func parseOtLevelChar(level byte) Level {
	switch level {
	case 'T':
		return TraceLevel
	case 'D':
		return DebugLevel
	case 'I':
		return InfoLevel
	case 'N':
		return NoteLevel
	case 'W':
		return WarnLevel
	case 'C', 'E':
		return ErrorLevel
	default:
		return DefaultLevel
	}
}

// ParseOtLogLine attempts to parse line as an OT generated log line with timestamp/level/message.
// Returns true if successful and also returns the determined log level of the log line.
func ParseOtLogLine(line string) (bool, Level) {
	logIdx := logPattern.FindStringSubmatchIndex(line)
	if logIdx == nil {
		return false, 0
	}
	return true, parseOtLevelChar(line[logIdx[2]])
}

func GetLevelString(level Level) string {
	switch level {
	case MicroLevel:
		return "micro"
	case TraceLevel:
		return "trace"
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case NoteLevel:
		return "note"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "crit"
	case OffLevel:
		return "off"
	default:
		Panicf("Unknown Level: %d", level)
		return ""
	}
}
