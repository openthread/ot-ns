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

package event

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"net"
	"net/netip"

	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/types"
)

type EventType = uint8

const (
	// Event type IDs (external, shared between OT-NS and OT node)
	EventTypeAlarmFired         EventType = 0
	EventTypeRadioReceived      EventType = 1
	EventTypeUartWrite          EventType = 2
	EventTypeRadioSpinelWrite   EventType = 3
	EventTypePostCmd            EventType = 4
	EventTypeStatusPush         EventType = 5
	EventTypeRadioCommStart     EventType = 6
	EventTypeRadioTxDone        EventType = 7
	EventTypeRadioChannelSample EventType = 8
	EventTypeRadioState         EventType = 9
	EventTypeRadioRxDone        EventType = 10
	EventTypeExtAddr            EventType = 11
	EventTypeNodeInfo           EventType = 12
	EventTypeNodeDisconnected   EventType = 14
	EventTypeRadioLog           EventType = 15
	EventTypeRadioRfSimParamGet EventType = 16
	EventTypeRadioRfSimParamSet EventType = 17
	EventTypeRadioRfSimParamRsp EventType = 18
	EventTypeLogWrite           EventType = 19
	EventTypeUdpToHost          EventType = 20
	EventTypeIp6ToHost          EventType = 21
	EventTypeUdpFromHost        EventType = 22
	EventTypeIp6FromHost        EventType = 23
)

const (
	InvalidTimestamp uint64 = math.MaxUint64
)

// Event format used by OT nodes.
const eventMsgHeaderLen = 19 // from OT platform-simulation.h struct Event { }
type Event struct {
	Delay uint64
	Type  EventType
	MsgId uint64
	//DataLen uint16
	Data []byte

	// metadata kept locally for this Event.
	NodeId       types.NodeId
	Timestamp    uint64
	MustDispatch bool
	Conn         net.Conn

	// supplementary payload data stored in Event.Data, depends on the event type.
	RadioCommData  RadioCommEventData
	RadioStateData RadioStateEventData
	NodeInfoData   NodeInfoEventData
	RfSimParamData RfSimParamEventData
	MsgToHostData  MsgToHostEventData
}

// All ...EventData formats below only used by OT nodes supporting advanced
// RF simulation.
const radioCommEventDataHeaderLen = 11 // from OT-RFSIM platform, event-sim.h struct
type RadioCommEventData struct {
	Channel  uint8
	PowerDbm int8
	Error    uint8
	Duration uint64
}

const radioStateEventDataHeaderLen = 14 // from OT-RFSIM platform, event-sim.h struct
type RadioStateEventData struct {
	Channel     uint8
	PowerDbm    int8
	RxSensDbm   int8
	EnergyState types.RadioStates
	SubState    types.RadioSubStates
	State       types.RadioStates
	RadioTime   uint64
}

const nodeInfoEventDataHeaderLen = 4 // from OT-RFSIM platform, otSimSendNodeInfoEvent()
type NodeInfoEventData struct {
	NodeId types.NodeId
}

const rfSimParamEventDataHeaderLen = 5 // from OT-RFSIM platform
type RfSimParamEventData struct {
	Param types.RfSimParam
	Value int32
}

const msgToHostEventDataHeaderLen = 36 // from OT-RFSIM platform
type MsgToHostEventData struct {
	SrcPort       uint16
	DstPort       uint16
	SrcIp6Address netip.Addr
	DstIp6Address netip.Addr
}

/*
RadioMessagePsduOffset is the offset of mPsdu data in a received OpenThread RadioMessage,
from OT-RFSIM platform, radio.h.

	struct RadioMessage
	{
		uint8_t mChannel;
		uint8_t mPsdu[OT_RADIO_FRAME_MAX_SIZE];
	} OT_TOOL_PACKED_END;
*/
const RadioMessagePsduOffset = 1

