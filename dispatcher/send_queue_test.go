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

func TestSendQueue_Add(t *testing.T) {
	q := newSendQueue()
	q.Add(2, 2, nil)
	q.Add(1, 1, nil)
	q.Add(3, 3, nil)
}

func TestSendQueue_Len(t *testing.T) {
	q := newSendQueue()
	assert.Equal(t, 0, q.Len())
	q.Add(2, 2, nil)
	assert.Equal(t, 1, q.Len())
	q.Add(1, 1, nil)
	assert.Equal(t, 2, q.Len())
	q.Add(3, 3, nil)
	assert.Equal(t, 3, q.Len())
}

func TestSendQueue_NextTimestamp(t *testing.T) {
	q := newSendQueue()
	assert.Equal(t, Ever, q.NextTimestamp())
	q.Add(2, 2, nil)
	assert.Equal(t, uint64(2), q.NextTimestamp())
	q.Add(1, 1, nil)
	assert.Equal(t, uint64(1), q.NextTimestamp())
	q.Add(3, 3, nil)
	assert.Equal(t, uint64(1), q.NextTimestamp())
}

func TestSendQueue_PopNext(t *testing.T) {
	q := newSendQueue()
	q.Add(2, 2, nil)
	q.Add(1, 1, nil)
	q.Add(3, 3, nil)

	it := q.PopNext()
	assert.True(t, it.NodeId == 1 && it.Timestamp == 1)
	it = q.PopNext()
	assert.True(t, it.NodeId == 2 && it.Timestamp == 2)
	it = q.PopNext()
	assert.True(t, it.NodeId == 3 && it.Timestamp == 3)
}
