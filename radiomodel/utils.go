// Copyright (c) 2022-2023, The OTNS Authors.
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

import "math"

// paround is a custom parameter rounding function (2 digits)
func paround(param float64) float64 {
	return math.Round(param*100.0) / 100.0
}

// addSignalPowersDbm calculates signal power in dBm of two added, uncorrelated, signals with powers p1 and p2 (dBm).
func addSignalPowersDbm(p1 DbValue, p2 DbValue) DbValue {
	if p1 > p2+15.0 { // avoid costly calculation where possible
		return p1
	}
	if p2 > p1+15.0 {
		return p2
	}
	return 10.0 * math.Log10(math.Pow(10, p1/10.0)+math.Pow(10, p2/10.0))
}

// clipRssi clips the RSSI value (in dBm, as DbValue) to int8 range for return to OT nodes.
func clipRssi(rssi DbValue) int8 {
	if rssi > RssiMax {
		rssi = RssiMax
	} else if rssi < RssiMin {
		rssi = RssiMinusInfinity
	}
	return int8(math.Round(rssi))
}
