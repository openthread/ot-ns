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

package radiomodel

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/prng"
	"github.com/openthread/ot-ns/types"
)

func TestApplyBerModelTraceLength(t *testing.T) {
	for _, tc := range []struct {
		name      string
		sirDb     DbValue
		psduBytes int
		wantError bool
	}{
		{"low SIR short frame", -11, 5, true},
		{"low SIR data frame", -11.837546, 87, true},
		{"low SIR maximum frame", -11, 127, true},
		{"computed SIR boundary", -10, 87, true},
		{"high SIR boundary", 6, 87, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data := make([]byte, event.RadioMessagePsduOffset+tc.psduBytes)
			data[0] = 11
			original := append([]byte(nil), data...)
			evt := event.Event{
				NodeId: 12,
				Data:   data,
			}
			// The intermediate-SIR path uses the package PRNG.
			prng.Init(1)
			logged, message := applyBerModel(tc.sirDb, 1, &evt)
			if logged != tc.wantError {
				t.Fatalf("logged = %t, want %t; trace = %q", logged, tc.wantError, message)
			}
			if tc.wantError {
				if evt.RadioCommData.Error != types.OT_ERROR_FCS {
					t.Fatalf("error = %d, want OT_ERROR_FCS", evt.RadioCommData.Error)
				}
				if want := fmt.Sprintf("PsduLen=%dB", tc.psduBytes); !strings.HasSuffix(message, want) {
					t.Errorf("trace = %q, want suffix %q", message, want)
				}
				if bytes.Equal(evt.Data[len(data)-2:], original[len(data)-2:]) {
					t.Error("FCS was not corrupted on failure")
				}
				if !bytes.Equal(evt.Data[:len(data)-2], original[:len(data)-2]) {
					t.Error("bytes before the FCS changed")
				}
			} else if message != "" || evt.RadioCommData.Error != types.OT_ERROR_NONE || !bytes.Equal(evt.Data, original) {
				t.Errorf("successful reception changed: error = %d, trace = %q", evt.RadioCommData.Error, message)
			}
			if !bytes.Equal(data, original) {
				t.Error("shared input data was modified")
			}
		})
	}
}

// TestBerModelReferenceValues checks the BER formula against the reference values used in the NS-3 LR-WPAN
// error model test, which implements the same IEEE 802.15.4-2006 E.4.1.8 formula for the 2.4 GHz O-QPSK PHY:
// https://gitlab.com/nsnam/ns-3-dev/-/blob/master/src/lr-wpan/test/lr-wpan-error-model-test.cc
func TestBerModelReferenceValues(t *testing.T) {
	refValues := []struct {
		sirDb     DbValue
		wantBer   float64
		tolerance float64
	}{
		{5, 7.38e-14, 0.01e-14},
		{2, 5.13e-7, 0.01e-7},
		{-1, 0.00114, 0.00001},
		{-4, 0.0391, 0.0001},
		{-7, 0.175, 0.001},
	}

	for _, tc := range refValues {
		t.Run(fmt.Sprintf("SIR %gdB", tc.sirDb), func(t *testing.T) {
			// As in NS-3, the BER is derived from the success rate of a single bit.
			ber := 1.0 - computePacketSuccessRate(tc.sirDb, 1)
			if math.Abs(ber-tc.wantBer) > tc.tolerance {
				t.Errorf("BER = %e, want %e +/- %e", ber, tc.wantBer, tc.tolerance)
			}
		})
	}
}
