// Copyright (c) 2020-2022, The OTNS Authors.
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

const (
	// DefaultDispatcherSpeed is used in a speed parameter, to indicate default speed.
	DefaultDispatcherSpeed float64 = -1.0
)

type OtJoinerState int

const (
	OtJoinerStateIdle      OtJoinerState = 0
	OtJoinerStateDiscover  OtJoinerState = 1
	OtJoinerStateConnect   OtJoinerState = 2
	OtJoinerStateConnected OtJoinerState = 3
	OtJoinerStateEntrust   OtJoinerState = 4
	OtJoinerStateJoined    OtJoinerState = 5
)

// WatchLogLevel is the log-level set for an individual node when watching it. Values inherit OT logging.h values
// and extend these with OT-NS specific items.
type WatchLogLevel int

const (
	WatchTraceLevel   WatchLogLevel = 6
	WatchDebugLevel   WatchLogLevel = 5
	WatchInfoLevel    WatchLogLevel = 4
	WatchNoteLevel    WatchLogLevel = 3
	WatchWarnLevel    WatchLogLevel = 2
	WatchCritLevel    WatchLogLevel = 1
	WatchOffLevel     WatchLogLevel = 0
	WatchDefaultLevel               = WatchInfoLevel
)

func ParseWatchLogLevel(level string) WatchLogLevel {
	switch level {
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

func min(t1 uint64, t2 uint64) uint64 {
	if t1 <= t2 {
		return t1
	}
	return t2
}
