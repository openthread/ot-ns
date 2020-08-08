package main

import (
	"testing"
	"time"

	"github.com/openthread/ot-ns/otnstester"
)

func TestAdd(t *testing.T) {
	ot := otnstester.NewOtnsTest(t)

	defer ot.Shutdown()

	testAddNode(ot)
}

func testAddNode(test *otnstester.OtnsTest) {
	test.Reset()

	nodeid := test.AddNode("router")
	test.ExpectTrue(nodeid == 1)
	test.Go(time.Second * 3)
	test.ExpectTrue(test.GetNodeState(nodeid) == RoleLeader)
	test.ExpectVisualizeAddNode(nodeid, 0, 0, DefaultRadioRange)

	router2 := test.AddNode("router")
	test.ExpectTrue(router2 == 2)
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(router2) == RoleRouter)
	test.ExpectVisualizeAddNode(router2, 0, 0, DefaultRadioRange)

	test.Command("add fed x 50 y 60")
	fed := 3
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(fed) == RoleChild)
	test.ExpectVisualizeAddNode(fed, 50, 60, DefaultRadioRange)

	fedInfo := test.ListNodes()[fed]
	test.ExpectTrue(fedInfo.X == 50)
	test.ExpectTrue(fedInfo.Y == 60)

	test.Command("add med x 10 y 20 rr 100")
	med := 4
	test.Go(time.Second * 10)
	test.ExpectTrue(test.GetNodeState(med) == RoleChild)
	test.ExpectVisualizeAddNode(med, 10, 20, 100)

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
