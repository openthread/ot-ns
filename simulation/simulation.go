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

// Package simulation implements utilities for running a simulation.
package simulation

import (
	"os"
	"time"

	"github.com/openthread/ot-ns/progctx"

	"github.com/openthread/ot-ns/dispatcher"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

// Simulation manages a running simulation.
type Simulation struct {
	ctx     *progctx.ProgCtx
	cfg     *Config
	nodes   map[NodeId]*Node
	d       *dispatcher.Dispatcher
	vis     visualize.Visualizer
	rawMode bool
}

// NewSimulation creates a new simulation.
func NewSimulation(ctx *progctx.ProgCtx, cfg *Config) (*Simulation, error) {
	s := &Simulation{
		ctx:     ctx,
		cfg:     cfg,
		nodes:   map[NodeId]*Node{},
		rawMode: cfg.RawMode,
	}

	// start the event_dispatcher for virtual time
	dispatcherCfg := dispatcher.DefaultConfig()
	dispatcherCfg.Speed = cfg.Speed
	s.d = dispatcher.NewDispatcher(s.ctx, dispatcherCfg, s)
	s.vis = s.d.GetVisualizer()
	if err := s.removeTmpDir(); err != nil {
		simplelogger.Errorf("remove tmp directory failed: %+v", err)
	}
	return s, nil
}

// AddNode adds a new node to the simulation.
func (s *Simulation) AddNode(cfg *NodeConfig) (*Node, error) {
	if cfg == nil {
		cfg = DefaultNodeConfig()
	}

	nodeid := cfg.ID
	if nodeid <= 0 {
		nodeid = s.genNodeId()
	}

	simplelogger.AssertNil(s.nodes[nodeid])
	node, err := newNode(s, nodeid, cfg)
	if err != nil {
		simplelogger.Errorf("simulation add node failed: %v", err)
		return nil, err
	}

	s.nodes[nodeid] = node

	extaddr := node.GetExtAddr()

	simplelogger.Infof("simulation:CtrlAddNode: %+v, rawMode=%v", cfg, s.rawMode)
	s.d.AddNode(nodeid, extaddr, cfg.X, cfg.Y, cfg.RadioRange, NodeMode{
		RxOnWhenIdle:       !cfg.RxOffWhenIdle,
		SecureDataRequests: true,
		FullThreadDevice:   !cfg.IsMtd,
		FullNetworkData:    true,
	})

	node.setupMode()

	if !s.rawMode {
		node.SetupNetworkParameters(s)
		node.Start()
	}

	return node, nil
}

func (s *Simulation) genNodeId() NodeId {
	nodeid := 1
	for s.nodes[nodeid] != nil {
		nodeid += 1
	}
	return nodeid
}

// Run runs the simulation until exit.
func (s *Simulation) Run() {
	s.ctx.WaitAdd("simulation", 1)
	defer s.ctx.WaitDone("simulation")
	defer simplelogger.Infof("simulation exit.")

	s.d.Run()
}

// Nodes returns all nodes of the simulation.
func (s *Simulation) Nodes() map[NodeId]*Node {
	return s.nodes
}

// MasterKey returns the default Master Key of the simulation.
func (s *Simulation) MasterKey() string {
	return s.cfg.MasterKey
}

// Panid gets the default Pan ID of the simulation.
func (s *Simulation) Panid() uint16 {
	return s.cfg.Panid
}

// Channel gets the default channel of the simulation.
func (s *Simulation) Channel() int {
	return s.cfg.Channel
}

// Stop stops the simulation.
func (s *Simulation) Stop() {
	simplelogger.Infof("stopping simulation ...")
	for _, node := range s.nodes {
		_ = node.Exit()
	}

	s.nodes = nil

	s.d.Stop()
}

// BinDir gets the binary directory for searching OpenThread executables.
func (s *Simulation) BinDir() string {
	return s.cfg.BinDir
}

// SetVisualizer sets the visualizer for the simulation.
func (s *Simulation) SetVisualizer(vis visualize.Visualizer) {
	simplelogger.AssertNotNil(vis)
	s.vis = vis
	s.d.SetVisualizer(vis)
	vis.SetController(NewSimulationController(s))
}

// OnNodeFail notifies the simulation of the failed node.
// It is part of the implementation of dispatcher.CallbackHandler.
func (s *Simulation) OnNodeFail(nodeid NodeId) {
	node := s.nodes[nodeid]
	simplelogger.AssertNotNil(node)
}

// OnNodeRecover notifies the simulation of the recovered node.
// It is part of the implementation of dispatcher.CallbackHandler.
func (s *Simulation) OnNodeRecover(nodeid NodeId) {
	node := s.nodes[nodeid]
	simplelogger.AssertNotNil(node)
}

// PostAsync posts a asynchronous task to be executed in the simulation goroutine.
func (s *Simulation) PostAsync(trivial bool, f func()) {
	s.d.PostAsync(trivial, f)
}

// Dispatcher gets the dispatcher of the simulation.
func (s *Simulation) Dispatcher() *dispatcher.Dispatcher {
	return s.d
}

// MoveNodeTo moves a node to the target position.
func (s *Simulation) MoveNodeTo(nodeid NodeId, x, y int) {
	dn := s.d.GetNode(nodeid)
	if dn == nil {
		simplelogger.Errorf("node not found: %d", nodeid)
		return
	}
	s.d.SetNodePos(nodeid, x, y)
}

// DeleteNode deletes a node from the simulation.
func (s *Simulation) DeleteNode(nodeid NodeId) error {
	node := s.nodes[nodeid]
	if node == nil {
		simplelogger.Errorf("delete node not found: %d", nodeid)
		return errors.Errorf("node not found")
	}

	_ = node.Exit()
	delete(s.nodes, nodeid)
	s.d.DeleteNode(nodeid)
	return nil
}

// SetNodeFailed sets if the node radio is failed.
func (s *Simulation) SetNodeFailed(id NodeId, failed bool) {
	s.d.SetNodeFailed(id, failed)
}

// ShowDemoLegend shows a demo legend in the visualization.
// It is not implemented yet.
func (s *Simulation) ShowDemoLegend(x int, y int, title string) {
	s.vis.ShowDemoLegend(x, y, title)
}

// SetSpeed sets the simulating speed.
func (s *Simulation) SetSpeed(speed float64) {
	s.d.SetSpeed(speed)
}

// GetSpeed gets the simulating speed.
func (s *Simulation) GetSpeed() float64 {
	return s.Dispatcher().GetSpeed()
}

// CountDown shows a count down in the visualization.
func (s *Simulation) CountDown(duration time.Duration, text string) {
	s.vis.CountDown(duration, text)
}

// Go continues the simulation for a given time duration in simulating time.
// It returns a channel for notifying that the duration is elapsed in simulation.
func (s *Simulation) Go(duration time.Duration) <-chan struct{} {
	return s.d.Go(duration)
}

func (s *Simulation) removeTmpDir() error {
	// tmp directory is used by nodes for saving *.flash files. Need to be removed when simulation started
	return os.RemoveAll("tmp")
}
