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

package main

import (
	"testing"
	"time"

	"github.com/openthread/ot-ns/otnstester"
)

func TestAddNodes(t *testing.T) {
	ot := otnstester.Instance(t)
	testAddNodes(ot)
	ot.Reset()
}

func testAddNodes(test *otnstester.OtnsTest) {
	test.Reset()

	nodeid := test.AddNode("router", 100, 100)
	test.ExpectTrue(nodeid == 1)
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(nodeid) == RoleLeader)
	test.ExpectVisualizeAddNode(nodeid, 100, 100, DefaultRadioRange)

	router2 := test.AddNode("router", 120, 120)
	test.ExpectTrue(router2 == 2)
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(router2) == RoleRouter)
	test.ExpectVisualizeAddNode(router2, 120, 120, DefaultRadioRange)

	test.Command("add fed x 50 y 60")
	fed := 3
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(fed) == RoleChild)
	test.ExpectVisualizeAddNode(fed, 50, 60, DefaultRadioRange)

	fedInfo := test.ListNodes()[fed]
	test.ExpectTrue(fedInfo.X == 50)
	test.ExpectTrue(fedInfo.Y == 60)

	test.Command("add med x 10 y 20 rr 121")
	med := 4
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(med) == RoleChild)
	test.ExpectVisualizeAddNode(med, 10, 20, 121)

	medInfo := test.ListNodes()[med]
	test.ExpectTrue(medInfo.X == 10)
	test.ExpectTrue(medInfo.Y == 20)

	test.Command("add sed x 30 y 40")
	sed := 5
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(sed) == RoleChild)
	test.ExpectVisualizeAddNode(sed, 30, 40, DefaultRadioRange)
	sedInfo := test.ListNodes()[sed]
	test.ExpectTrue(sedInfo.X == 30)
	test.ExpectTrue(sedInfo.Y == 40)
}
