// Copyright (c) 2020-2024, The OTNS Authors.
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
	"math"
	"time"

	. "github.com/openthread/ot-ns/types"
)

const (
	// DefaultDispatcherSpeed is used in a speed parameter, to indicate Dispatcher's current default speed.
	DefaultDispatcherSpeed float64 = -1.0
	Ever                   uint64  = math.MaxUint64 / 2
	MaxSimulateSpeed               = 1000000
	DefaultReadTimeout             = time.Second * 10
)

type TimeWindowStats struct {
	WinStartUs uint64
	WinWidthUs uint64
	PhyStats   map[NodeId]PhyStats

	statsWinStart map[NodeId]PhyStats // internal bookkeeping: stats at window start
}

func defaultTimeWindowStats() TimeWindowStats {
	return TimeWindowStats{
		WinStartUs:    0,
		WinWidthUs:    1e6,
		PhyStats:      make(map[NodeId]PhyStats),
		statsWinStart: make(map[NodeId]PhyStats),
	}
}

func min(t1 uint64, t2 uint64) uint64 {
	if t1 <= t2 {
		return t1
	}
	return t2
}
