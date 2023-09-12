// Copyright (c) 2020-2023, The OTNS Authors.
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
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"strings"
	"unicode"

	"github.com/simonlingoogle/go-simplelogger"
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
)

const (
	InvalidTimestamp uint64 = math.MaxUint64
)

// Event format used by OT nodes.
const EventMsgHeaderLen = 19 // from OT platform-simulation.h struct Event { }
type Event struct {
	Delay uint64
	Type  EventType
	MsgId uint64
	//DataLen uint16
	Data []byte

	// metadata kept locally for this Event.
	NodeId       NodeId
	Timestamp    uint64
	MustDispatch bool
	Conn         net.Conn

	// supplementary payload data stored in Event.Data, depends on the event type.
	RadioCommData  RadioCommEventData
	RadioStateData RadioStateEventData
	NodeInfoData   NodeInfoEventData
}

// All ...EventData formats below only used by OT nodes supporting advanced
// RF simulation.
const RadioCommEventDataHeaderLen = 11 // from OT-RFSIM platform, event-sim.h struct
type RadioCommEventData struct {
	Channel  uint8
	PowerDbm int8
	Error    uint8
	Duration uint64
}

const RadioStateEventDataHeaderLen = 13 // from OT-RFSIM platform, event-sim.h struct
type RadioStateEventData struct {
	Channel     uint8
	PowerDbm    int8
	EnergyState RadioStates
	SubState    RadioSubStates
	State       RadioStates
	RadioTime   uint64
}

const NodeInfoEventDataHeaderLen = 4 // from OT-RFSIM platform, otSimSendNodeInfoEvent()
type NodeInfoEventData struct {
	NodeId NodeId
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
	default:
		break
	}

	payload := append(extraFields, e.Data...)
	msg := make([]byte, EventMsgHeaderLen+len(payload))
	binary.LittleEndian.PutUint64(msg[:8], e.Delay) // e.Timestamp is not sent, only e.Delay.
	msg[8] = e.Type
	binary.LittleEndian.PutUint64(msg[9:17], e.MsgId)
	binary.LittleEndian.PutUint16(msg[17:19], uint16(len(payload)))
	n := copy(msg[EventMsgHeaderLen:], payload)
	simplelogger.AssertTrue(n == len(payload))

	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into the Event object e.
// It returns the number of bytes used from `data` for the Deserialize operation, or 0 if the data buffer
// is incomplete i.e. does not contain one entire serialized Event.
func (e *Event) Deserialize(data []byte) int {
	n := len(data)
	if n < EventMsgHeaderLen {
		return 0
	}
	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.Type = data[8]
	e.MsgId = binary.LittleEndian.Uint64(data[9:17])
	datalen := binary.LittleEndian.Uint16(data[17:19])
	var payloadOffset uint16 = 0
	if datalen > uint16(n-EventMsgHeaderLen) {
		return 0
	}
	e.Data = data[EventMsgHeaderLen : EventMsgHeaderLen+datalen]

	// Detect composite event types
	switch e.Type {
	case EventTypeRadioChannelSample:
		e.RadioCommData = deserializeRadioCommData(e.Data)
		payloadOffset += RadioCommEventDataHeaderLen
	case EventTypeRadioRxDone:
		fallthrough
	case EventTypeRadioCommStart:
		e.RadioCommData = deserializeRadioCommData(e.Data)
		payloadOffset += RadioCommEventDataHeaderLen
		simplelogger.AssertEqual(e.RadioCommData.Channel, e.Data[payloadOffset]) // channel is stored twice.
	case EventTypeRadioState:
		e.RadioStateData = deserializeRadioStateData(e.Data)
		payloadOffset += RadioStateEventDataHeaderLen
	case EventTypeNodeInfo:
		e.NodeInfoData = deserializeNodeInfoData(e.Data)
		payloadOffset += NodeInfoEventDataHeaderLen
	default:
		break
	}

	data2 := make([]byte, datalen-payloadOffset)
	copy(data2, e.Data[payloadOffset:payloadOffset+datalen])
	e.Data = data2

	// e.Timestamp is not in the event, so set to invalid initially.
	e.Timestamp = InvalidTimestamp

	return int(EventMsgHeaderLen + datalen)
}

func deserializeRadioCommData(data []byte) RadioCommEventData {
	simplelogger.AssertTrue(len(data) >= RadioCommEventDataHeaderLen)
	s := RadioCommEventData{
		Channel:  data[0],
		PowerDbm: int8(data[1]),
		Error:    data[2],
		Duration: binary.LittleEndian.Uint64(data[3:]),
	}
	return s
}

func deserializeRadioStateData(data []byte) RadioStateEventData {
	simplelogger.AssertTrue(len(data) >= RadioStateEventDataHeaderLen)
	s := RadioStateEventData{
		Channel:     data[0],
		PowerDbm:    int8(data[1]),
		EnergyState: RadioStates(data[2]),
		SubState:    RadioSubStates(data[3]),
		State:       RadioStates(data[4]),
		RadioTime:   binary.LittleEndian.Uint64(data[5:13]),
	}
	return s
}

func deserializeNodeInfoData(data []byte) NodeInfoEventData {
	simplelogger.AssertTrue(len(data) >= NodeInfoEventDataHeaderLen)
	s := NodeInfoEventData{
		NodeId: NodeId(binary.LittleEndian.Uint32(data[0:4])),
	}
	return s
}

// Copy creates a (struct) copy of the Event.
func (e Event) Copy() Event {
	newEv := e
	return newEv
}

func (e *Event) String() string {
	paylStr := ""
	if len(e.Data) > 0 {
		paylStr = fmt.Sprintf(",payl=%s", keepPrintableChars(string(e.Data)))
	}
	s := fmt.Sprintf("Ev{%2d,nid=%d,mid=%d,dly=%v%s}", e.Type, e.NodeId, e.MsgId, e.Delay, paylStr)
	return s
}

func keepPrintableChars(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, s)
}
