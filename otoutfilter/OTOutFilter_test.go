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

package otoutfilter

import (
	"io"
	"strings"
	"testing"
)

func TestOTOutFilter(t *testing.T) {
	input := "> cmd1\n" +
		"Done\n" +
		"> cmd2\n" +
		"Error: fail\n" +
		"\n" +
		"> cmd3\n" +
		//-|C|W|N|I|D
		"Any 00:00:17.817 [-]log1\n" +
		"B01:00:17.817 [C]log2\n" +
		"C 02:43:37.817 [W] log3\n" +
		"D  33:00:17.817 [N] log4\n" +
		"E text 44:33:22.123 [I]log5\n" +
		"F text [x] no log\n" +
		"00:00:00.000 [INFO]-CORE----: Notifier: StateChanged (0x01001009) [Ip6+ LLAddr Ip6Mult+ NetifState]\n" +
		"00:00:00.000 [NOTE]-CLI-----: Output: > Done\n" +
		"00:00:00.000 [DEBG]-PLAT----: Clear ExtAddr entries\n" +
		"30:30:23.456 [WARN]-PLAT----: some text\n" +
		"G[C]log2\n" +
		"H[W] log3\n" +
		"I[N] log4\n" +
		"JKL[I]log5\n" +
		"\n[D]log6\n" +
		"Done\n" +
		""
	expectOutput := "cmd1\n" +
		"Done\n" +
		"cmd2\n" +
		"Error: fail\n" +
		"\n" +
		"cmd3\n" +
		"" +
		"F text [x] no log\n" +
		"\n" +
		"Done\n" +
		""

	r := NewOTOutFilter(strings.NewReader(input), "Node<1> - ", nil)
	output, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(output) != expectOutput {
		t.Fatalf("output %#v, expect: %#v", string(output), expectOutput)
	}
}
