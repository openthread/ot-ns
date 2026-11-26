// Copyright (c) 2026, The OTNS Authors.
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
	"github.com/stretchr/testify/require"
)

const (
	testPingDstAddr1 = "fdde:ad00:beef:0:0:ff:fe00:1000"
	testPingDstAddr2 = "fdde:ad00:beef:0:0:ff:fe00:2000"
)

func TestCollectPingsFlushesTimedOutRequests(t *testing.T) {
	d := &Dispatcher{CurTime: 1_000_000}
	node := mockNode1()
	node.D = d

	node.onPingRequest(1_000_000, testPingDstAddr1, 16)
	node.onPingRequest(5_000_000, testPingDstAddr2, 32)
	require.Len(t, node.pendingPings, 2)

	// At exact timeout boundary (timestamp + maxPingDelayUs == CurTime), request is not yet expired.
	d.CurTime = 1_000_000 + maxPingDelayUs
	assert.Empty(t, node.CollectPings())
	require.Len(t, node.pendingPings, 2)

	// Advancing past timestamp + maxPingDelayUs with no onPingReply flushes only the expired request.
	d.CurTime = 1_000_000 + maxPingDelayUs + 1
	results := node.CollectPings()
	require.Len(t, results, 1)
	assert.Equal(t, &PingResult{
		Dst:      testPingDstAddr1,
		DataSize: 16,
		Delay:    maxPingDelayUs,
	}, results[0])
	require.Len(t, node.pendingPings, 1)
	assert.Equal(t, testPingDstAddr2, node.pendingPings[0].Dst)

	// Subsequent CollectPings before the second request expires returns empty while keeping it pending.
	assert.Empty(t, node.CollectPings())
	require.Len(t, node.pendingPings, 1)

	// Advancing past the second request's timeout flushes it as well.
	d.CurTime = 5_000_000 + maxPingDelayUs + 1
	results = node.CollectPings()
	require.Len(t, results, 1)
	assert.Equal(t, &PingResult{
		Dst:      testPingDstAddr2,
		DataSize: 32,
		Delay:    maxPingDelayUs,
	}, results[0])
	assert.Empty(t, node.pendingPings)
}

func TestCollectPingsPreservesRepliedAndTimedOutResults(t *testing.T) {
	d := &Dispatcher{CurTime: 1_000_000}
	node := mockNode1()
	node.D = d

	node.onPingRequest(1_000_000, testPingDstAddr1, 16)
	d.CurTime = 1_025_000
	node.onPingReply(1_000_000, testPingDstAddr1, 16, 64)

	// Next ping gets no reply because destination becomes unreachable.
	node.onPingRequest(2_000_000, testPingDstAddr1, 16)
	d.CurTime = 2_000_000 + maxPingDelayUs + 100

	results := node.CollectPings()
	require.Len(t, results, 2)
	assert.Equal(t, &PingResult{
		Dst:      testPingDstAddr1,
		DataSize: 16,
		Delay:    25_000,
	}, results[0])
	assert.Equal(t, &PingResult{
		Dst:      testPingDstAddr1,
		DataSize: 16,
		Delay:    maxPingDelayUs,
	}, results[1])
	assert.Empty(t, node.pendingPings)
}

func TestCollectPingsNilDispatcherSafe(t *testing.T) {
	node := mockNode1()
	node.onPingRequest(1_000_000, testPingDstAddr1, 16)
	assert.Empty(t, node.CollectPings())
	assert.Len(t, node.pendingPings, 1)
}
