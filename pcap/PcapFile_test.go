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
	pcap, err := NewFile("test.pcap")
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

	assert.True(t, pcapFileHeaderSize == getFileSize(t, "test.pcap"))

	for i := 0; i < 10; i++ {
		err = pcap.AppendFrame(0, []byte{0x0})
		if err != nil {
			t.Fatal(err)
		}

		err = pcap.Sync()
		if err != nil {
			t.Fatal(err)
		}
		assert.True(t, pcapFileHeaderSize+(pcapFrameHeaderSize+1)*(i+1) == getFileSize(t, "test.pcap"))
	}
}

func getFileSize(t *testing.T, fp string) int {
	info, err := os.Stat(fp)
	if err != nil {
		t.Fatal(err)
	}

	return int(info.Size())
}
