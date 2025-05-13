// Copyright (c) 2020-2022, The OTNS Authors.
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

	. "github.com/openthread/ot-ns/event"
	. "github.com/openthread/ot-ns/types"
	"github.com/stretchr/testify/assert"
)

func TestSendQueue_Add(t *testing.T) {
	q := newSendQueue()
	q.Add(&Event{Timestamp: 2, NodeId: 2})
	q.Add(&Event{Timestamp: 1, NodeId: 1})
	q.Add(&Event{Timestamp: 3, NodeId: 3})
}

func TestSendQueue_Len(t *testing.T) {
	q := newSendQueue()
	assert.Equal(t, 0, q.Len())
	q.Add(&Event{Timestamp: 2, NodeId: 2})
	assert.Equal(t, 1, q.Len())
	q.Add(&Event{Timestamp: 1, NodeId: 1})
	assert.Equal(t, 2, q.Len())
	q.Add(&Event{Timestamp: 3, NodeId: 3})
	assert.Equal(t, 3, q.Len())
}

func TestSendQueue_NextTimestamp(t *testing.T) {
	q := newSendQueue()
	assert.Equal(t, Ever, q.NextTimestamp())
	q.Add(&Event{Timestamp: 2, NodeId: 2, Data: []byte{0, 1, 2, 3, 4, 5}})
	assert.Equal(t, uint64(2), q.NextTimestamp())
	q.Add(&Event{Timestamp: 1, NodeId: 1})
	assert.Equal(t, uint64(1), q.NextTimestamp())
	q.Add(&Event{Timestamp: 3, NodeId: 3})
	assert.Equal(t, uint64(1), q.NextTimestamp())
}

func TestSendQueue_NextEvent(t *testing.T) {
	q := newSendQueue()
	q.Add(&Event{Timestamp: 2, NodeId: 2, Data: []byte{0, 1, 2, 3, 4, 5}})
	assert.Equal(t, uint64(2), q.NextEvent().Timestamp)
	assert.Equal(t, NodeId(2), q.NextEvent().NodeId)
	assert.Equal(t, []byte{0, 1, 2, 3, 4, 5}, q.NextEvent().Data)
	q.Add(&Event{Timestamp: 1, NodeId: 1})
	assert.Equal(t, uint64(1), q.NextEvent().Timestamp)
	assert.Equal(t, NodeId(1), q.NextEvent().NodeId)
	assert.Equal(t, []byte(nil), q.NextEvent().Data)
	q.Add(&Event{Timestamp: 3, NodeId: 3, Data: []byte{4, 5, 6}})
	assert.Equal(t, uint64(1), q.NextEvent().Timestamp)
	assert.Equal(t, NodeId(1), q.NextEvent().NodeId)
	assert.Equal(t, []byte(nil), q.NextEvent().Data)
}

func TestSendQueue_PopNext(t *testing.T) {
	q := newSendQueue()
	q.Add(&Event{Timestamp: 2, NodeId: 2})
	q.Add(&Event{Timestamp: 1, NodeId: 1})
	q.Add(&Event{Timestamp: 3, NodeId: 3})

	ev := q.PopNext()
	assert.True(t, ev.NodeId == 1 && ev.Timestamp == 1)
	ev = q.PopNext()
	assert.True(t, ev.NodeId == 2 && ev.Timestamp == 2)
	ev = q.PopNext()
	assert.True(t, ev.NodeId == 3 && ev.Timestamp == 3)
}
