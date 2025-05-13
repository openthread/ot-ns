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

	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

type alarmEvent struct {
	NodeId    NodeId
	Timestamp uint64 // timestamp of next alarm

	index int
}

type alarmQueue []*alarmEvent

func (aq alarmQueue) Len() int {
	return len(aq)
}

func (aq alarmQueue) Less(i, j int) bool {
	return aq[i].Timestamp < aq[j].Timestamp
}

func (aq alarmQueue) Swap(i, j int) {
	a, b := aq[i], aq[j]
	if a.index != i && b.index != j {
		logger.Panicf("wrong index")
	}

	aq[i], aq[j] = b, a             // swap the elements
	aq[i].index, aq[j].index = i, j // fix the indexes
}

func (aq *alarmQueue) Push(x interface{}) {
	e := x.(*alarmEvent)
	*aq = append(*aq, e)
	e.index = len(*aq) - 1
}

func (aq *alarmQueue) Pop() (elem interface{}) {
	eqlen := len(*aq)
	elem = (*aq)[eqlen-1]
	*aq = (*aq)[:eqlen-1]
	return
}

type alarmMgr struct {
	q      alarmQueue
	events map[NodeId]*alarmEvent
}

func newAlarmMgr() *alarmMgr {
	mgr := &alarmMgr{
		q:      alarmQueue{},
		events: map[NodeId]*alarmEvent{},
	}

	heap.Init(&mgr.q)
	return mgr
}

func (am *alarmMgr) AddNode(nodeid NodeId) {
	e := am.events[nodeid]
	logger.AssertNil(e)

	e = &alarmEvent{
		NodeId:    nodeid,
		Timestamp: Ever,
	}
	heap.Push(&am.q, e)
	am.events[nodeid] = e
}

func (am *alarmMgr) SetNotified(nodeid NodeId) {
	am.SetTimestamp(nodeid, Ever)
}

func (am *alarmMgr) SetTimestamp(nodeid int, timestamp uint64) {
	e := am.events[nodeid]
	logger.AssertNotNil(e)

	if e.Timestamp != timestamp {
		e.Timestamp = timestamp
		heap.Fix(&am.q, e.index)
	}
}

func (am *alarmMgr) GetTimestamp(nodeid int) uint64 {
	e := am.events[nodeid]
	logger.AssertNotNil(e)

	return e.Timestamp
}

func (am *alarmMgr) NextAlarm() *alarmEvent {
	if len(am.q) == 0 {
		return nil
	}

	return am.q[0]
}

func (am *alarmMgr) NextTimestamp() uint64 {
	if len(am.q) == 0 {
		return Ever
	}

	return am.q[0].Timestamp
}

func (am *alarmMgr) DeleteNode(id NodeId) {
	e := am.events[id]
	logger.AssertNotNil(e)
	heap.Remove(&am.q, e.index)
	delete(am.events, id)
}
