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
	EventTypeAlarmFired            uint8 = 0
	EventTypeRadioFrameToNode      uint8 = 1
	EventTypeUartWrite             uint8 = 2
	EventTypeStatusPush            uint8 = 5
	EventTypeRadioTxDone           uint8 = 6
	EventTypeRadioFrameToSim       uint8 = 8
	EventTypeRadioFrameSimInternal uint8 = 9
)

type eventType = uint8

type Event struct {
	Timestamp uint64
	Delay     uint64
	Type      eventType
	Param     int8
	NodeId    NodeId
	Data      []byte
	SrcAddr   *net.UDPAddr
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
	msg := make([]byte, 12+len(e.Data))
	// e.Timestamp is not sent, only e.Delay.
	binary.LittleEndian.PutUint64(msg[:8], e.Delay)
	msg[8] = e.Type
	msg[9] = byte(e.Param)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(len(e.Data)))
	n := copy(msg[12:], e.Data)
	simplelogger.AssertTrue(n == len(e.Data))
	return msg
}

// Deserialize deserializes []byte Event fields (as received from OpenThread node) into Event object e.
func (e *Event) Deserialize(data []byte) {
	var n uint16
	n = uint16(len(data))
	if n < 12 {
		simplelogger.Panicf("Event.Deserialize() message length too short: %d", n)
	}

	e.Delay = binary.LittleEndian.Uint64(data[:8])
	e.Type = data[8]
	e.Param = int8(data[9])
	datalen := binary.LittleEndian.Uint16(data[10:12])
	simplelogger.AssertTrue(datalen == (n - 12))
	data2 := make([]byte, datalen)
	copy(data2, data[12:n])
	// e.Timestamp is not deserialized (not present)
	e.Data = data2
}
