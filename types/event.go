// Copyright (c) 2020-2022, The OTNS Authors.
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
	EventTypeAlarmFired        EventType = 0
	EventTypeRadioReceived     EventType = 1
	EventTypeUartWrite         EventType = 2
	EventTypeRadioSpinelWrite  EventType = 3
	EventTypePostCmd           EventType = 4
	EventTypeStatusPush        EventType = 5
	EventTypeRadioCommRx       EventType = 6
	EventTypeRadioTxDone       EventType = 7
	EventTypeChannelSample     EventType = 8
	EventTypeChannelSampleDone EventType = 9
	EventTypeRadioCommTx       EventType = 10
	EventTypeRadioState        EventType = 11
)

const (
	InvalidTimestamp uint64 = math.MaxUint64
)

// Event format used by OT nodes.
const EventMsgHeaderLen = 11 // from OT platform-simulation.h struct Event { }
type Event struct {
	Delay uint64
	Type  EventType
	//DataLen uint16
	Data []byte

	// metadata kept locally for this Event.
	NodeId    NodeId
	SrcAddr   *net.UDPAddr
	Timestamp uint64

	// supplementary payload data stored in Event.Data, depends on the event type.
	TxData         TxEventData
	RxData         RxEventData
	TxDoneData     TxDoneEventData
	AlarmData      AlarmEventData
	ChanData       ChanEventData
	ChanDoneData   ChanDoneEventData
	RadioStateData RadioStateEventData
}

// All ...EventData formats below only used by OT nodes supporting advanced
// RF simulation.
const AlarmDataHeaderLen = 8 // from OT platform-simulation.h struct
type AlarmEventData struct {
	MsgId uint64
}

const TxEventDataHeaderLen = 2 // from OT platform-simulation.h struct
type TxEventData struct {
	Channel uint8
	TxPower int8
}

const RxEventDataHeaderLen = 3 // from OT platform-simulation.h struct
type RxEventData struct {
	Channel uint8
	Error   uint8
	Rssi    int8
}

const TxDoneEventDataHeaderLen = 2 // from OT platform-simulation.h struct
type TxDoneEventData struct {
	Channel uint8
	Error   uint8
}

const ChanEventDataHeaderLen = 1 //
type ChanEventData struct {
	Channel uint8
}

const ChanDoneEventDataHeaderLen = 2 //
type ChanDoneEventData struct {
	Channel uint8
	Rssi    int8
}

const RadioStateEventDataHeaderLen = 3 //
type RadioStateEventData struct {
	Channel uint8
	State   RadioStates
	TxPower int8
}

/* RadioMessagePsduOffset is the offset of Psdu data in a received OpenThread RadioMessage type.
type RadioMessage struct {
	Channel       uint8
	Psdu          byte[]
}
*/
const RadioMessagePsduOffset = 1

// Serialize serializes this Event into []byte to send to OpenThread node,
// including fields partially.
func (e *Event) Serialize() []byte {
	// Detect composite event types for which struct data is serialized.
	var extraFields []byte
	switch e.Type {
	case EventTypeRadioCommRx:
		extraFields = serializeRadioRxData(&e.RxData)
	case EventTypeRadioTxDone:
		extraFields = serializeRadioTxDoneData(&e.TxDoneData)
	case EventTypeChannelSampleDone:
		extraFields = serializeChanDoneData(&e.ChanDoneData)
	default:
		break
	}

	payload := append(extraFields, e.Data...)
	msg := make([]byte, EventMsgHeaderLen+len(payload))
	binary.LittleEndian.PutUint64(msg[:8], e.Delay) // e.Timestamp is not sent, only e.Delay.
	msg[8] = e.Type
	binary.LittleEndian.PutUint16(msg[9:11], uint16(len(payload)))
	n := copy(msg[EventMsgHeaderLen:], payload)
	simplelogger.AssertTrue(n == len(payload))

	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into Event object e.
func (e *Event) Deserialize(data []byte) {
	n := len(data)
	if n < EventMsgHeaderLen {
		simplelogger.Panicf("Event.Deserialize() message length too short: %d", n)
	}
	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.Type = data[8]
	datalen := binary.LittleEndian.Uint16(data[9:11])
	var payloadOffset uint16 = 0
	simplelogger.AssertTrue(datalen == uint16(n-EventMsgHeaderLen))
	e.Data = data[EventMsgHeaderLen:]

	// Detect composite event types
	switch e.Type {
	case EventTypeAlarmFired:
		if len(e.Data) >= 8 {
			e.AlarmData = AlarmEventData{MsgId: binary.LittleEndian.Uint64(e.Data[:8])}
		}
	case EventTypeRadioCommTx:
		e.TxData = deserializeRadioTxData(e.Data)
		payloadOffset += TxEventDataHeaderLen
		simplelogger.AssertEqual(e.TxData.Channel, e.Data[payloadOffset]) // channel is stored twice.
	case EventTypeChannelSample:
		e.ChanData = deserializeChanData(e.Data)
		payloadOffset += ChanEventDataHeaderLen
	case EventTypeRadioState:
		e.RadioStateData = deserializeRadioStateData(e.Data)
		payloadOffset += RadioStateEventDataHeaderLen
	default:
		break
	}

	data2 := make([]byte, datalen-payloadOffset)
	copy(data2, e.Data[payloadOffset:])
	e.Data = data2

	// e.Timestamp is not in the event, so set to invalid initially.
	e.Timestamp = InvalidTimestamp
}

func deserializeRadioTxData(data []byte) TxEventData {
	simplelogger.AssertTrue(len(data) >= TxEventDataHeaderLen)
	txData := TxEventData{Channel: data[0], TxPower: int8(data[1])}
	return txData
}

func deserializeChanData(data []byte) ChanEventData {
	simplelogger.AssertTrue(len(data) >= ChanEventDataHeaderLen)
	chanData := ChanEventData{
		Channel: data[0],
	}
	return chanData
}

func deserializeChanDoneData(data []byte) ChanDoneEventData {
	simplelogger.AssertTrue(len(data) >= ChanDoneEventDataHeaderLen)
	chanData := ChanDoneEventData{
		Channel: data[0],
		Rssi:    int8(data[1]),
	}
	return chanData
}

func deserializeRadioStateData(data []byte) RadioStateEventData {
	simplelogger.AssertTrue(len(data) >= RadioStateEventDataHeaderLen)
	stateData := RadioStateEventData{
		Channel: data[0],
		State:   RadioStates(data[1]),
		TxPower: int8(data[2]),
	}
	return stateData
}

func serializeRadioRxData(rxData *RxEventData) []byte {
	b := []byte{0, 0, 0}
	b[0] = rxData.Channel
	b[1] = rxData.Error
	b[2] = byte(rxData.Rssi)
	return b
}

func serializeRadioTxDoneData(txDoneData *TxDoneEventData) []byte {
	b := []byte{0, 0}
	b[0] = txDoneData.Channel
	b[1] = txDoneData.Error
	return b
}

func serializeChanDoneData(chanData *ChanDoneEventData) []byte {
	b := []byte{0, 0}
	b[0] = chanData.Channel
	b[1] = byte(chanData.Rssi)
	return b
}

// Copy creates a (struct) copy of the Event.
func (e Event) Copy() Event {
	newEv := e
	return newEv
}

func (e *Event) String() string {
	paylStr := ""
	if len(e.Data) > 0 {
		paylStr = fmt.Sprintf(",payl=%v", keepPrintableChars(string(e.Data)))
	}
	s := fmt.Sprintf("Ev{%2d,dly=%v%v}", e.Type, e.Delay, paylStr)
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
