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

package pcap

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
)

type FrameType int

const (
	FrameTypeOff FrameType = iota
	FrameTypeWpan
	FrameTypeWpanTap
	FrameTypeUnknown
)

const (
	FrameTypeOffStr     string = "off"
	FrameTypeWpanStr    string = "wpan"
	FrameTypeWpanTapStr string = "wpan-tap"
)

const (
	dltIeee802154       = 195
	pcapMagicNumber     = 0xA1B2C3D4
	pcapVersionMajor    = 2
	pcapVersionMinor    = 4
	pcapFileHeaderSize  = 24
	pcapFrameHeaderSize = 16
)

const (
	// 802.15.4 frame that uses a Reserved type and is only included as t=0 simulation time reference for the PCAP file.
	timeReference802154frameData string = "\x04\x21This is an OTNS simulation PCAP-start t=0 reference frame.\x61\x3f"
)

// File represents a PCAP file
type File interface {
	AppendFrame(frame Frame) error
	Sync() error
	Close() error
}

// Frame represents a single radio frame that can be added to a PCAP file
type Frame struct {
	Timestamp uint64
	Data      []byte
	Channel   ChannelId
	Rssi      float32
}

type wpanFile struct {
	fd *os.File
}

// NewFile creates a new PCAP file with all frames using specified frameType
func NewFile(filename string, frameType FrameType, useTimeRefFrame bool) (File, error) {
	var f File
	var err error

	switch frameType {
	case FrameTypeWpan:
		f, err = newWpanFile(filename)
	case FrameTypeWpanTap:
		f, err = newWpanTapFile(filename)
	default:
		f, err = nil, fmt.Errorf("invalid PCAP frame type: %d", frameType)
	}

	if useTimeRefFrame && err == nil && f != nil {
		logger.PanicfIfError(f.AppendFrame(Frame{
			Timestamp: 0,
			Data:      []byte(timeReference802154frameData),
			Channel:   0,
			Rssi:      0,
		}), "PCAP file time-reference 0 frame could not be written")
	}

	return f, err
}

func ParseFrameTypeStr(tp string) FrameType {
	switch tp {
	case FrameTypeOffStr:
		return FrameTypeOff
	case FrameTypeWpanStr:
		return FrameTypeWpan
	case FrameTypeWpanTapStr:
		return FrameTypeWpanTap
	default:
		return FrameTypeUnknown
	}
}

func newWpanFile(filename string) (File, error) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	pf := &wpanFile{
		fd: fd,
	}

	if err = pf.writeHeader(); err != nil {
		_ = pf.Close()
		return nil, err
	}

	return pf, nil
}

func (pf *wpanFile) AppendFrame(frame Frame) error {
	var header [pcapFrameHeaderSize]byte
	sec := uint32(frame.Timestamp / 1000000)
	usec := uint32(frame.Timestamp % 1000000)
	binary.LittleEndian.PutUint32(header[:4], sec)
	binary.LittleEndian.PutUint32(header[4:8], usec)
	plen := uint32(len(frame.Data))
	binary.LittleEndian.PutUint32(header[8:12], plen)
	binary.LittleEndian.PutUint32(header[12:16], plen)

	var err error

	_, err = pf.fd.Write(header[:])
	if err != nil {
		return err
	}

	_, err = pf.fd.Write(frame.Data)
	return err
}

func (pf *wpanFile) Sync() error {
	return pf.fd.Sync()
}

func (pf *wpanFile) Close() error {
	return pf.fd.Close()
}

func (pf *wpanFile) writeHeader() error {
	var header [pcapFileHeaderSize]byte
	binary.LittleEndian.PutUint32(header[:4], pcapMagicNumber)
	binary.LittleEndian.PutUint16(header[4:6], pcapVersionMajor)
	binary.LittleEndian.PutUint16(header[6:8], pcapVersionMinor)
	binary.LittleEndian.PutUint32(header[8:12], 0)
	binary.LittleEndian.PutUint32(header[12:16], 0)
	binary.LittleEndian.PutUint32(header[16:20], 256)
	binary.LittleEndian.PutUint32(header[20:24], dltIeee802154)
	if _, err := pf.fd.Write(header[:]); err != nil {
		return err
	}
	return pf.fd.Sync()
}
