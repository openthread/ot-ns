// Copyright (c) 2022-2024, The OTNS Authors.
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
	"net/netip"
	"testing"

	"github.com/openthread/ot-ns/types"
	"github.com/stretchr/testify/assert"
)

func TestDeserializeAlarmEvent(t *testing.T) {
	data, _ := hex.DecodeString("12120000000000000021222300000000000000")
	var ev Event
	n := ev.Deserialize(data)
	assert.True(t, ev.Delay == 4626)
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
	assert.True(t, ev.Delay == 16909060)
	assert.Equal(t, EventTypeRadioCommStart, ev.Type)
	assert.Equal(t, uint64(4), ev.MsgId)
	assert.True(t, ev.RadioCommData.Channel == 12)
	assert.True(t, ev.RadioCommData.PowerDbm == -10)
	assert.True(t, types.OT_ERROR_FCS == ev.RadioCommData.Error)
	assert.True(t, ev.RadioCommData.Duration == 42)
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
	assert.True(t, ev.Delay == 0)
	assert.Equal(t, EventTypeNodeInfo, ev.Type)
	assert.Equal(t, uint64(18302628885633695744), ev.MsgId)
	assert.Equal(t, 32, ev.NodeInfoData.NodeId)
	assert.Equal(t, len(data), n)

	data, _ = hex.DecodeString("00000000000000000cfe00000000000000040081800a00")
	n = ev.Deserialize(data)
	assert.True(t, ev.Delay == 0)
	assert.Equal(t, EventTypeNodeInfo, ev.Type)
	assert.Equal(t, uint64(254), ev.MsgId)
	assert.Equal(t, 688257, ev.NodeInfoData.NodeId)
	assert.Equal(t, len(data), n)
}

func TestSerializeRfSimGetEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("040302010000000010ffff00000000000005000200000000")
	evData := RfSimParamEventData{
		Param: types.ParamCslAccuracy,
		Value: 0, // value is encoded, but not used by OT-RFSIM platform.
	}
	ev := &Event{
		Delay:          16909060,
		Type:           EventTypeRadioRfSimParamGet,
		MsgId:          65535,
		RfSimParamData: evData,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestSerializeRfSimSetEvent(t *testing.T) {
	dataExpected, _ := hex.DecodeString("040302010000000011feff000000000000050001abffffff")
	evData := RfSimParamEventData{
		Param: types.ParamCcaThreshold,
		Value: -85,
	}
	ev := &Event{
		Delay:          16909060,
		Type:           EventTypeRadioRfSimParamSet,
		MsgId:          65534,
		RfSimParamData: evData,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
}

func TestDeserializeRfSimRspEvent(t *testing.T) {
	data, _ := hex.DecodeString("0403020100000000120400000000000000050002d2040000")
	var ev Event
	ev.Deserialize(data)
	assert.True(t, ev.Delay == 16909060)
	assert.Equal(t, EventTypeRadioRfSimParamRsp, ev.Type)
	assert.Equal(t, uint64(4), ev.MsgId)
	assert.Equal(t, types.ParamCslAccuracy, ev.RfSimParamData.Param)
	assert.Equal(t, int32(1234), ev.RfSimParamData.Value)
}

func TestDeserializeMsgToHostEvents(t *testing.T) {
	data, _ := hex.DecodeString("00000000000000001404000000000000002900efbe3316fe800000000000000000000000001234fe80000000000000000000000000beef0102030405")
	testIp6Addr, _ := hex.DecodeString("fe800000000000000000000000001234")
	testIp6Addr2, _ := hex.DecodeString("fe80000000000000000000000000beef")
	var ev Event
	ev.Deserialize(data)
	assert.True(t, ev.Delay == 0)
	assert.Equal(t, EventTypeUdpToHost, ev.Type)
	assert.Equal(t, uint64(4), ev.MsgId)
	assert.Equal(t, uint16(48879), ev.MsgToHostData.SrcPort)
	assert.Equal(t, uint16(5683), ev.MsgToHostData.DstPort)
	assert.Equal(t, testIp6Addr, ev.MsgToHostData.SrcIp6Address.AsSlice())
	assert.Equal(t, testIp6Addr2, ev.MsgToHostData.DstIp6Address.AsSlice())

	// try other event type with same payload structure
	data[8] = EventTypeIp6ToHost
	ev.Deserialize(data)
	assert.True(t, ev.Delay == 0)
	assert.Equal(t, EventTypeIp6ToHost, ev.Type)
	assert.Equal(t, uint64(4), ev.MsgId)
	assert.Equal(t, uint16(48879), ev.MsgToHostData.SrcPort)
	assert.Equal(t, uint16(5683), ev.MsgToHostData.DstPort)
	assert.Equal(t, testIp6Addr, ev.MsgToHostData.SrcIp6Address.AsSlice())
	assert.Equal(t, testIp6Addr2, ev.MsgToHostData.DstIp6Address.AsSlice())
}

func TestSerializeMsgToHostEvents(t *testing.T) {
	dataExpected, _ := hex.DecodeString("00000000000000001704000000000000002900efbe3316fe800000000000000000000000001234fe80abcd00000000000000000000abcd0102030405")
	evData := MsgToHostEventData{
		SrcPort:       48879,
		DstPort:       5683,
		SrcIp6Address: netip.MustParseAddr("fe80::1234"),
		DstIp6Address: netip.MustParseAddr("fe80:abcd::abcd"),
	}
	ev := &Event{
		Delay:         0,
		Type:          EventTypeIp6FromHost,
		MsgId:         4,
		Data:          []byte{1, 2, 3, 4, 5},
		MsgToHostData: evData,
	}
	data := ev.Serialize()
	assert.Equal(t, dataExpected, data)
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

func TestDeserializeEventFromLargerData(t *testing.T) {
	data, _ := hex.DecodeString("0403020100000000060c0d0e0f00000000100002b01140e201000000000002102030400809070605040302010000")
	ev := Event{}
	n := ev.Deserialize(data)
	assert.Equal(t, 35, n)
	assert.Equal(t, 35-eventMsgHeaderLen-radioCommEventDataHeaderLen, len(ev.Data))
}

func TestDeserializeEventFromTooLittleData(t *testing.T) {
	// header intact, but incomplete data (i.e. datalen too large)
	data, _ := hex.DecodeString("0403020100000000060c0d0e0f00000000100002b01140e201000000000002102030")
	ev := Event{}
	n := ev.Deserialize(data)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, len(ev.Data))

	data, _ = hex.DecodeString("0403020100000000060c0d0e0f00000000")
	ev = Event{}
	n = ev.Deserialize(data)
	assert.Equal(t, 0, n)
	assert.Equal(t, 0, len(ev.Data))
}
