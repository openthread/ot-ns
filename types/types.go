// Copyright (c) 2020, The OTNS Authors.
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

// Package types defines the common types used in OTNS.
package types

// NodeId represents a node ID which starts from 1.
type NodeId = int

const (
	InvalidNodeId   NodeId = 0  // invalid node ID.
	BroadcastNodeId NodeId = -1 // node ID for broadcasting messages.
)

// NodeMode defines node mode.
type NodeMode struct {
	RxOnWhenIdle       bool // radio RX on when idle
	SecureDataRequests bool // secure data requests
	FullThreadDevice   bool // full Thread device
	FullNetworkData    bool // full network data
}

// DefaultNodeMode returns a default NodeMode.
func DefaultNodeMode() NodeMode {
	return NodeMode{
		RxOnWhenIdle:       true,
		SecureDataRequests: true,
		FullThreadDevice:   true,
		FullNetworkData:    true,
	}
}

// OtDeviceRole represents the device role.
type OtDeviceRole int

const (
	OtDeviceRoleDisabled OtDeviceRole = 0 // The Thread stack is disabled.
	OtDeviceRoleDetached OtDeviceRole = 1 // Not currently participating in a Thread network/partition.
	OtDeviceRoleChild    OtDeviceRole = 2 // The Thread Child role.
	OtDeviceRoleRouter   OtDeviceRole = 3 // The Thread Router role.
	OtDeviceRoleLeader   OtDeviceRole = 4 // The Thread Leader role.
)

// OtJoinerState represents a joiner state.
type OtJoinerState int

const (
	OtJoinerStateIdle      OtJoinerState = 0 // Joiner is idle
	OtJoinerStateDiscover  OtJoinerState = 1 // Joiner is discovering
	OtJoinerStateConnect   OtJoinerState = 2 // Joiner is connecting
	OtJoinerStateConnected OtJoinerState = 3 // Joiner is connected
	OtJoinerStateEntrust   OtJoinerState = 4 // Joiner is entrusted
	OtJoinerStateJoined    OtJoinerState = 5 // Joiner is joined
)
