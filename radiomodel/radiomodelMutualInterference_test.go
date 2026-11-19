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

package radiomodel

import (
	"math"
	"testing"

	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/prng"
	"github.com/openthread/ot-ns/types"
)

type eventQueueStub struct {
	events []*event.Event
}

func (q *eventQueueStub) Add(evt *event.Event) {
	q.events = append(q.events, evt)
}

// TestRssiOnChannelDeterministic checks that the RSSI sampled on a channel with several concurrent transmitters
// does not depend on Go's randomized map iteration order. addSignalPowersDbm() drops a signal that is more than
// 15 dB below the running total, which makes the sum depend on the order in which transmitters are added. With
// transmitters at -60, -76 and -76 dBm over a -95 dBm noise floor, adding the strong one first gives -60.00 dBm
// and adding it last gives -59.79 dBm.
func TestRssiOnChannelDeterministic(t *testing.T) {
	prng.Init(1)
	rm := NewRadioModel("MutualInterference").(*RadioModelMutualInterference)
	params := rm.GetParameters()
	params.NoiseFloorDbm = -95
	params.ShadowFadingSigmaDb = 0 // no fading, so the RSSI is set by Tx power and path loss only.
	params.TimeFadingSigmaMaxDb = 0

	const ch = types.ChannelId(11)
	// All nodes are co-located: zero path loss, so each transmitter's RSSI equals its Tx power.
	cfg := &RadioNodeConfig{X: 100, Y: 100, RadioRange: 100}
	rx := NewRadioNode(1, cfg)
	rm.AddNode(rx)

	q := &eventQueueStub{}
	for i, txPowerDbm := range []int8{-76, -76, -60} {
		tx := NewRadioNode(2+i, cfg)
		rm.AddNode(tx)
		rm.HandleEvent(tx, q, &event.Event{
			Type:      event.EventTypeRadioCommStart,
			Timestamp: 1000,
			NodeId:    tx.Id,
			RadioCommData: event.RadioCommEventData{
				Channel:  ch,
				PowerDbm: txPowerDbm,
				Duration: 1000,
			},
		})
	}
	if n := len(rm.activeTransmitters[ch]); n != 3 {
		t.Fatalf("active transmitters = %d, want 3", n)
	}

	want := rm.getRssiOnChannel(rx, ch)
	if math.Abs(want-(-60.0)) > 0.5 {
		t.Fatalf("RSSI on channel = %f dBm, want about -60 dBm", want)
	}
	// A single call has a fair chance of hitting either order, so repeat it to make a nondeterministic sum show.
	for i := 0; i < 1000; i++ {
		if got := rm.getRssiOnChannel(rx, ch); got != want {
			t.Fatalf("call %d: RSSI on channel = %.15f dBm, first call gave %.15f dBm", i, got, want)
		}
	}
}
