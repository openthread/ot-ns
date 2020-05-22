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

package visualize_grpc

import (
	"github.com/openthread/ot-ns/threadconst"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type grpcNode struct {
	nodeid      NodeId
	extaddr     uint64
	x           int
	y           int
	radioRange  int
	mode        NodeMode
	rloc16      uint16
	role        visualize.OtDeviceRole
	partitionId uint32
	failed      bool
	parent      uint64
	routerTable map[uint64]struct{}
	childTable  map[uint64]struct{}
}

func newGprcNode(id NodeId, x int, y int, radioRange int, mode NodeMode) *grpcNode {
	gn := &grpcNode{
		nodeid:      id,
		extaddr:     InvalidExtAddr,
		x:           x,
		y:           y,
		radioRange:  radioRange,
		mode:        mode,
		rloc16:      threadconst.InvalidRloc16,
		role:        visualize.OtDeviceRoleDisabled,
		partitionId: 0,
		failed:      false,
		parent:      0,
		routerTable: map[uint64]struct{}{},
		childTable:  map[uint64]struct{}{},
	}
	return gn
}
