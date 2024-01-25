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
	test.Start("testAddNodes")

	nodeid := test.AddNode("router", 100, 100)
	test.ExpectEqual(1, nodeid)
	test.Go(time.Second * 10)
	test.ExpectEqual(RoleLeader, test.GetNodeState(nodeid))
	test.ExpectVisualizeAddNode(nodeid, 100, 100, DefaultRadioRange)

	router2 := test.AddNode("router", 120, 120)
	test.ExpectEqual(2, router2)
	test.Go(time.Second * 121)
	test.ExpectEqual(RoleRouter, test.GetNodeState(router2))
	test.ExpectVisualizeAddNode(router2, 120, 120, DefaultRadioRange)

	test.Command("add fed x 50 y 60")
	fed := 3
	test.Go(time.Second * 10)
	test.ExpectEqual(RoleChild, test.GetNodeState(fed))
	test.ExpectVisualizeAddNode(fed, 50, 60, DefaultRadioRange)

	fedInfo := test.ListNodes()[fed]
	test.ExpectEqual(50, fedInfo.X)
	test.ExpectEqual(60, fedInfo.Y)
	test.ExpectEqual(0, fedInfo.Z)

	test.Command("add med x 10 y 20 z 1 rr 121")
	med := 4
	test.Go(time.Second * 10)
	test.ExpectEqual(RoleChild, test.GetNodeState(med))
	test.ExpectVisualizeAddNode(med, 10, 20, 121)

	medInfo := test.ListNodes()[med]
	test.ExpectEqual(10, medInfo.X)
	test.ExpectEqual(20, medInfo.Y)
	test.ExpectEqual(1, medInfo.Z)

	test.Command("add sed x 30 y 40")
	sed := 5
	test.Go(time.Second * 10)
	test.ExpectEqual(RoleChild, test.GetNodeState(sed))
	test.ExpectVisualizeAddNode(sed, 30, 40, DefaultRadioRange)
	sedInfo := test.ListNodes()[sed]
	test.ExpectEqual(30, sedInfo.X)
	test.ExpectEqual(40, sedInfo.Y)
	test.ExpectEqual(0, sedInfo.Z)
}

func TestDelManyNodes(t *testing.T) {
	ot := otnstester.Instance(t)
	testDelManyNodes(ot)
	ot.Reset()
}

func testDelManyNodes(test *otnstester.OtnsTest) {
	test.Start("testDelManyNodes")

	for i := 0; i < 32; i++ {
		test.AddNode("router", (i%6)*100, (i/6)*150)
	}

	test.Go(time.Second * 10)
	list := test.ListNodes()
	test.ExpectEqual(32, len(list))

	for i := 0; i < 32; i++ {
		test.DeleteNode(i + 1)
		list = test.ListNodes()
		test.ExpectEqual(31-i, len(list))
		test.Go(time.Second * 5)
	}

	list = test.ListNodes()
	test.ExpectEqual(0, len(list))
}
