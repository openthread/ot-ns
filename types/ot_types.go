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

// OpenThread-specific types and definitions.

package types

// OT_ERROR_* error codes from OpenThread that can be sent by OT-NS to the OT nodes.
// (See OpenThread error.h for details)
const (
	OT_ERROR_NONE   = 0
	OT_ERROR_ABORT  = 11
	OT_ERROR_FCS    = 17
	OT_TX_TYPE_INTF = 192 // special status used for interference signals Tx
)

const (
	InvalidRloc16   uint16 = 0xfffe
	BroadcastRloc16 uint16 = 0xffff
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

// IEEE 802.15.4-2015 and other PHY related parameters, includes 2.4 GHz O-QPSK PHY
// these assumptions are hardcoded into the OT node stack and reproduced here.
const (
	MinChannelNumber  ChannelId = 0  // below 11 are sub-Ghz channels for 802.15.4-2015
	MaxChannelNumber  ChannelId = 39 // above 26 are currently used as pseudo-BLE-adv-channels
	TimeUsPerBit                = 4
	PhyHeaderLenBytes           = 6
	MacFrameLenBytes            = 127
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
