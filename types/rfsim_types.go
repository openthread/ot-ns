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

package types

import "math"

// RSSI parameter encodings for communication with OT node
const (
	RssiInvalid       = 127
	RssiMax           = 126
	RssiMin           = -126
	RssiMinusInfinity = -127
)

// RfSim radio parameters, see ot-rfsim repo for values
type RfSimParam uint8

const (
	ParamRxSensitivity  RfSimParam = 0
	ParamCcaThreshold   RfSimParam = 1
	ParamCslAccuracy    RfSimParam = 2
	ParamCslUncertainty RfSimParam = 3
	ParamTxInterferer   RfSimParam = 4
	ParamUnknown        RfSimParam = 255
)

// Rfsim radio parameter values
type RfSimParamValue int32

const (
	RfSimValueInvalid RfSimParamValue = math.MaxInt32
)

var RfSimParamsList = []RfSimParam{ParamRxSensitivity, ParamCcaThreshold, ParamCslAccuracy, ParamCslUncertainty, ParamTxInterferer}
var RfSimParamNamesList = []string{"rxsens", "ccath", "cslacc", "cslunc", "txintf"}
var RfSimParamUnitsList = []string{"dBm", "dBm", "PPM", "10-us", "%"}

func ParseRfSimParam(parName string) RfSimParam {
	for i := 0; i < len(RfSimParamsList); i++ {
		if parName == RfSimParamNamesList[i] {
			return RfSimParamsList[i]
		}
	}
	return ParamUnknown
}
