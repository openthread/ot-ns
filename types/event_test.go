// Copyright (c) 2022, The OTNS Authors.
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

package types

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeserializeAlarmEvent(t *testing.T) {
	data, _ := hex.DecodeString("1212000000000000000000")
	var ev Event
	ev.Deserialize(data)
	assert.True(t, 4626 == ev.Delay)
	assert.Equal(t, EventTypeAlarmFired, ev.Type)
}

func TestSerializeAlarmEvent(t *testing.T) {
	ev := &Event{Delay: 53716, Type: EventTypeAlarmFired}
	data := ev.Serialize()
	assert.True(t, len(data) == 11)
	assert.True(t, data[0] == 0xd4)
	assert.True(t, data[1] == 0xd1)
}

func TestDeserializeRadioTxEvent(t *testing.T) {
	data, _ := hex.DecodeString("04030201000000001109000cf6a30c1020304050")
	var ev Event
	ev.Deserialize(data)
	assert.True(t, 16909060 == ev.Delay)
	assert.Equal(t, EventTypeRadioTx, ev.Type)
	assert.True(t, -93 == ev.TxData.CcaEdTresh)
	assert.True(t, 12 == ev.TxData.Channel)
	assert.True(t, -10 == ev.TxData.TxPower)
	assert.Equal(t, []byte{12, 0x10, 0x20, 0x30, 0x40, 0x50}, ev.Data)
}

func TestSerializeRadioRxEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("04030201000000001008000b11b0ff10203040")
	rxEvtData := RxEventData{
		Channel: 11,
		Error:   OT_ERROR_FCS,
		Rssi:    -80,
	}
	framePayload := []byte{0xff, 0x10, 0x20, 0x30, 0x40}
	ev := &Event{
		Delay:  16909060,
		Type:   EventTypeRadioRx,
		RxData: rxEvtData,
		Data:   framePayload,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestSerializeRadioTxDoneEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("04030201000000001202000b00")
	evData := TxDoneEventData{
		Channel: 11,
		Error:   OT_ERROR_NONE,
	}
	ev := &Event{
		Delay:      16909060,
		Type:       EventTypeRadioTxDone,
		TxDoneData: evData,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestEventCopy(t *testing.T) {
	ev := &Event{
		Type:  EventTypeRadioTxDone,
		Delay: 123,
		TxDoneData: TxDoneEventData{
			Channel: 42,
			Error:   OT_ERROR_CHANNEL_ACCESS_FAILURE,
		},
	}
	evCopy := ev.Copy()

	// modify original
	ev.Delay += 1
	ev.TxDoneData.Channel = 11
	ev.TxDoneData.Error = OT_ERROR_NONE

	assert.Equal(t, uint64(123), evCopy.Delay)
	assert.Equal(t, uint8(42), evCopy.TxDoneData.Channel)
	assert.Equal(t, uint8(OT_ERROR_CHANNEL_ACCESS_FAILURE), evCopy.TxDoneData.Error)
}
