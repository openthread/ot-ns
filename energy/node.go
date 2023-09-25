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

package energy

import (
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

type NodeEnergy struct {
	nodeId int
	radio  RadioStatus
}

func (node *NodeEnergy) ComputeRadioState(timestamp uint64) {
	delta := timestamp - node.radio.Timestamp
	switch node.radio.State {
	case RadioDisabled:
		node.radio.SpentDisabled += delta
	case RadioSleep:
		node.radio.SpentSleep += delta
	case RadioTx:
		node.radio.SpentTx += delta
	case RadioRx:
		node.radio.SpentRx += delta
	default:
		logger.Panicf("unknown radio state: %v", node.radio.State)
	}
	node.radio.Timestamp = timestamp
}

func (node *NodeEnergy) SetRadioState(state RadioStates, timestamp uint64) {
	//Mandatory: compute energy consumed by the radio first.
	node.ComputeRadioState(timestamp)
	node.radio.State = state
}

func newNode(nodeID int, timestamp uint64) *NodeEnergy {
	node := &NodeEnergy{
		nodeId: nodeID,
		radio: RadioStatus{
			State:         RadioDisabled,
			SpentDisabled: 0.0,
			SpentSleep:    0.0,
			SpentRx:       0.0,
			SpentTx:       0.0,
			Timestamp:     timestamp,
		},
	}
	return node
}
