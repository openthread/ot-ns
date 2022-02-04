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
	"net"
)

const (
	EventTypeAlarmFired                 uint8 = 0
	EventTypeRadioFrameToNode           uint8 = 1
	EventTypeUartWrite                  uint8 = 2
	EventTypeStatusPush                 uint8 = 5
	EventTypeRadioTxDone                uint8 = 6
	EventTypeRadioFrameToSim            uint8 = 8
	EventTypeRadioFrameAckToSim         uint8 = 9
	EventTypeRadioFrameToNodeInterfered uint8 = 10
)

type eventType = uint8

type Event struct {
	Timestamp  uint64
	Delay      uint64
	Type       eventType
	Param1     int8
	Param2     int8
	NodeId     NodeId
	Data       []byte
	SrcAddr    *net.UDPAddr
	IsInternal bool
}

/* RadioMessagePsduOffset is the offset of Psdu data in a received OpenThread RadioMessage type.
type RadioMessage struct {
	Channel       uint8
	Psdu          byte[]
}
*/
const RadioMessagePsduOffset = 1

// Serialize serializes this Event into []byte to send to OpenThread node, including fields partially.
func (e *Event) Serialize() []byte {
	msg := make([]byte, 17+len(e.Data))
	// e.Timestamp is not sent, only e.Delay.
	binary.LittleEndian.PutUint64(msg[:8], e.Delay)
	binary.LittleEndian.PutUint32(msg[8:12], uint32(e.NodeId))
	msg[12] = e.Type
	msg[13] = byte(e.Param1)
	msg[14] = byte(e.Param2)
	binary.LittleEndian.PutUint16(msg[15:17], uint16(len(e.Data)))
	n := copy(msg[17:], e.Data)
	simplelogger.AssertTrue(n == len(e.Data))
	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into Event object e.
func (e *Event) Deserialize(data []byte) {
	var n uint16
	n = uint16(len(data))
	if n < 17 {
		simplelogger.Panicf("Event.Deserialize() message length too short: %d", n)
	}

	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.NodeId = NodeId(binary.LittleEndian.Uint32(data[8:12]))
	e.Type = data[12]
	e.Param1 = int8(data[13])
	e.Param2 = int8(data[14])
	datalen := binary.LittleEndian.Uint16(data[15:17])
	simplelogger.AssertTrue(datalen == (n - 17))
	data2 := make([]byte, datalen)
	copy(data2, data[17:n])
	// e.Timestamp is not deserialized (not present)
	e.Timestamp = 0
	e.Data = data2
}
