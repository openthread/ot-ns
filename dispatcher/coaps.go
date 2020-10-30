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

package dispatcher

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

type CoapType int

const (
	CoapTypeConfirmable    CoapType = 0
	CoapTypeNonConfirmable CoapType = 1
	CoapTypeAcknowledgment CoapType = 2
	CoapTypeReset          CoapType = 3
)

type CoapCode int

type CoapMessageRecvInfo struct {
	Timestamp uint64 `yaml:"time"`
	DstNode   NodeId `yaml:"dst"`
	SrcAddr   string `yaml:"src_addr"`
	SrcPort   int    `yaml:"src_port"`
}

type CoapMessage struct {
	Timestamp uint64                `yaml:"time"`
	SrcNode   NodeId                `yaml:"src"`
	ID        int                   `yaml:"id"`
	Type      CoapType              `yaml:"type"`
	Code      CoapCode              `yaml:"code"`
	URI       string                `yaml:"uri,omitempty"`
	DstAddr   string                `yaml:"dst_addr"`
	DstPort   int                   `yaml:"dst_port"`
	Error     string                `yaml:"error,omitempty"`
	Receivers []CoapMessageRecvInfo `yaml:"receivers,flow"`
}

type coapsHandler struct {
	messages []*CoapMessage
}

func (coaps *coapsHandler) OnSend(curTime uint64, nodeId NodeId, messageId int, coapType CoapType, coapCode CoapCode, uri string, peerAddr string, peerPort int) {
	coaps.messages = append(coaps.messages, &CoapMessage{
		Timestamp: curTime,
		SrcNode:   nodeId,
		ID:        messageId,
		Type:      coapType,
		Code:      coapCode,
		URI:       uri,
		DstAddr:   peerAddr,
		DstPort:   peerPort,
	})
}

func (coaps *coapsHandler) OnRecv(curTime uint64, nodeId NodeId, messageId int, coapType CoapType, coapCode CoapCode, uri string, peerAddr string, peerPort int) {
	msg := coaps.findMessage(messageId, coapType, coapCode, uri)
	if msg == nil {
		simplelogger.Warnf("CoAP message %d,%d,%d,%s not sent but received by Node %d", messageId, coapType, coapCode, uri, nodeId)
		return
	}

	msg.Receivers = append(msg.Receivers, CoapMessageRecvInfo{
		Timestamp: curTime,
		DstNode:   nodeId,
		SrcAddr:   peerAddr,
		SrcPort:   peerPort,
	})
}

func (coaps *coapsHandler) OnSendError(nodeId NodeId, messageId int, coapType CoapType, coapCode CoapCode, uri string, peerAddr string, peerPort int, error string) {
	msg := coaps.findMessage(messageId, coapType, coapCode, uri)
	if msg == nil {
		simplelogger.Warnf("CoAP message %d,%d,%d,%s not sent but received by Node %d", messageId, coapType, coapCode, uri, nodeId)
		return
	}

	msg.Error = error
}

func (coaps *coapsHandler) findMessage(id int, coapType CoapType, coapCode CoapCode, uri string) *CoapMessage {
	for i := len(coaps.messages) - 1; i >= 0; i-- {
		msg := coaps.messages[i]
		if msg.ID == id && msg.Type == coapType && msg.Code == coapCode && msg.URI == uri {
			return msg
		}
	}

	return nil
}

func (coaps *coapsHandler) DumpMessages() (ret []*CoapMessage) {
	ret, coaps.messages = coaps.messages, nil
	return
}

func newCoapsHandler() *coapsHandler {
	coaps := &coapsHandler{
		messages: nil,
	}
	return coaps
}
