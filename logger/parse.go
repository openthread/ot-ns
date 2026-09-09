// Copyright (c) 2022-2026, The OTNS Authors.
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
	"strings"
)

const (
	OffLevelString     = "off"
	NoneLevelString    = "none"
	DefaultLevelString = "default"
	AltOtLogMarker     = "| "
	StdOtLogMarker     = ": "
)

// Example Posix ot-cli OTNS status push: 00:00:02.233 [-] Otns----------: transmit=11,d841,121,ffff
var (
	logPattern               = regexp.MustCompile(`\[(-|C|W|N|I|D|T|CRIT|WARN|NOTE|INFO|DEBG)]`)
	otnsStatusPushLogPattern = regexp.MustCompile(`\[-] Otns-+: (.*)$`)
	syslogPrefixPattern      = regexp.MustCompile(`^(\S+\[\d+\]: )`)
)

// ParseLevelString parses a log level string entered in the CLI or as cmdline argument.
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
	case OffLevelString, NoneLevelString:
		return OffLevel, nil
	case DefaultLevelString, "def":
		return DefaultLevel, nil
	default:
		return DefaultLevel, fmt.Errorf("invalid log level string: %s", level)
	}
}

func parseOtLevelChar(level byte) Level {
	switch level {
	case 'T': // not natively emitted by OpenThread: may be emitted by OT-RFSIM and OTNS.
		return TraceLevel
	case 'D':
		return DebugLevel
	case 'I':
		return InfoLevel
	case 'N':
		return NoteLevel
	case 'W':
		return WarnLevel
	case 'C':
		return ErrorLevel
	default:
		return DefaultLevel
	}
}

// ParseOtLogLine attempts to parse 'line' as an OT-generated log line with timestamp/level/message.
// Returns true if successful and also returns the determined log level of the log line.
func ParseOtLogLine(line string) (bool, Level) {
	logIdx := logPattern.FindStringSubmatchIndex(line)
	if logIdx == nil {
		return false, OffLevel
	}
	return true, parseOtLevelChar(line[logIdx[2]])
}

// ParseOtnsStatusPush parses an OT Posix host log line for OTNS status push events, coming from
// the OTNS module, and extracts the status message, if present.
// Returns true and the extracted status if a match is found, else returns false and an empty string.
func ParseOtnsStatusPush(line string) (bool, string) {
	match := otnsStatusPushLogPattern.FindStringSubmatch(line)
	if len(match) < 2 {
		return false, ""
	}
	return true, match[1]
}

// ParseSyslogPrefix checks if 'line' starts with a syslog-style prefix of the form "exename[pid]: ".
// Returns the prefix string (including the trailing space) if found, else returns "".
func ParseSyslogPrefix(line string) string {
	m := syslogPrefixPattern.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	return m[1]
}

// GetLevelString returns the canonical string representation of a Level for use in the OTNS CLI.
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

// setAlternativeOtLogMarker replaces the standard OT log marker with the alternative marker
// in the log line of an RCP process. String index search is used because the time format can
// get longer when the node's time advances beyond 24h.
// Example line (the first ': ' string is index 31):
// 00:00:00.900 [I] Mle-----------: Send Parent Request to routers (ff02:0:0:0:0:0:0:2)
func setAlternativeOtLogMarker(levelAndMsg string) string {
	idx := strings.Index(levelAndMsg, StdOtLogMarker)
	if idx >= 31 {
		return levelAndMsg[:idx] + AltOtLogMarker + levelAndMsg[idx+len(StdOtLogMarker):]
	}
	return levelAndMsg
}
