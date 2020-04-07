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

package simulation

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

type simulationController struct {
	sim *Simulation
}

func (sc *simulationController) CtrlSetSpeed(speed float64) error {
	sim := sc.sim
	sim.PostAsync(true, func() {
		sim.SetSpeed(speed)
	})
	return nil
}

func (sc *simulationController) CtrlSetNodeFailed(nodeid NodeId, failed bool) error {
	sim := sc.sim
	sim.PostAsync(true, func() {
		sim.SetNodeFailed(nodeid, failed)
	})
	return nil
}

func (sc *simulationController) CtrlDeleteNode(nodeid NodeId) error {
	sim := sc.sim
	sim.PostAsync(true, func() {
		_ = sim.DeleteNode(nodeid)
	})
	return nil
}

func (sc *simulationController) CtrlMoveNodeTo(nodeid NodeId, x, y int) error {
	sim := sc.sim
	sim.PostAsync(true, func() {
		sim.MoveNodeTo(nodeid, x, y)
	})
	return nil
}

func (sc *simulationController) CtrlAddNode(x, y int, isRouter bool, mode NodeMode) error {
	sim := sc.sim
	nodeCfg := DefaultNodeConfig()
	nodeCfg.IsRouter = isRouter
	nodeCfg.IsMtd = !isRouter && !mode.FullThreadDevice
	nodeCfg.RxOffWhenIdle = !isRouter && !mode.RxOnWhenIdle
	nodeCfg.X, nodeCfg.Y = x, y

	sim.PostAsync(true, func() {
		simplelogger.Infof("CtrlAddNode: %+v", nodeCfg)
		_, err := sim.AddNode(nodeCfg)
		if err != nil {
			simplelogger.Errorf("add node failed: %v", err)
			return
		}
	})
	return nil
}

type readonlySimulationController struct {
}

func (r readonlySimulationController) CtrlSetSpeed(speed float64) error {
	return readonlySimulationError
}

func (r readonlySimulationController) CtrlSetNodeFailed(nodeid NodeId, failed bool) error {
	return readonlySimulationError
}

var readonlySimulationError = errors.Errorf("simulation is readonly")

func (r readonlySimulationController) CtrlAddNode(x, y int, router bool, mode NodeMode) error {
	return readonlySimulationError
}

func (r readonlySimulationController) CtrlMoveNodeTo(nodeid NodeId, x, y int) error {
	return readonlySimulationError
}

func (r readonlySimulationController) CtrlDeleteNode(nodeid NodeId) error {
	return readonlySimulationError
}

func NewSimulationController(sim *Simulation) visualize.SimulationController {
	if !sim.cfg.ReadOnly {
		return &simulationController{sim}
	} else {
		return readonlySimulationController{}
	}
}
