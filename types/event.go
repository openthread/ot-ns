// Copyright (c) 2020, The OTNS Authors.
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
)

const (
	EventTypeAlarmFired       uint8 = 0
	EventTypeRadioReceived    uint8 = 1 // Rx of frame from OTNS to OT-node
	EventTypeUartWrite        uint8 = 2
	EventTypeRadioSpinelWrite uint8 = 3
	EventTypeOtnsStatusPush   uint8 = 5

	EventTypeRadioTx           uint8 = 17 // Tx of frame from OT-node to OTNS
	EventTypeRadioTxDone       uint8 = 18 // Tx-done signal from OTNS to OT-node
	EventTypeRadioTxAck        uint8 = 19 // Tx of Ack from OT-node to OTNS
	EventTypeRadioRxInterfered uint8 = 20 // Rx of interfered frame from OTNS to OT-node

	EventTypeV2Format uint8 = 130

	// EventMessageV1HeaderLen int = 11
	EventMessageV2HeaderLen int = 26
)

type eventType = uint8

type Event struct {
	MsgId      uint64
	Timestamp  uint64
	Delay      uint64
	Type       eventType
	Param1     int8
	Param2     int8
	NodeId     NodeId
	Data       []byte
	IsInternal bool
}

/* RadioMessagePsduOffset is the offset of Psdu data in a received OpenThread RadioMessage type.
type RadioMessage struct {
	Channel       uint8
	Psdu          byte[]
}
*/
const RadioMessagePsduOffset = 1

// Serialize serializes this Event into []byte to send to OpenThread node that supports the V2
// event message format (builds with OT_SIMULATION_RF_EXT_MODELS=ON), including fields partially.
func (e *Event) Serialize() []byte {
	msg := make([]byte, EventMessageV2HeaderLen+len(e.Data))
	// e.Timestamp is not sent, only e.Delay.
	binary.LittleEndian.PutUint64(msg[:8], e.Delay)
	msg[8] = EventTypeV2Format
	binary.LittleEndian.PutUint16(msg[9:11], uint16(len(e.Data)))
	msg[11] = e.Type
	binary.LittleEndian.PutUint32(msg[12:16], uint32(e.NodeId))
	msg[16] = byte(e.Param1)
	msg[17] = byte(e.Param2)
	binary.LittleEndian.PutUint64(msg[18:26], e.MsgId)
	n := copy(msg[EventMessageV2HeaderLen:], e.Data)
	simplelogger.AssertTrue(n == len(e.Data))
	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into Event object e.
func (e *Event) Deserialize(data []byte) {
	n := len(data)
	if n < EventMessageV2HeaderLen {
		simplelogger.Panicf("Event.Deserialize() message length too short: %d", n)
	}
	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.Type = data[8]
	datalen := binary.LittleEndian.Uint16(data[9:11])

	simplelogger.AssertTrue(e.Type == EventTypeV2Format)
	e.Type = data[11]
	e.NodeId = NodeId(binary.LittleEndian.Uint32(data[12:16]))
	e.Param1 = int8(data[16])
	e.Param2 = int8(data[17])
	e.MsgId = binary.LittleEndian.Uint64(data[18:26])
	simplelogger.AssertTrue(datalen == uint16(n-EventMessageV2HeaderLen))
	data2 := make([]byte, datalen)
	copy(data2, data[EventMessageV2HeaderLen:n])
	e.Data = data2
	// e.Timestamp is not deserialized (not present)
	e.Timestamp = 0
}
