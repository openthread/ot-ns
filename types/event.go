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

	"github.com/simonlingoogle/go-simplelogger"
	"math"
)

type eventType = uint8

const (
	// Event type IDs (external, shared between OT-NS and OT node)
	EventTypeAlarmFired       eventType = 0
	EventTypeRadioReceived    eventType = 1 // Rx of frame from OTNS to OT-node
	EventTypeUartWrite        eventType = 2
	EventTypeRadioSpinelWrite eventType = 3
	EventTypePostCmd          eventType = 4
	EventTypeOtnsStatusPush   eventType = 5
	EventTypeRadioRx          eventType = 16 // Rx of frame from OTNS to OT-node with additional status info
	EventTypeRadioTx          eventType = 17 // Tx of frame from OT-node to OTNS
	EventTypeRadioTxDone      eventType = 18 // Tx-done signal from OTNS to OT-node

	// Internal radiomodel events
	EventTypeRadioTxOngoing eventType = 128
)

const EventMsgHeaderLen = 11 // from OT platform-simulation.h struct Event { }
type Event struct {
	Delay uint64
	Type  eventType
	//DataLen uint16
	Data []byte

	// supplementary information stored depending on event type.
	NodeId     NodeId
	Timestamp  uint64
	TxData     TxEventData
	RxData     RxEventData
	TxDoneData TxDoneEventData
}

const TxEventDataHeaderLen = 3 // from OT platform-simulation.h struct
type TxEventData struct {
	Channel    uint8
	TxPower    int8
	CcaEdTresh int8
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
	// Detect composite event types
	extraFields := []byte{}
	switch e.Type {
	case EventTypeRadioRx:
		extraFields = serializeRadioRxData(e.RxData)
	default:
		break
	}

	payload := append(extraFields, e.Data...)
	msg := make([]byte, EventMsgHeaderLen+len(payload))
	// e.Timestamp is not sent, only e.Delay.
	binary.LittleEndian.PutUint64(msg[:8], e.Delay)
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
	case EventTypeRadioTx:
		e.TxData = deserializeRadioTxData(e.Data)
		payloadOffset += TxEventDataHeaderLen
	default:
		break
	}

	data2 := make([]byte, datalen-payloadOffset)
	copy(data2, e.Data[payloadOffset:])
	e.Data = data2

	// e.Timestamp is not in the event, so set to invalid initially.
	e.Timestamp = math.MaxUint64
}

// DeserializeRadioTxEvent deserializes the specific extra TxEvent parameters that are provided in the
// RadioTx event.
func deserializeRadioTxData(data []byte) TxEventData {
	n := len(data)
	simplelogger.AssertTrue(n >= TxEventDataHeaderLen)
	txData := TxEventData{Channel: data[0], TxPower: int8(data[1]), CcaEdTresh: int8(data[2])}
	return txData
}

func serializeRadioRxData(rxData RxEventData) []byte {
	b := []byte{0, 0, 0}
	b[0] = rxData.Channel
	b[1] = rxData.Error
	b[2] = uint8(rxData.Rssi)
	return b
}

// Copy creates a (struct) copy of the Event.
func (e Event) Copy() Event {
	newEv := e
	return newEv
}
