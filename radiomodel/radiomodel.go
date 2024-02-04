// Copyright (c) 2022-2024, The OTNS Authors.
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

	. "github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/prng"
	. "github.com/openthread/ot-ns/types"
)

type DbValue = float64

const UndefinedDbValue DbValue = math.MaxFloat64

// IEEE 802.15.4-2015 related parameters for 2.4 GHz O-QPSK PHY
const (
	MinChannelNumber     ChannelId = 0  // below 11 are sub-Ghz channels for 802.15.4-2015
	MaxChannelNumber     ChannelId = 39 // above 26 are currently used as pseudo-BLE-adv-channels
	DefaultChannelNumber ChannelId = 11
	TimeUsPerBit                   = 4
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

	// CheckRadioReachable checks if the srcNode radio can reach the dstNode radio, now, with a >0 probability.
	CheckRadioReachable(srcNode *RadioNode, dstNode *RadioNode) bool

	// GetTxRssi calculates at what RSSI level a radio frame Tx would be received by
	// dstNode, according to the radio model, in the ideal case of no other transmitters/interferers.
	// It returns the expected RSSI value at dstNode, or RssiMinusInfinity if the RSSI value will
	// fall below the minimum Rx sensitivity of the dstNode.
	GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbValue

	// OnEventDispatch is called when the Dispatcher sends an Event to a particular dstNode. The method
	// implementation may e.g. apply interference to a frame in transit, prior to delivery of the
	// frame at a single receiving radio dstNode, or apply loss of the frame, or set additional info
	// in the event. Returns true if event can be dispatched, false if not (e.g. due to Rx radio not
	// able to detect the frame).
	OnEventDispatch(srcNode *RadioNode, dstNode *RadioNode, evt *Event) bool

	// OnNextEventTime is called when the Dispatcher moves the simulation time to a higher timestamp ts,
	// where new event(s) will be executed.
	OnNextEventTime(ts uint64)

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node is the RadioNode object equivalent to evt.NodeId. Newly generated events may be put back into
	// the EventQueue q for scheduled processing.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)

	// GetName gets the display name of this RadioModel.
	GetName() string

	// GetParameters gets the parameters of this RadioModel. These may be modified during operation.
	GetParameters() *RadioModelParams

	// OnParametersModified is called when one or more parameters (RadioModelParams) were modified.
	OnParametersModified()

	// init initializes the RadioModel.
	init()
}

// NewRadioModel creates a new RadioModel with given name, or nil if model not found.
func NewRadioModel(modelName string) RadioModel {
	var model RadioModel
	rndSeed := prng.NewRadioModelRandomSeed()

	switch modelName {
	case "Ideal", "I", "1":
		model = &RadioModelIdeal{name: "Ideal", params: newRadioModelParams()}
		p := model.GetParameters()
		p.IsDiscLimit = true
		p.RssiMinDbm = -60.0
		p.RssiMaxDbm = -60.0

	case "Ideal_Rssi", "IR", "2", "default":
		model = &RadioModelIdeal{
			name:   "Ideal_Rssi",
			params: newRadioModelParams(),
		}
		p := model.GetParameters()
		setIndoorModelParamsItu(p)
		p.IsDiscLimit = true
	case "MutualInterference", "MI", "M", "3":
		model = &RadioModelMutualInterference{
			name:       "MutualInterference",
			params:     newRadioModelParams(),
			prevParams: *newRadioModelParams(),
			fading:     newFadingModel(rndSeed),
		}
		setIndoorModelParams3gpp(model.GetParameters())
	case "MIDisc", "MID", "4":
		model = &RadioModelMutualInterference{
			name:   "MIDisc",
			params: newRadioModelParams(),
			fading: newFadingModel(rndSeed),
		}
		p := model.GetParameters()
		setIndoorModelParams3gpp(p)
		p.IsDiscLimit = true
	case "Outdoor", "5":
		model = &RadioModelMutualInterference{
			name:   "Outdoor",
			params: newRadioModelParams(),
			fading: newFadingModel(rndSeed),
		}
		setOutdoorModelParams(model.GetParameters())
	default:
		model = nil
	}
	if model != nil {
		model.init()
	}
	return model
}
