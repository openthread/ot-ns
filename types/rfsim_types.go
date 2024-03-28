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

// OT-RFSIM-platform-specific types and definitions.

package types

import "math"

// RSSI parameter encodings for communication with OT node
const (
	RssiInvalid       = 127
	RssiMax           = 126
	RssiMin           = -126
	RssiMinusInfinity = -127
)

type RadioSubStates byte

const (
	RFSIM_RADIO_SUBSTATE_READY             RadioSubStates = 0
	RFSIM_RADIO_SUBSTATE_IFS_WAIT          RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_CCA            RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_CCA_TO_TX      RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_FRAME_ONGOING  RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_TX_TO_RX       RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_TX_TO_AIFS     RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_AIFS_WAIT      RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_TX_ACK_RX_ONGOING RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_RX_FRAME_ONGOING  RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_RX_AIFS_WAIT      RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_RX_ACK_TX_ONGOING RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_RX_TX_TO_RX       RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_RX_ENERGY_SCAN    RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_STARTUP           RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_INVALID           RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_AWAIT_CCA         RadioSubStates = iota
	RFSIM_RADIO_SUBSTATE_CW_BACKOFF        RadioSubStates = iota
)

func (s RadioSubStates) String() string {
	switch s {
	case RFSIM_RADIO_SUBSTATE_READY:
		return "Ready__"
	case RFSIM_RADIO_SUBSTATE_IFS_WAIT:
		return "IFS____"
	case RFSIM_RADIO_SUBSTATE_TX_CCA:
		return "CCA____"
	case RFSIM_RADIO_SUBSTATE_TX_CCA_TO_TX:
		return "CCA2Tx_"
	case RFSIM_RADIO_SUBSTATE_TX_FRAME_ONGOING:
		return "FrameTx"
	case RFSIM_RADIO_SUBSTATE_TX_TX_TO_RX:
		return "Tx2Rx__"
	case RFSIM_RADIO_SUBSTATE_TX_TX_TO_AIFS:
		return "Tx2AIFS"
	case RFSIM_RADIO_SUBSTATE_TX_AIFS_WAIT:
		return "TxAIFS_"
	case RFSIM_RADIO_SUBSTATE_TX_ACK_RX_ONGOING:
		return "AckRx__"
	case RFSIM_RADIO_SUBSTATE_RX_FRAME_ONGOING:
		return "FrameRx"
	case RFSIM_RADIO_SUBSTATE_RX_AIFS_WAIT:
		return "RxAIFS_"
	case RFSIM_RADIO_SUBSTATE_RX_ACK_TX_ONGOING:
		return "AckTx__"
	case RFSIM_RADIO_SUBSTATE_RX_TX_TO_RX:
		return "AckT2Rx"
	case RFSIM_RADIO_SUBSTATE_RX_ENERGY_SCAN:
		return "EnrScan"
	case RFSIM_RADIO_SUBSTATE_STARTUP:
		return "Startup"
	case RFSIM_RADIO_SUBSTATE_INVALID:
		return "Invalid"
	case RFSIM_RADIO_SUBSTATE_AWAIT_CCA:
		return "WaitCCA"
	case RFSIM_RADIO_SUBSTATE_CW_BACKOFF:
		return "CwBackf"
	default:
		return "???????"
	}
}

// RfSim radio parameters, see ot-rfsim repo for values
type RfSimParam uint8

const (
	ParamRxSensitivity  RfSimParam = 0
	ParamCcaThreshold   RfSimParam = 1
	ParamCslAccuracy    RfSimParam = 2
	ParamCslUncertainty RfSimParam = 3
	ParamTxInterferer   RfSimParam = 4
	ParamClockDrift     RfSimParam = 5
	ParamUnknown        RfSimParam = 255
)

// Rfsim radio parameter values
type RfSimParamValue int32

const (
	RfSimValueInvalid RfSimParamValue = math.MaxInt32
)

var RfSimParamsList = []RfSimParam{ParamRxSensitivity, ParamCcaThreshold, ParamCslAccuracy, ParamCslUncertainty,
	ParamTxInterferer, ParamClockDrift}
var RfSimParamNamesList = []string{"rxsens", "ccath", "cslacc", "cslunc", "txintf", "clkdrift"}
var RfSimParamUnitsList = []string{"dBm", "dBm", "PPM", "10-us", "%", "PPM"}

func ParseRfSimParam(parName string) RfSimParam {
	for i := 0; i < len(RfSimParamsList); i++ {
		if parName == RfSimParamNamesList[i] {
			return RfSimParamsList[i]
		}
	}
	return ParamUnknown
}
