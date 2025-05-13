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

package pcap

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPcapFile(t *testing.T) {
	pcapFilename := "test.pcap"
	pcap, err := NewFile(pcapFilename, FrameTypeWpan, false)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pcap.Close()
	}()

	err = pcap.Sync()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, pcapFileHeaderSize, getFileSize(t, pcapFilename))

	for i := 0; i < 10; i++ {
		frame := Frame{
			Timestamp: uint64(i) * 1000,
			Data:      []byte{0x12, 0x10, 0xa6, 0x80, 0x65},
			Channel:   12,
			Rssi:      -60.0,
		}
		err = pcap.AppendFrame(frame)
		if err != nil {
			t.Fatal(err)
		}

		err = pcap.Sync()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, pcapFileHeaderSize+(pcapFrameHeaderSize+5)*(i+1) == getFileSize(t, pcapFilename))
	}
}

func TestPcapTapFile(t *testing.T) {
	pcapFilename := "test_tap.pcap"
	pcap, err := NewFile(pcapFilename, FrameTypeWpanTap, false)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pcap.Close()
	}()

	err = pcap.Sync()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, pcapFileHeaderSize, getFileSize(t, pcapFilename))

	for i := 0; i < 10; i++ {
		frame := Frame{
			Timestamp: uint64(i) * 1000,
			Data:      []byte{0x12, 0x10, 0x30, 0x3f, 0x94},
			Channel:   uint8(i + 11),
			Rssi:      -60.0 + float32(i),
		}
		err = pcap.AppendFrame(frame)
		if err != nil {
			t.Fatal(err)
		}

		err = pcap.Sync()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, pcapFileHeaderSize+(pcapFrameHeaderSize+pcapTapFrameHeaderSize+5)*(i+1), getFileSize(t, pcapFilename))
	}
}

func TestPcapFileWithTimeRefFrame(t *testing.T) {
	pcapFilename := "test_timerefframe.pcap"
	pcap, err := NewFile(pcapFilename, FrameTypeWpan, true)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = pcap.Close()
	}()

	err = pcap.Sync()
	if err != nil {
		t.Fatal(err)
	}
	pcapStartSize := getFileSize(t, pcapFilename)

	for i := 0; i < 10; i++ {
		frame := Frame{
			Timestamp: 500 + uint64(i)*1000,
			Data:      []byte{0x12, 0x10, 0xa6, 0x80, 0x65},
			Channel:   25,
			Rssi:      0.0,
		}
		err = pcap.AppendFrame(frame)
		if err != nil {
			t.Fatal(err)
		}

		err = pcap.Sync()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, pcapStartSize+(pcapFrameHeaderSize+5)*(i+1) == getFileSize(t, pcapFilename))
	}
}

func getFileSize(t *testing.T, fp string) int {
	info, err := os.Stat(fp)
	if err != nil {
		t.Fatal(err)
	}

	return int(info.Size())
}
