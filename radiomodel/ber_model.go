// Copyright (c) 2023, The OTNS Authors.
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
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"

	"github.com/openthread/ot-ns/logger"

	. "github.com/openthread/ot-ns/event"
	. "github.com/openthread/ot-ns/types"
)

// reference: IEEE 802.15.4-2006,E.4.1.8 Bit Error Rate (BER) calculations
// see also NS-3 LR-WPAN error model where this is used.
var (
	binomialCoeff = []float64{120, -560, 1820, -4368, 8008, -11440, 12870, -11440, 8008, -4368, 1820, -560, 120, -16, 1}
)

func applyBerModel(sirDb DbValue, srcNodeId NodeId, evt *Event) (bool, string) {
	pSuccess := 1.0
	var nbits int
	// if sirDb >= 6.0, then ratio SIR=~2, and pSuccess for any regular 15.4 frame is =~ 1.0 always.
	// Save time (?) by not doing the calculation then.
	if sirDb < 6.0 {
		pSuccess, nbits = computePacketSuccessRate(sirDb, evt.RadioCommData.Duration)
	}
	if pSuccess < 1.0 && rand.Float64() > pSuccess {
		evt.Data = interferePsduData(evt.Data)
		evt.RadioCommData.Error = OT_ERROR_FCS
		logMsg := fmt.Sprintf("applied OT_ERROR_FCS sirDb=%f src=%d dst=%d Psuc=%f FrLen=%dB",
			sirDb, srcNodeId, evt.NodeId, pSuccess, nbits/8)
		return true, logMsg
	}
	return false, ""
}

func computePacketSuccessRate(sirDb DbValue, frameDurationUs uint64) (float64, int) {
	nbits := float64(frameDurationUs / TimeUsPerBit)
	ber := 0.0
	snr := math.Pow(10, sirDb/10.0)
	for idx, coeff := range binomialCoeff {
		k := float64(idx + 2)
		ber += coeff * math.Exp(20.0*snr*(1.0/(k+1)-1.0))
	}

	ber = ber * 8.0 / 15.0 / 16.0

	ber = math.Min(ber, 1.0)
	psuc := math.Pow(1.0-ber, nbits)

	return psuc, int(nbits)
}

// interferePsduData simulates bit-error(s) on PSDU data
func interferePsduData(data []byte) []byte {
	logger.AssertTrue(len(data) >= 2)

	// modify MAC frame FCS, as a substitute for interfered frame.
	// a copy of the slice is made to avoid race conditions with other goroutines that may use the data frame.
	dl := len(data)
	fcs := binary.LittleEndian.Uint16(data[dl-2 : dl])
	intfData := append([]byte(nil), data...)
	binary.LittleEndian.PutUint16(intfData[dl-2:dl], fcs+42)
	return intfData
}
