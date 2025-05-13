// Copyright (c) 2020-2023, The OTNS Authors.
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

package dispatcher

import (
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/types"
)

type rloc16Map map[uint16][]*Node

func (m rloc16Map) Remove(rloc16 uint16, node *Node) {
	logger.AssertTrue(rloc16 != types.InvalidRloc16)
	logger.AssertTrue(m.Contains(rloc16, node))
	m[rloc16] = m.removeFromList(m[rloc16], node)
	logger.AssertFalse(m.Contains(rloc16, node))
}

func (m rloc16Map) Add(rloc16 uint16, node *Node) {
	logger.AssertTrue(rloc16 != types.InvalidRloc16)
	logger.AssertFalse(m.Contains(rloc16, node))
	m[rloc16] = append(m[rloc16], node)
	logger.AssertTrue(m.Contains(rloc16, node))
}

func (m rloc16Map) Contains(rloc16 uint16, node *Node) bool {
	for _, n := range m[rloc16] {
		if n == node {
			return true
		}
	}
	return false
}

func (m rloc16Map) removeFromList(nodes []*Node, node *Node) []*Node {
	for i, n := range nodes {
		if n == node {
			return append(nodes[:i], nodes[i+1:]...)
		}
	}

	return nodes
}
