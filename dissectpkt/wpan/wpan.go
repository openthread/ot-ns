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

package wpan

import (
	"encoding/binary"
	"fmt"

	"github.com/openthread/ot-ns/types"
)

type FrameType = uint16

const (
	FrameTypeBeacon  FrameType = 0
	FrameTypeData    FrameType = 1
	FrameTypeAck     FrameType = 2
	FrameTypeCommand FrameType = 3
)

// Values for both Src and Dst addressing modes, Table 7-3, 802.15.4-2015.
const (
	AddrModeNone     = 0
	AddrModeReserved = 1
	AddrModeShort    = 2
	AddrModeExtended = 3
)

type FrameControl uint16

func (fc FrameControl) String() string {
	return fmt.Sprintf("0x%04x", uint16(fc))
}

func (fc FrameControl) FrameType() FrameType {
	return FrameType(fc & 0x0007)
}

func (fc FrameControl) SecurityEnabled() bool {
	return (fc & 0x0008) != 0
}

func (fc FrameControl) FramePending() bool {
	return (fc & 0x0010) != 0
}

func (fc FrameControl) AckRequest() bool {
	return (fc & 0x0020) != 0
}

func (fc FrameControl) PanidCompression() bool {
	return (fc & 0x0040) != 0
}

func (fc FrameControl) SequenceNumberSuppression() bool {
	return (fc & 0x0100) != 0
}

func (fc FrameControl) IEPresent() bool {
	return (fc & 0x0200) != 0
}

func (fc FrameControl) DestAddrMode() uint16 {
	return uint16((fc & 0x0c00) >> 10)
}

func (fc FrameControl) SourceAddrMode() uint16 {
	return uint16((fc & 0xc000) >> 14)
}

func (fc FrameControl) FrameVersion() uint16 {
	return uint16((fc & 0x3000) >> 12)
}

func (fc *FrameControl) Dissect(bytes []byte) {
	*fc = FrameControl(binary.LittleEndian.Uint16(bytes))
}

func (fc *FrameControl) HasDestPanIdField() bool {
	if fc.FrameVersion() <= 1 {
		return true
	}
	dam := fc.DestAddrMode()
	sam := fc.SourceAddrMode()
	if dam != AddrModeNone && sam != AddrModeNone {
		return true
	}
	pc := fc.PanidCompression()
	if dam == AddrModeExtended && sam == AddrModeExtended {
		return !pc
	}
	if sam == AddrModeNone && dam != AddrModeNone && !pc {
		return true
	}
	if sam == AddrModeNone && dam == AddrModeNone && pc {
		return true
	}
	return false
}

func (fc *FrameControl) HasSourcePanIdField() bool {
	dam := fc.DestAddrMode()
	sam := fc.SourceAddrMode()
	pc := fc.PanidCompression()
	if fc.FrameVersion() <= 1 {
		if sam != AddrModeNone && !pc {
			return true
		}
		return false
	}
	if dam == AddrModeExtended && sam == AddrModeExtended && !pc {
		return false
	}
	if sam == AddrModeNone {
		return false
	}
	return !pc
}

type MacFrame struct {
	Channel         uint8
	FrameControl    FrameControl
	Seq             uint8
	DstPanId        uint16
	SrcPanId        uint16
	DstAddrShort    uint16
	SrcAddrShort    uint16
	DstAddrExtended uint64
	SrcAddrExtended uint64
	LengthBytes     uint16
	PhyHdrLength    uint16
}

func (f *MacFrame) String() string {
	if f.FrameControl.FrameType() == FrameTypeAck {
		return fmt.Sprintf("ACK,FC:%s,Seq:%d", f.FrameControl, f.Seq)
	}

	var dstAddrS string
	dstAddrMode := f.FrameControl.DestAddrMode()
	if dstAddrMode == AddrModeShort {
		dstAddrS = fmt.Sprintf("%04x", f.DstAddrShort)
	} else if dstAddrMode == AddrModeExtended {
		dstAddrS = fmt.Sprintf("%016x", f.DstAddrExtended)
	} else {
		dstAddrS = "-"
	}

	return fmt.Sprintf("MAC,FC:%s,Seq:%d,Dst:%s", f.FrameControl, f.Seq, dstAddrS)
}

func Dissect(data []byte) *MacFrame {
	frame := &MacFrame{}
	frame.Channel = data[0] // not part of 802.15.4, but part of our custom frame format.
	frame.LengthBytes = uint16(len(data) - 1)
	frame.PhyHdrLength = types.PhyHeaderLenBytes
	frame.FrameControl.Dissect(data[1:3])
	if frame.FrameControl.FrameType() > FrameTypeCommand {
		return frame // for unsupported frame types.
	}

	n := 3
	if !frame.FrameControl.SequenceNumberSuppression() {
		frame.Seq = data[n]
		n += 1
	}
	if frame.FrameControl.HasDestPanIdField() {
		frame.DstPanId = binary.LittleEndian.Uint16(data[n : n+2])
		n += 2
	}

	switch frame.FrameControl.DestAddrMode() {
	case AddrModeExtended:
		frame.DstAddrExtended = binary.LittleEndian.Uint64(data[n : n+8])
		n += 8
	case AddrModeShort:
		frame.DstAddrShort = binary.LittleEndian.Uint16(data[n : n+2])
		n += 2
	default:
		break
	}

	if frame.FrameControl.HasSourcePanIdField() {
		frame.SrcPanId = binary.LittleEndian.Uint16(data[n : n+2])
		n += 2
	}

	switch frame.FrameControl.SourceAddrMode() {
	case AddrModeExtended:
		frame.SrcAddrExtended = binary.LittleEndian.Uint64(data[n : n+8])
	case AddrModeShort:
		frame.SrcAddrShort = binary.LittleEndian.Uint16(data[n : n+2])
	default:
		break
	}

	return frame
}
