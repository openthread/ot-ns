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

// Package visualize implements the web visualization.
package visualize

import (
	"fmt"
	"time"

	"github.com/openthread/ot-ns/dissectpkt/wpan"
	. "github.com/openthread/ot-ns/types"
)

// Visualizer defines the interface of a visualizer.
type Visualizer interface {
	// Run runs the visualizer.
	Run()
	// Stop stops the visualizer.
	Stop()
	// AddNode adds a new node.
	AddNode(nodeid NodeId, extaddr uint64, x int, y int, radioRange int, mode NodeMode)
	// SetNodeRloc16 sets the node RLOC16.
	SetNodeRloc16(nodeid NodeId, rloc16 uint16)
	// SetNodeRole sets the node role.
	SetNodeRole(nodeid NodeId, role OtDeviceRole)
	Send(srcid NodeId, dstid NodeId, mvinfo *MsgVisualizeInfo)
	// SetNodePartitionId sets the node partition ID.
	SetNodePartitionId(nodeid NodeId, parid uint32)
	// SetSpeed sets the simulating speed.
	SetSpeed(speed float64)
	// AdvanceTime sets time to new timestamp.
	AdvanceTime(ts uint64, speed float64)
	// OnNodeFail notifies of a failed node.
	OnNodeFail(nodeId NodeId)
	// OnNodeRecover notifies of a recovered node.
	OnNodeRecover(nodeId NodeId)
	// SetController sets the simulation controller.
	SetController(ctrl SimulationController)
	// SetNodePos
	SetNodePos(nodeid NodeId, x, y int)
	// DeleteNode deletes a node.
	DeleteNode(id NodeId)
	// AddRouterTable adds a new router to the router table of a node.
	AddRouterTable(id NodeId, extaddr uint64)
	// RemoveRouterTable removes a router from the router table of a node.
	RemoveRouterTable(id NodeId, extaddr uint64)
	// AddChildTable adds a child to the child table of a node.
	AddChildTable(id NodeId, extaddr uint64)
	// RemoveChildTable removes a child from the child table of a node.
	RemoveChildTable(id NodeId, extaddr uint64)
	// ShowDemoLegend shows the demo legend.
	ShowDemoLegend(x int, y int, title string)
	// CountDown creates a new countdown with specified text and duration.
	CountDown(duration time.Duration, text string)
	// SetParent sets the parent of a node.
	SetParent(id NodeId, extaddr uint64)
	// OnExtAddrChange notifies of Extended Address change of a node.
	OnExtAddrChange(id NodeId, extaddr uint64)
}

// MsgVisualizeInfo contains visualization information of a message.
type MsgVisualizeInfo struct {
	Channel         uint8             // Message channel
	FrameControl    wpan.FrameControl // WPAN Frame Control
	Seq             uint8             // WPAN Sequence
	DstAddrShort    uint16            // Destination Short Address
	DstAddrExtended uint64            // Destination Extended Address
}

// MsgVisualizeInfo returns a short string label for the message.
func (info *MsgVisualizeInfo) Label() string {
	frameType := info.FrameControl.FrameType()
	if frameType == wpan.FrameTypeAck {
		return fmt.Sprintf("ACK%03d", info.Seq)
	} else if info.FrameControl.SecurityEnabled() {
		return fmt.Sprintf("MAC%03d", info.Seq)
	} else {
		return fmt.Sprintf("MLE%03d", info.Seq)
	}
}
