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

// OT-NS related and general-use types and definitions.

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

	InvalidChannel ChannelId = 0xff
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

type NodeStats struct {
	NumNodes      int
	NumLeaders    int
	NumPartitions int
	NumRouters    int
	NumEndDevices int
	NumDetached   int
	NumDisabled   int
	NumSleepy     int
	NumFailed     int
}
