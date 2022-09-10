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
	"container/heap"

	. "github.com/openthread/ot-ns/types"
)

type sendQueue struct {
	q []*Event
}

func (sq sendQueue) Len() int {
	return len(sq.q)
}

func (sq sendQueue) Less(i, j int) bool {
	return sq.q[i].Timestamp < sq.q[j].Timestamp
}

func (sq sendQueue) Swap(i, j int) {
	sq.q[i], sq.q[j] = sq.q[j], sq.q[i]
}

func (sq *sendQueue) Push(x interface{}) {
	sq.q = append(sq.q, x.(*Event))
}

func (sq *sendQueue) Pop() (elem interface{}) {
	eqlen := len(sq.q)
	elem = sq.q[eqlen-1]
	sq.q = sq.q[:eqlen-1]
	return
}

func (sq sendQueue) NextTimestamp() uint64 {
	if len(sq.q) > 0 {
		return sq.q[0].Timestamp
	} else {
		return Ever
	}
}

func (sq sendQueue) NextEvent() *Event {
	if len(sq.q) > 0 {
		return sq.q[0]
	} else {
		return nil
	}
}

func (sq *sendQueue) Add(timestamp uint64, id NodeId, data []byte) {
	heap.Push(sq, &Event{
		Timestamp: timestamp,
		NodeId:    id,
		Data:      data,
	})
}

func (sq *sendQueue) AddEvent(evt *Event) {
	heap.Push(sq, evt)
}

func (sq *sendQueue) PopNext() *Event {
	return heap.Pop(sq).(*Event)
}

func newSendQueue() *sendQueue {
	sq := &sendQueue{
		q: []*Event{},
	}
	heap.Init(sq)
	return sq
}
