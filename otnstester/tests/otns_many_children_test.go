// Copyright (c) 2023-2024, The OTNS Authors.
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

package main

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/openthread/ot-ns/otnstester"
)

func TestAddManySEDs(t *testing.T) {
	ot := otnstester.Instance(t)
	testAddManySEDs(ot)
	ot.Reset()
}

func testAddManySEDs(test *otnstester.OtnsTest) {
	test.Start("testAddManySEDs")
	test.Command("radiomodel MutualInterference")

	x := 500
	y := 500
	nodeid := test.AddNode("router", x, y)
	test.ExpectEqual(1, nodeid)
	test.Go(time.Second * 10)
	test.ExpectEqual(RoleLeader, test.GetNodeState(nodeid))

	N := 10 // number of SED Children - 10 is the limit for a minimally Thread compliant Router.
	var r float64
	for n := 1; n <= N; n++ {
		fra := float64(n) / float64(N)
		r = rand.Float64()*60.0 + 60.0
		test.AddNodeRr("sed", int(float64(x)+r*math.Sin(2.0*math.Pi*fra)),
			int(float64(y)+r*math.Cos(2.0*math.Pi*fra)), 200)
		test.Go(time.Millisecond * 2200)
	}
	test.Go(time.Second * 60)

	for n := 2; n <= N+1; n++ {
		test.ExpectEqual(RoleChild, test.GetNodeState(n))
	}
}
