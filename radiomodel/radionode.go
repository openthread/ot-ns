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

	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	Id NodeId

	// TxPower contains the last Tx power used by the node.
	TxPower DbValue

	// RxSensitivity contains the Rx sensitivity in dBm of the node.
	RxSensitivity DbValue

	// RadioRange is the radio range as configured by the simulation for this node.
	RadioRange float64

	// RadioState is the current radio's state; RadioTx only when physically transmitting.
	RadioState    RadioStates
	RadioSubState RadioSubStates

	// RadioChannel is the current radio's channel (For Rx, Tx, or sampling).
	RadioChannel ChannelId

	// Node position in units/pixels.
	X, Y, Z float64

	// rssiSampleMax tracks the max RSSI detected during a channel sampling operation.
	rssiSampleMax DbValue

	stats RadioNodeStats
}

type RadioNodeConfig struct {
	X, Y, Z    int
	RadioRange int
}

type RadioNodeStats struct {
	NumBytesTx int
}

func NewRadioNode(nodeid NodeId, cfg *RadioNodeConfig) *RadioNode {
	rn := &RadioNode{
		Id:            nodeid,
		TxPower:       RssiInvalid,
		RxSensitivity: RssiInvalid,
		X:             float64(cfg.X),
		Y:             float64(cfg.Y),
		Z:             float64(cfg.Z),
		RadioRange:    float64(cfg.RadioRange),
		RadioChannel:  InvalidChannel,
		rssiSampleMax: RssiMinusInfinity,
		stats:         RadioNodeStats{},
	}
	return rn
}

func (rn *RadioNode) SetChannel(ch ChannelId) {
	logger.AssertTrue(ch >= MinChannelNumber && ch <= MaxChannelNumber)
	rn.RadioChannel = ch
}

func (rn *RadioNode) SetRadioState(state RadioStates, subState RadioSubStates) {
	rn.RadioState = state
	rn.RadioSubState = subState
}

func (rn *RadioNode) SetRxSensitivity(rxSens DbValue) {
	rn.RxSensitivity = rxSens
}

func (rn *RadioNode) SetNodePos(x, y, z int) {
	// simplified model: ignore pos changes during Rx.
	rn.X, rn.Y, rn.Z = float64(x), float64(y), float64(z)
}

// GetDistanceTo gets the distance to another RadioNode (in grid/pixel units).
func (rn *RadioNode) GetDistanceTo(other *RadioNode) (dist float64) {
	dx := other.X - rn.X
	dy := other.Y - rn.Y
	dz := other.Z - rn.Z
	dist = math.Sqrt(dx*dx + dy*dy + dz*dz)
	return
}
