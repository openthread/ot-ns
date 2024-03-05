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

package types

import (
	"fmt"
	"math"
)

type NodeId = int
type ChannelId = uint8

const (
	InvalidNodeId         NodeId = 0
	BroadcastNodeId       NodeId = -1
	InitialDispatcherPort        = 9000

	// InvalidExtAddr defines the invalid extended address for nodes.
	InvalidExtAddr       uint64 = math.MaxUint64
	InvalidThreadVersion uint16 = 0
)

// Node types and roles
const (
	FED    = "fed"
	MED    = "med"
	SED    = "sed"
	SSED   = "ssed"
	ROUTER = "router"
	REED   = "reed"
	BR     = "br"
	MTD    = "mtd"
	FTD    = "ftd"
	WIFI   = "wifi" // Wi-Fi interferer node
)

func GetNodeName(id NodeId) string {
	spacing := "  "
	if id >= 100 {
		spacing = ""
	} else if id >= 10 {
		spacing = " "
	}
	return fmt.Sprintf("Node<%d>%s", id, spacing)
}

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
	AddrTypeAloc      AddrType = "aloc"
	AddrTypeLinkLocal AddrType = "linklocal"
	AddrTypeSlaac     AddrType = "slaac"
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
		return "INVALID"
	}
}

type OtJoinerState int

const (
	OtJoinerStateIdle      OtJoinerState = 0
	OtJoinerStateDiscover  OtJoinerState = 1
	OtJoinerStateConnect   OtJoinerState = 2
	OtJoinerStateConnected OtJoinerState = 3
	OtJoinerStateEntrust   OtJoinerState = 4
	OtJoinerStateJoined    OtJoinerState = 5
)

type RadioStates byte

const (
	RadioDisabled RadioStates = 0
	RadioSleep    RadioStates = 1
	RadioRx       RadioStates = 2
	RadioTx       RadioStates = 3
	RadioInvalid  RadioStates = 255
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
		return "INVALID"
	}
}

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
