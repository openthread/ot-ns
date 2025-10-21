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

package otoutfilter

import (
	"io/ioutil"
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
		//NONE|CRIT|WARN|NOTE|INFO|DEBG
		"A[NONE]log1\n" +
		"B[CRIT]log2\n" +
		"C[WARN]log3\n" +
		"D[NOTE]log4\n" +
		"E[INFO]log5\n" +
		"\n[DEBG]log6\n" +
		"Done\n" +
		""
	expectOutput := "cmd1\n" +
		"Done\n" +
		"cmd2\n" +
		"Error: fail\n" +
		"\n" +
		"cmd3\n" +
		"ABCDE\n" +
		"Done\n" +
		""

	r := NewOTOutFilter(strings.NewReader(input), "Node<1>")
	output, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	if string(output) != expectOutput {
		t.Fatalf("output %#v, expect: %#v", string(output), expectOutput)
	}
}
