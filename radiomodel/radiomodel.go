// Copyright (c) 2022, The OTNS Authors.
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
	"math"

	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// IEEE 802.15.4-2015 related parameters
type DbmValue = int8

const (
	MinChannelNumber     ChannelId = 0 // below 11 are sub-Ghz channels for 802.15.4-2015
	MaxChannelNumber     ChannelId = 26
	DefaultChannelNumber ChannelId = 11
)

// default radio & simulation parameters
const (
	receiveSensitivityDbm DbmValue = -100 // TODO for now MUST be manually kept equal to OT: SIM_RECEIVE_SENSITIVITY
	defaultTxPowerDbm     DbmValue = 0    // Default, RadioTxEvent msg will override it. OT: SIM_TX_POWER

	// Handtuned - for indoor model, how many meters r is RadioRange disc until Link
	// quality drops below 2 (10 dB margin).
	radioRangeIndoorDistInMeters = 26.70
)

// RSSI parameter encodings
const (
	RssiInvalid       DbmValue = 127
	RssiMax           DbmValue = 126
	RssiMin           DbmValue = -126
	RssiMinusInfinity DbmValue = -127
)

// EventQueue is the abstraction of the queue where the radio model sends its outgoing (new) events to.
type EventQueue interface {
	Add(*Event)
}

// RadioModel provides access to any type of radio model.
type RadioModel interface {

	// AddNode registers a (new) RadioNode to the model.
	AddNode(nodeid NodeId, radioNode *RadioNode)

	// DeleteNode removes a RadioNode from the model.
	DeleteNode(nodeid NodeId)

	// CheckRadioReachable checks if the srcNode radio can reach the dstNode radio, now.
	CheckRadioReachable(srcNode *RadioNode, dstNode *RadioNode) bool

	// GetTxRssi calculates at what RSSI level a radio frame Tx would be received by
	// dstNode, according to the radio model, in the ideal case of no other transmitters/interferers.
	// It returns the expected RSSI value at dstNode, or RssiMinusInfinity if the RSSI value will
	// fall below the minimum Rx sensitivity of the dstNode.
	GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbmValue

	// OnEventDispatch is called when the Dispatcher sends an Event to a particular dstNode. The method
	// implementation may e.g. apply interference to a frame in transit, prior to delivery of the
	// frame at a single receiving radio dstNode, or set additional info in the event.
	// Returns true if event can be dispatched, false if not.
	OnEventDispatch(srcNode *RadioNode, dstNode *RadioNode, evt *Event) bool

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node is the RadioNode object equivalent to evt.NodeId. Newly generated events may be put back into
	// the EventQueue q for scheduled processing.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)

	// GetName gets the display name of this RadioModel.
	GetName() string

	// init initializes the RadioModel.
	init()
}

// Create creates a new RadioModel with given name, or nil if model not found.
func Create(modelName string) RadioModel {
	var model RadioModel
	switch modelName {
	case "Ideal", "I", "1":
		model = &RadioModelIdeal{
			Name:      "Ideal",
			FixedRssi: -60,
		}
	case "Ideal_Rssi", "IR", "2", "default":
		model = &RadioModelIdeal{
			Name:            "Ideal_Rssi",
			UseVariableRssi: true,
		}
	case "MutualInterference", "MI", "M", "3":
		model = &RadioModelMutualInterference{
			MinSirDb: 1, // minimum Signal-to-Interference (SIR) (dB) required to detect signal
		}
	default:
		model = nil
	}
	if model != nil {
		model.init()
	}
	return model
}

// interferePsduData simulates the interference (garbling) of PSDU data based on a given SIR level (dB).
func interferePsduData(data []byte, sirDb float64) []byte {
	simplelogger.AssertTrue(len(data) >= 2)
	intfData := data
	if sirDb < 0 {
		// modify MAC frame FCS, as a substitute for interfered frame.
		intfData[len(data)-2]++
		intfData[len(data)-1]++
	}
	return intfData
}

// computeIndoorRssi computes the RSSI for a receiver at distance dist, using a simple indoor exponent=3.xx loss model.
func computeIndoorRssi(srcRadioRange float64, dist float64, txPower int8, rxSensitivity int8) int8 {
	pathloss := 0.0
	distMeters := dist * radioRangeIndoorDistInMeters / srcRadioRange
	if distMeters >= 0.072 {
		pathloss = 35.0*math.Log10(distMeters) + 40.0
	}
	rssi := float64(txPower) - pathloss
	rssiInt := int(math.Round(rssi))
	// constrain RSSI value to int8 and return it. If RSSI is below the receiver's rxSensitivity,
	// then return the RssiMinusInfinity value.
	if rssiInt >= int(RssiInvalid) {
		rssiInt = int(RssiMax)
	} else if rssiInt < int(RssiMin) || rssiInt < int(rxSensitivity) {
		rssiInt = int(RssiMinusInfinity)
	}
	return int8(rssiInt)
}
