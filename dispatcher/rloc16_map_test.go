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

package dispatcher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	node1 = &Node{Id: 0x1, Rloc16: 0x1}
	node2 = &Node{Id: 0x2, Rloc16: 0x2}
	node3 = &Node{Id: 0x3, Rloc16: 0x3}
)

func TestRloc16Map_Add(t *testing.T) {
	rm := make(rloc16Map)
	rm.Add(node1.Rloc16, node1)
	rm.Add(node2.Rloc16, node2)
	rm.Add(node3.Rloc16, node3)
}

func TestRloc16Map_Contains(t *testing.T) {
	rm := make(rloc16Map)
	assert.False(t, rm.Contains(node1.Rloc16, node1))
	rm.Add(node1.Rloc16, node1)
	assert.True(t, rm.Contains(node1.Rloc16, node1))

	assert.False(t, rm.Contains(node2.Rloc16, node2))
	rm.Add(node2.Rloc16, node2)
	assert.True(t, rm.Contains(node2.Rloc16, node2))

	assert.False(t, rm.Contains(node3.Rloc16, node3))
	rm.Add(node3.Rloc16, node3)
	assert.True(t, rm.Contains(node3.Rloc16, node3))
}

func TestRloc16Map_Remove(t *testing.T) {
	rm := make(rloc16Map)
	rm.Add(node1.Rloc16, node1)
	rm.Add(node2.Rloc16, node2)
	rm.Add(node3.Rloc16, node3)

	assert.True(t, rm.Contains(node1.Rloc16, node1))
	assert.True(t, rm.Contains(node2.Rloc16, node2))
	assert.True(t, rm.Contains(node3.Rloc16, node3))

	rm.Remove(node1.Rloc16, node1)
	assert.False(t, rm.Contains(node1.Rloc16, node1))
	rm.Remove(node2.Rloc16, node2)
	assert.False(t, rm.Contains(node2.Rloc16, node2))
	rm.Remove(node3.Rloc16, node3)
	assert.False(t, rm.Contains(node3.Rloc16, node3))
}
