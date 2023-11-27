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

package pcap

import (
	"encoding/binary"
	"math"
	"os"
)

// wpan-tap / DLT IEEE802 15 4 TAP specification is at
// https://gitlab.com/exegin/ieee802-15-4-tap
const (
	dltIeee802154Tap       = 283
	pcapTapFrameHeaderSize = 28
)

const (
	tlvFcsType           = 0
	tlvRss               = 1
	tlvChannelAssignment = 3
	tlvSofTimestamp      = 5
	tlvLqi               = 10
)

type tapFile struct {
	fd *os.File
}

func newWpanTapFile(filename string) (File, error) {
	fd, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	pf := &tapFile{
		fd: fd,
	}

	if err = pf.writeHeader(); err != nil {
		_ = pf.Close()
		return nil, err
	}

	return pf, nil
}

func setTlv(hdr *[pcapFrameHeaderSize + pcapTapFrameHeaderSize]byte, idx *int, tlvType uint16, data []byte) {
	var l uint16
	lenData := uint16(len(data))
	l = lenData & 0xFFFC
	if lenData&0x0003 > 0 {
		l += 4
	}
	tlv := make([]byte, 4+l)
	binary.LittleEndian.PutUint16(tlv[0:2], tlvType)
	binary.LittleEndian.PutUint16(tlv[2:4], lenData)
	copy(tlv[4:], data[:])
	copy(hdr[*idx:], tlv[:])
	*idx += int(4 + l)
}

func (pf *tapFile) AppendFrame(frame Frame) error {
	var header [pcapFrameHeaderSize + pcapTapFrameHeaderSize]byte

	sec := uint32(frame.Timestamp / 1000000)
	usec := uint32(frame.Timestamp % 1000000)
	binary.LittleEndian.PutUint32(header[:4], sec)
	binary.LittleEndian.PutUint32(header[4:8], usec)
	frLen := uint32(len(frame.Data)) + pcapTapFrameHeaderSize
	binary.LittleEndian.PutUint32(header[8:12], frLen)
	binary.LittleEndian.PutUint32(header[12:pcapFrameHeaderSize], frLen)

	// append fields per wpan-tap spec https://gitlab.com/exegin/ieee802-15-4-tap
	n := pcapFrameHeaderSize
	header[n] = 0 // wpan-tap version
	n += 1
	header[n] = 0 // reserved
	n += 1
	binary.LittleEndian.PutUint16(header[n:n+2], pcapTapFrameHeaderSize) // header len
	n += 2
	setTlv(&header, &n, tlvFcsType, []byte{1}) // 1 == 16-bit CRC (=FCS)
	rssFloat := frame.Rssi
	rssValue := make([]byte, 4)
	binary.LittleEndian.PutUint32(rssValue, math.Float32bits(rssFloat))
	setTlv(&header, &n, tlvRss, rssValue)
	channelAssign := make([]byte, 3)
	binary.LittleEndian.PutUint16(channelAssign, uint16(frame.Channel))
	channelAssign[2] = 0 // 0 == IEEE 802.15.4 channel page 0
	setTlv(&header, &n, tlvChannelAssignment, channelAssign)

	var err error

	_, err = pf.fd.Write(header[:])
	if err != nil {
		return err
	}

	_, err = pf.fd.Write(frame.Data)
	return err
}

func (pf *tapFile) Sync() error {
	return pf.fd.Sync()
}

func (pf *tapFile) Close() error {
	return pf.fd.Close()
}

func (pf *tapFile) writeHeader() error {
	var header [pcapFileHeaderSize]byte
	binary.LittleEndian.PutUint32(header[:4], pcapMagicNumber)
	binary.LittleEndian.PutUint16(header[4:6], pcapVersionMajor)
	binary.LittleEndian.PutUint16(header[6:8], pcapVersionMinor)
	binary.LittleEndian.PutUint32(header[8:12], 0)
	binary.LittleEndian.PutUint32(header[12:16], 0)
	binary.LittleEndian.PutUint32(header[16:20], 256)
	binary.LittleEndian.PutUint32(header[20:24], dltIeee802154Tap)
	if _, err := pf.fd.Write(header[:]); err != nil {
		return err
	}
	return pf.fd.Sync()
}
