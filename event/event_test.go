// Copyright (c) 2022-2023, The OTNS Authors.
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

package event

import (
	"encoding/hex"
	"testing"

	"github.com/openthread/ot-ns/types"
	"github.com/stretchr/testify/assert"
)

func TestDeserializeAlarmEvent(t *testing.T) {
	data, _ := hex.DecodeString("12120000000000000021222300000000000000")
	var ev Event
	n := ev.Deserialize(data)
	assert.True(t, 4626 == ev.Delay)
	assert.Equal(t, EventTypeAlarmFired, ev.Type)
	assert.Equal(t, uint64(2302497), ev.MsgId)
	assert.Equal(t, len(data), n)
}

func TestSerializeAlarmEvent(t *testing.T) {
	ev := &Event{Delay: 53716, Type: EventTypeAlarmFired}
	data := ev.Serialize()
	assert.True(t, len(data) == 19)
	assert.True(t, data[0] == 0xd4)
	assert.True(t, data[1] == 0xd1)
}

func TestDeserializeRadioCommEvent(t *testing.T) {
	data, _ := hex.DecodeString("040302010000000006040000000000000011000cf6112a000000000000000c1020304050")
	var ev Event
	n := ev.Deserialize(data)
	assert.True(t, 16909060 == ev.Delay)
	assert.Equal(t, EventTypeRadioCommStart, ev.Type)
	assert.Equal(t, uint64(4), ev.MsgId)
	assert.True(t, 12 == ev.RadioCommData.Channel)
	assert.True(t, -10 == ev.RadioCommData.PowerDbm)
	assert.True(t, types.OT_ERROR_FCS == ev.RadioCommData.Error)
	assert.True(t, 42 == ev.RadioCommData.Duration)
	assert.Equal(t, []byte{12, 0x10, 0x20, 0x30, 0x40, 0x50}, ev.Data)
	assert.Equal(t, len(data), n)
}

func TestDeserializeRadioStateEvent(t *testing.T) {
	data, _ := hex.DecodeString("0403020100000000090a000000000000000e000d05ab030b0240e2010000000000")
	var ev Event
	n := ev.Deserialize(data)
	assert.Equal(t, uint64(16909060), ev.Delay)
	assert.Equal(t, EventTypeRadioState, ev.Type)
	assert.Equal(t, uint64(10), ev.MsgId)
	assert.Equal(t, uint8(13), ev.RadioStateData.Channel)
	assert.Equal(t, int8(5), ev.RadioStateData.PowerDbm)
	assert.Equal(t, int8(-85), ev.RadioStateData.RxSensDbm)
	assert.Equal(t, types.RadioTx, ev.RadioStateData.EnergyState)
	assert.Equal(t, types.RFSIM_RADIO_SUBSTATE_RX_ACK_TX_ONGOING, ev.RadioStateData.SubState)
	assert.Equal(t, types.RadioRx, ev.RadioStateData.State)
	assert.Equal(t, uint64(123456), ev.RadioStateData.RadioTime)
	assert.Equal(t, len(data), n)
}

func TestDeserializeMultiple(t *testing.T) {
	data1, _ := hex.DecodeString("0403020100000000090a000000000000000e000d059c030b0240e2010000000000")
	data2, _ := hex.DecodeString("040302010000000006040000000000000011000cf6112a000000000000000c1020304050")
	data3, _ := hex.DecodeString("aabbccddeeff1122341122334455667788")
	data := append(data1, data2...)
	data = append(data, data3...)

	var ev Event
	n1 := ev.Deserialize(data)
	assert.Equal(t, uint64(16909060), ev.Delay)
	assert.Equal(t, EventTypeRadioState, ev.Type)
	assert.Equal(t, uint64(10), ev.MsgId)
	assert.Equal(t, uint8(13), ev.RadioStateData.Channel)
	assert.Equal(t, int8(5), ev.RadioStateData.PowerDbm)
	assert.Equal(t, int8(-100), ev.RadioStateData.RxSensDbm)
	assert.Equal(t, types.RadioTx, ev.RadioStateData.EnergyState)
	assert.Equal(t, types.RFSIM_RADIO_SUBSTATE_RX_ACK_TX_ONGOING, ev.RadioStateData.SubState)
	assert.Equal(t, types.RadioRx, ev.RadioStateData.State)
	assert.Equal(t, uint64(123456), ev.RadioStateData.RadioTime)
	assert.Equal(t, len(data1), n1)

	n2 := ev.Deserialize(data[n1:])
	assert.Equal(t, EventTypeRadioCommStart, ev.Type)
	assert.Equal(t, len(data2), n2)

	n3 := ev.Deserialize(data[n1+n2:])
	assert.Equal(t, 0, n3)
}

func TestSerializeRadioCommStartEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("0403020100000000060c0d0e0f00000000100002b01140e20100000000000210203040")
	rxEvtData := RadioCommEventData{
		Channel:  2,
		Error:    types.OT_ERROR_FCS,
		PowerDbm: -80,
		Duration: 123456,
	}
	framePayload := []byte{2, 0x10, 0x20, 0x30, 0x40}
	ev := &Event{
		Delay:         16909060,
		Type:          EventTypeRadioCommStart,
		MsgId:         252579084,
		RadioCommData: rxEvtData,
		Data:          framePayload,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestSerializeRadioCommTxDoneEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("0403020100000000075500000000000000100002b00040e20100000000000210203040")
	evtData := RadioCommEventData{
		Channel:  2,
		Error:    types.OT_ERROR_NONE,
		PowerDbm: -80,
		Duration: 123456,
	}
	framePayload := []byte{2, 0x10, 0x20, 0x30, 0x40}
	ev := &Event{
		Delay:         16909060,
		Type:          EventTypeRadioTxDone,
		MsgId:         85,
		RadioCommData: evtData,
		Data:          framePayload,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestSerializeRadioRxDoneEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("04030201000000000affff0000000000000b000baf0040e2010000000000")
	evData := RadioCommEventData{
		Channel:  11,
		Error:    types.OT_ERROR_NONE,
		PowerDbm: -81,
		Duration: 123456,
	}
	ev := &Event{
		Delay:         16909060,
		Type:          EventTypeRadioRxDone,
		MsgId:         65535,
		RadioCommData: evData,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestDeserializeNodeInfoEvent(t *testing.T) {
	data, _ := hex.DecodeString("00000000000000000c00000000000000fe040020000000")
	var ev Event
	n := ev.Deserialize(data)
	assert.True(t, 0 == ev.Delay)
	assert.Equal(t, EventTypeNodeInfo, ev.Type)
	assert.Equal(t, uint64(18302628885633695744), ev.MsgId)
	assert.Equal(t, 32, ev.NodeInfoData.NodeId)
	assert.Equal(t, len(data), n)

	data, _ = hex.DecodeString("00000000000000000cfe00000000000000040081800a00")
	n = ev.Deserialize(data)
	assert.True(t, 0 == ev.Delay)
	assert.Equal(t, EventTypeNodeInfo, ev.Type)
	assert.Equal(t, uint64(254), ev.MsgId)
	assert.Equal(t, 688257, ev.NodeInfoData.NodeId)
	assert.Equal(t, len(data), n)
}

func TestEventCopy(t *testing.T) {
	ev := &Event{
		Type:  EventTypeRadioRxDone,
		MsgId: 11234,
		Delay: 123,
		RadioCommData: RadioCommEventData{
			Channel: 42,
			Error:   types.OT_ERROR_FCS,
		},
	}
	evCopy := ev.Copy()
	assert.Equal(t, ev.Serialize(), evCopy.Serialize())

	// modify original
	ev.Delay += 1
	ev.RadioCommData.Channel = 11
	ev.RadioCommData.Error = types.OT_ERROR_NONE

	// check that copy is not modified
	assert.Equal(t, uint64(123), evCopy.Delay)
	assert.Equal(t, uint8(42), evCopy.RadioCommData.Channel)
	assert.Equal(t, uint8(types.OT_ERROR_FCS), evCopy.RadioCommData.Error)
	assert.Equal(t, uint64(11234), evCopy.MsgId)
}
