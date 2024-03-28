// Copyright (c) 2024, The OTNS Authors.
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

package simulation

import . "github.com/openthread/ot-ns/types"

type KpiTimeUs struct {
	StartTimeUs uint64 `json:"start"`
	EndTimeUs   uint64 `json:"end"`
	PeriodUs    uint64 `json:"duration"`
}

type KpiTimeSec struct {
	StartTimeSec float64 `json:"start"`
	EndTimeSec   float64 `json:"end"`
	PeriodSec    float64 `json:"duration"`
}

type KpiChannel struct {
	TxTimeUs     uint64  `json:"tx_time_us"`
	TxPercentage float64 `json:"tx_percent"`
	NumFrames    uint64  `json:"tx_frames"`
	AvgFps       float64 `json:"tx_avg_fps"`
}

type KpiMac struct {
	NoAckPercentage map[NodeId]float64 `json:"noack_percent"`
}

type KpiCoapUri struct {
	Count     uint64  `json:"tx"`
	CountLost uint64  `json:"tx_lost"`
	LatencyMs float64 `json:"avg_latency_ms"`
}

type KpiCoap struct {
	Uri map[string]*KpiCoapUri `json:"uri"`
}

type Kpi struct {
	FileTime string                   `json:"created"`
	Status   string                   `json:"status"`
	TimeUs   KpiTimeUs                `json:"time_us"`
	TimeSec  KpiTimeSec               `json:"time_sec"`
	Channels map[ChannelId]KpiChannel `json:"channels"`
	Mac      KpiMac                   `json:"mac"`
	Counters map[NodeId]NodeCounters  `json:"counters"`
	Coap     KpiCoap                  `json:"coap"`
}
