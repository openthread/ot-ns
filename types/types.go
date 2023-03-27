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

package types

import (
	"math"

	"github.com/simonlingoogle/go-simplelogger"
)

type NodeId = int
type ChannelId = int

const (
	MaxNodeId       NodeId = 0xffff
	InvalidNodeId   NodeId = 0
	BroadcastNodeId NodeId = -1
)

const (
	// InvalidExtAddr defines the invalid extended address for nodes.
	InvalidExtAddr uint64 = math.MaxUint64
)

type NodeMode struct {
	RxOnWhenIdle     bool
	FullThreadDevice bool
	FullNetworkData  bool
}

func DefaultNodeMode() NodeMode {
	return NodeMode{
		RxOnWhenIdle:     true,
		FullThreadDevice: true,
		FullNetworkData:  true,
	}
}

func ParseNodeMode(s string) (mode NodeMode) {
	for _, c := range s {
		switch c {
		case 'r':
			mode.RxOnWhenIdle = true
		case 'd':
			mode.FullThreadDevice = true
		case 'n':
			mode.FullNetworkData = true
		}
	}
	return
}

type AddrType string

const (
	AddrTypeAny       AddrType = "any"
	AddrTypeMleid     AddrType = "mleid"
	AddrTypeRloc      AddrType = "rloc"
	AddrTypeLinkLocal AddrType = "linklocal"
)

type OtDeviceRole int

const (
	OtDeviceRoleDisabled OtDeviceRole = 0 ///< The Thread stack is disabled.
	OtDeviceRoleDetached OtDeviceRole = 1 ///< Not currently participating in a Thread network/partition.
	OtDeviceRoleChild    OtDeviceRole = 2 ///< The Thread Child role.
	OtDeviceRoleRouter   OtDeviceRole = 3 ///< The Thread Router role.
	OtDeviceRoleLeader   OtDeviceRole = 4 ///< The Thread Leader role.
)

func (r OtDeviceRole) String() string {
	switch r {
	case OtDeviceRoleDisabled:
		return "disabled"
	case OtDeviceRoleDetached:
		return "detached"
	case OtDeviceRoleChild:
		return "child"
	case OtDeviceRoleRouter:
		return "router"
	case OtDeviceRoleLeader:
		return "leader"
	default:
		simplelogger.Panicf("invalid device role: %v", r)
		return "invalid"
	}
}

type RadioStates byte

const (
	RadioDisabled RadioStates = 0
	RadioSleep    RadioStates = 1
	RadioRx       RadioStates = 2
	RadioTx       RadioStates = 3
)

func (s RadioStates) String() string {
	switch s {
	case RadioDisabled:
		return "Off"
	case RadioSleep:
		return "Slp"
	case RadioRx:
		return "Rx_"
	case RadioTx:
		return "Tx_"
	default:
		simplelogger.Panicf("invalid RadioState: %v", s)
		return "invalid"
	}
}

type RadioSubStates byte

const (
	OT_RADIO_SUBSTATE_READY             RadioSubStates = 0
	OT_RADIO_SUBSTATE_IFS_WAIT          RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_CCA            RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_CCA_TO_TX      RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_FRAME_ONGOING  RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_TX_TO_RX       RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_TX_TO_AIFS     RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_AIFS_WAIT      RadioSubStates = iota
	OT_RADIO_SUBSTATE_TX_ACK_RX_ONGOING RadioSubStates = iota
	OT_RADIO_SUBSTATE_RX_FRAME_ONGOING  RadioSubStates = iota
	OT_RADIO_SUBSTATE_RX_AIFS_WAIT      RadioSubStates = iota
	OT_RADIO_SUBSTATE_RX_ACK_TX_ONGOING RadioSubStates = iota
	OT_RADIO_SUBSTATE_RX_TX_TO_RX       RadioSubStates = iota
	OT_RADIO_SUBSTATE_RX_ENERGY_SCAN    RadioSubStates = iota
	OT_RADIO_SUBSTATE_STARTUP           RadioSubStates = iota
	OT_RADIO_SUBSTATE_INVALID           RadioSubStates = iota
)

func (s RadioSubStates) String() string {
	switch s {
	case OT_RADIO_SUBSTATE_READY:
		return "Ready__"
	case OT_RADIO_SUBSTATE_IFS_WAIT:
		return "IFS____"
	case OT_RADIO_SUBSTATE_TX_CCA:
		return "CCA____"
	case OT_RADIO_SUBSTATE_TX_CCA_TO_TX:
		return "CCA2Tx_"
	case OT_RADIO_SUBSTATE_TX_FRAME_ONGOING:
		return "FrameTx"
	case OT_RADIO_SUBSTATE_TX_TX_TO_RX:
		return "Tx2Rx__"
	case OT_RADIO_SUBSTATE_TX_TX_TO_AIFS:
		return "Tx2AIFS"
	case OT_RADIO_SUBSTATE_TX_AIFS_WAIT:
		return "TxAIFS_"
	case OT_RADIO_SUBSTATE_TX_ACK_RX_ONGOING:
		return "AckRx__"
	case OT_RADIO_SUBSTATE_RX_FRAME_ONGOING:
		return "FrameRx"
	case OT_RADIO_SUBSTATE_RX_AIFS_WAIT:
		return "RxAIFS_"
	case OT_RADIO_SUBSTATE_RX_ACK_TX_ONGOING:
		return "AckTx__"
	case OT_RADIO_SUBSTATE_RX_TX_TO_RX:
		return "AckT2Rx"
	case OT_RADIO_SUBSTATE_RX_ENERGY_SCAN:
		return "EnrScan"
	case OT_RADIO_SUBSTATE_STARTUP:
		return "Startup"
	case OT_RADIO_SUBSTATE_INVALID:
		return "Invalid"
	default:
		simplelogger.Panicf("invalid RadioSubState: %v", s)
		return ""
	}
}