// Serialize serializes this Event into []byte to send to OpenThread node,
// including fields partially.
func (e *Event) Serialize() []byte {
	// Detect composite event types for which struct data is serialized.
	var extraFields []byte
	switch e.Type {
	case EventTypeRadioChannelSample:
		fallthrough
	case EventTypeRadioRxDone:
		fallthrough
	case EventTypeRadioTxDone:
		fallthrough
	case EventTypeRadioCommStart:
		extraFields = []byte{e.RadioCommData.Channel, byte(e.RadioCommData.PowerDbm), e.RadioCommData.Error,
			0, 0, 0, 0, 0, 0, 0, 0}
		binary.LittleEndian.PutUint64(extraFields[3:], e.RadioCommData.Duration)
	case EventTypeRadioRfSimParamSet:
		fallthrough
	case EventTypeRadioRfSimParamGet:
		extraFields = []byte{byte(e.RfSimParamData.Param), 0, 0, 0, 0}
		binary.LittleEndian.PutUint32(extraFields[1:], uint32(e.RfSimParamData.Value))
	case EventTypeUdpFromHost,
		EventTypeIp6FromHost:
		extraFields = make([]byte, msgToHostEventDataHeaderLen)
		binary.LittleEndian.PutUint16(extraFields[0:2], e.MsgToHostData.SrcPort)
		binary.LittleEndian.PutUint16(extraFields[2:4], e.MsgToHostData.DstPort)
		copy(extraFields[4:20], e.MsgToHostData.SrcIp6Address.AsSlice())
		copy(extraFields[20:36], e.MsgToHostData.DstIp6Address.AsSlice())
	default:
		break
	}

	payload := append(extraFields, e.Data...)
	msg := make([]byte, eventMsgHeaderLen+len(payload))
	binary.LittleEndian.PutUint64(msg[:8], e.Delay) // e.Timestamp is not sent, only e.Delay.
	msg[8] = e.Type
	binary.LittleEndian.PutUint64(msg[9:17], e.MsgId)
	binary.LittleEndian.PutUint16(msg[17:19], uint16(len(payload)))
	n := copy(msg[eventMsgHeaderLen:], payload)
	logger.AssertTrue(n == len(payload))

	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into the Event object e.
// It returns the number of bytes used from `data` for the Deserialize operation, or 0 if the data buffer
// is incomplete i.e. does not contain one entire serialized Event.
func (e *Event) Deserialize(data []byte) int {
	n := len(data)
	if n < eventMsgHeaderLen {
		return 0
	}
	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.Type = data[8]
	e.MsgId = binary.LittleEndian.Uint64(data[9:17])
	datalen := binary.LittleEndian.Uint16(data[17:19])
	var payloadOffset uint16 = 0
	if datalen > uint16(n-eventMsgHeaderLen) {
		return 0
	}
	e.Data = data[eventMsgHeaderLen : eventMsgHeaderLen+datalen]

	// Detect composite event types
	switch e.Type {
	case EventTypeRadioChannelSample:
		e.RadioCommData = deserializeRadioCommData(e.Data)
		payloadOffset += radioCommEventDataHeaderLen
	case EventTypeRadioRxDone:
		fallthrough
	case EventTypeRadioCommStart:
		e.RadioCommData = deserializeRadioCommData(e.Data)
		payloadOffset += radioCommEventDataHeaderLen
		logger.AssertEqual(e.RadioCommData.Channel, e.Data[payloadOffset]) // channel is stored twice.
	case EventTypeRadioState:
		e.RadioStateData = deserializeRadioStateData(e.Data)
		payloadOffset += radioStateEventDataHeaderLen
	case EventTypeNodeInfo:
		e.NodeInfoData = deserializeNodeInfoData(e.Data)
		payloadOffset += nodeInfoEventDataHeaderLen
	case EventTypeRadioRfSimParamRsp:
		e.RfSimParamData = deserializeRfSimParamData(e.Data)
		payloadOffset += rfSimParamEventDataHeaderLen
	case EventTypeUdpToHost:
		fallthrough
	case EventTypeIp6ToHost:
		e.MsgToHostData = deserializeMsgToHostData(e.Data)
		payloadOffset += msgToHostEventDataHeaderLen
	default:
		break
	}

	data2 := make([]byte, datalen-payloadOffset)
	copy(data2, e.Data[payloadOffset:datalen])
	e.Data = data2

	// e.Timestamp is not in the event, so set to invalid initially.
	e.Timestamp = InvalidTimestamp

	return int(eventMsgHeaderLen + datalen)
}

func deserializeRadioCommData(data []byte) RadioCommEventData {
	logger.AssertTrue(len(data) >= radioCommEventDataHeaderLen)
	s := RadioCommEventData{
		Channel:  data[0],
		PowerDbm: int8(data[1]),
		Error:    data[2],
		Duration: binary.LittleEndian.Uint64(data[3:]),
	}
	return s
}

func deserializeRadioStateData(data []byte) RadioStateEventData {
	logger.AssertTrue(len(data) >= radioStateEventDataHeaderLen)
	s := RadioStateEventData{
		Channel:     data[0],
		PowerDbm:    int8(data[1]),
		RxSensDbm:   int8(data[2]),
		EnergyState: types.RadioStates(data[3]),
		SubState:    types.RadioSubStates(data[4]),
		State:       types.RadioStates(data[5]),
		RadioTime:   binary.LittleEndian.Uint64(data[6:14]),
	}
	return s
}

func deserializeNodeInfoData(data []byte) NodeInfoEventData {
	logger.AssertTrue(len(data) >= nodeInfoEventDataHeaderLen)
	s := NodeInfoEventData{
		NodeId: types.NodeId(binary.LittleEndian.Uint32(data[0:4])),
	}
	return s
}

func deserializeRfSimParamData(data []byte) RfSimParamEventData {
	logger.AssertTrue(len(data) >= rfSimParamEventDataHeaderLen)
	s := RfSimParamEventData{
		Param: types.RfSimParam(data[0]),
		Value: int32(binary.LittleEndian.Uint32(data[1:5])),
	}
	return s
}

func deserializeMsgToHostData(data []byte) MsgToHostEventData {
	logger.AssertTrue(len(data) >= msgToHostEventDataHeaderLen)
	ip6Addr := [16]byte{}
	srcIp6 := [16]byte{}
	copy(srcIp6[:], data[4:20])
	copy(ip6Addr[:], data[20:36])

	s := MsgToHostEventData{
		SrcPort:       binary.LittleEndian.Uint16(data[0:2]),
		DstPort:       binary.LittleEndian.Uint16(data[2:4]),
		SrcIp6Address: netip.AddrFrom16(srcIp6),
		DstIp6Address: netip.AddrFrom16(ip6Addr),
	}
	return s
}

// Copy creates a (struct) copy of the Event.
func (e *Event) Copy() Event {
	newEv := *e
	return newEv
}

func (e *Event) String() string {
	paylStr := ""
	if len(e.Data) > 0 {
		paylStr = fmt.Sprintf(",payl=%s", hex.EncodeToString(e.Data))
	}
	s := fmt.Sprintf("Ev{%2d,nid=%d,mid=%d,dly=%v%s}", e.Type, e.NodeId, e.MsgId, e.Delay, paylStr)
	return s
}

/*
func keepPrintableChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}
*/
