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

func TestParseOtLogLineAndAdaptMarker(t *testing.T) {
	const marker = "| "

	// [NOTE] format
	line1 := "[NOTE]-BBA-----: BackboneAgent: Backbone Router becomes Primary!"
	ok, level, result := ParseOtLogLineAndAdaptMarker(line1, marker)
	assert.True(t, ok)
	assert.Equal(t, NoteLevel, level)
	assert.Equal(t, "[NOTE]-BBA-----| BackboneAgent: Backbone Router becomes Primary!", result)

	// Standard [D] format
	line2 := "00:00:13.613 [D] SubMac--------: RadioState: Receive -> CsmaBackoff"
	ok, level, result = ParseOtLogLineAndAdaptMarker(line2, marker)
	assert.True(t, ok)
	assert.Equal(t, DebugLevel, level)
	assert.Equal(t, "00:00:13.613 [D] SubMac--------| RadioState: Receive -> CsmaBackoff", result)

	// Non-compliant format (extra space in module name)
	line2b := "00:00:13.613 [D] SubMac -------: RadioState: Receive -> CsmaBackoff"
	ok, level, result = ParseOtLogLineAndAdaptMarker(line2b, marker)
	assert.True(t, ok)
	assert.Equal(t, DebugLevel, level)
	assert.Equal(t, line2b, result)

	// longer format
	line3 := "00:00:00.001 [D] P-SpinelDriv[INFO]-HELPER--: Set state callback: OK"
	ok, level, result = ParseOtLogLineAndAdaptMarker(line3, marker)
	assert.True(t, ok)
	assert.Equal(t, DebugLevel, level)
	assert.Equal(t, "00:00:00.001 [D] P-SpinelDriv[INFO]-HELPER--| Set state callback: OK", result)

	// Non-log line: returns false and empty string.
	line4 := "not a log line"
	ok, level, result = ParseOtLogLineAndAdaptMarker(line4, marker)
	assert.False(t, ok)
	assert.Equal(t, OffLevel, level)
	assert.Equal(t, line4, result)

	line5 := "[NOTE]-AGENT---: Running 0.3.0-thread-reference-20250612-327-g111e78d0-dirty"
	ok, level, result = ParseOtLogLineAndAdaptMarker(line5, marker)
	assert.True(t, ok)
	assert.Equal(t, NoteLevel, level)
	assert.Equal(t, "[NOTE]-AGENT---| Running 0.3.0-thread-reference-20250612-327-g111e78d0-dirty", result)
}
