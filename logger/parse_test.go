// Copyright (c) 2026, The OTNS Authors.
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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSyslogPrefix(t *testing.T) {
	// Typical RCP syslog prefix with relative path and pid.
	line1 := "./ot-rfsim/ot-versions/ot-cli[119560]: Running OPENTHREAD/thread-reference"
	assert.Equal(t, "./ot-rfsim/ot-versions/ot-cli[119560]: ", ParseSyslogPrefix(line1))

	// Bare executable name.
	line2 := "ot-cli[1]: Thread version: 5"
	assert.Equal(t, "ot-cli[1]: ", ParseSyslogPrefix(line2))

	// Normal OT log line — must not match.
	line3 := "00:00:00.000 [D] Platform------: Clear ShortAddr entries"
	assert.Equal(t, "", ParseSyslogPrefix(line3))

	// Empty string — must not match.
	assert.Equal(t, "", ParseSyslogPrefix(""))
}

func TestParseOtLogLine(t *testing.T) {
	// Standard OT single-character level markers.
	ok, lvl := ParseOtLogLine("00:00:12.894 [D] SubMac--------: RadioState: Receive -> CsmaBackoff")
	assert.True(t, ok)
	assert.Equal(t, DebugLevel, lvl)

	ok, lvl = ParseOtLogLine("00:00:32.298 [I] Mle-----------: Send Advertisement (ff02:0:0:0:0:0:0:1)")
	assert.True(t, ok)
	assert.Equal(t, InfoLevel, lvl)

	ok, lvl = ParseOtLogLine("00:00:00.035 [N] Mle-----------: Attach attempt 1, AnyPartition reattaching with Active Dataset")
	assert.True(t, ok)
	assert.Equal(t, NoteLevel, lvl)

	ok, lvl = ParseOtLogLine("00:00:01.000 [W] Mac-----------: Frame tx attempt failed")
	assert.True(t, ok)
	assert.Equal(t, WarnLevel, lvl)

	ok, lvl = ParseOtLogLine("00:00:01.000 [C] Core----------: Assertion failed")
	assert.True(t, ok)
	assert.Equal(t, ErrorLevel, lvl)

	// The '-' marker (e.g. OTNS status push lines) maps to the default level.
	ok, lvl = ParseOtLogLine("00:00:00.035 [-] Otns----------: transmit=11,d841,179,ffff")
	assert.True(t, ok)
	assert.Equal(t, DefaultLevel, lvl)

	// Posix host multi-character level markers.
	ok, lvl = ParseOtLogLine("[NOTE]-AGENT---: Thread version: 1.4.0")
	assert.True(t, ok)
	assert.Equal(t, NoteLevel, lvl)

	ok, lvl = ParseOtLogLine("[CRIT]-Core----: terminate called after an exception")
	assert.True(t, ok)
	assert.Equal(t, ErrorLevel, lvl)

	ok, lvl = ParseOtLogLine("[WARN]-Mac-----: This is just a mockup test message")
	assert.True(t, ok)
	assert.Equal(t, WarnLevel, lvl)

	ok, lvl = ParseOtLogLine("[INFO]-Cli-----: command done; mockup test message")
	assert.True(t, ok)
	assert.Equal(t, InfoLevel, lvl)

	ok, lvl = ParseOtLogLine("[DEBG]-Platform: state updated; mockup test message")
	assert.True(t, ok)
	assert.Equal(t, DebugLevel, lvl)

	ok, lvl = ParseOtLogLine("(OTNS)       [T] RadioState----: EnergyState=Tx_ SubState=FrameTx RadioState=Tx_ Ch=11 RadioTime=147876808 NextStTime=+2433")
	assert.True(t, ok)
	assert.Equal(t, TraceLevel, lvl)

	// Lines without a recognizable level marker.
	ok, lvl = ParseOtLogLine("not a log line")
	assert.False(t, ok)
	assert.Equal(t, OffLevel, lvl)

	ok, lvl = ParseOtLogLine("")
	assert.False(t, ok)
	assert.Equal(t, OffLevel, lvl)
}

func TestParseOtnsStatusPush(t *testing.T) {
	// Typical OTNS status push line.
	ok, status := ParseOtnsStatusPush("00:00:04.248 [-] Otns----------: transmit=11,d841,17,ffff")
	assert.True(t, ok)
	assert.Equal(t, "transmit=11,d841,17,ffff", status)

	// A single dash after 'Otns' is enough to match.
	ok, status = ParseOtnsStatusPush("00:00:02.233 [-] Otns-: role=2")
	assert.True(t, ok)
	assert.Equal(t, "role=2", status)

	// Matches even when a syslog prefix precedes the log line.
	ok, status = ParseOtnsStatusPush("./my/path/ot-cli[42]: 00:00:02.233 [-] Otns----------: extaddr=0123456789abcdef")
	assert.True(t, ok)
	assert.Equal(t, "extaddr=0123456789abcdef", status)

	// An empty status after the marker still counts as a match.
	ok, status = ParseOtnsStatusPush("00:00:02.233 [-] Otns----------: ")
	assert.True(t, ok)
	assert.Equal(t, "", status)

	// A different OT module must not match.
	ok, status = ParseOtnsStatusPush("00:00:00.000 [D] Platform------: Clear ShortAddr entries")
	assert.False(t, ok)
	assert.Equal(t, "", status)

	// 'Otns' without any dash separator must not match.
	ok, status = ParseOtnsStatusPush("00:00:02.233 [-] Otns: transmit=1")
	assert.False(t, ok)
	assert.Equal(t, "", status)

	// Missing the ': ' separator must not match.
	ok, status = ParseOtnsStatusPush("00:00:02.233 [-] Otns---------- message here")
	assert.False(t, ok)
	assert.Equal(t, "", status)

	// Non-log lines must not match.
	ok, status = ParseOtnsStatusPush("not a log line")
	assert.False(t, ok)
	assert.Equal(t, "", status)

	ok, status = ParseOtnsStatusPush("")
	assert.False(t, ok)
	assert.Equal(t, "", status)
}

func TestSetAlternativeOtLogMarker(t *testing.T) {
	// [NOTE] format of a Posix host process: not touched, since not an RCP output
	line1 := "[NOTE]-BBA-----: BackboneAgent: Backbone Router becomes Primary!"
	result := setAlternativeOtLogMarker(line1)
	assert.Equal(t, line1, result)

	// Standard [D] format
	line2 := "00:00:13.613 [D] SubMac--------: RadioState: Receive -> CsmaBackoff"
	result = setAlternativeOtLogMarker(line2)
	assert.Equal(t, "00:00:13.613 [D] SubMac--------| RadioState: Receive -> CsmaBackoff", result)

	// longer format (hypothetical)
	line3 := "00:00:00.001 [D] P-SpinelDrivTest-HELPER--: Set state callback: OK"
	result = setAlternativeOtLogMarker(line3)
	assert.Equal(t, "00:00:00.001 [D] P-SpinelDrivTest-HELPER--| Set state callback: OK", result)

	// Marker not found: unchanged
	line4 := "not a log line"
	result = setAlternativeOtLogMarker(line4)
	assert.Equal(t, line4, result)
}
