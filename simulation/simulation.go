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
	"os"
	"sort"
	"time"

	"github.com/openthread/ot-ns/progctx"

	"github.com/openthread/ot-ns/dispatcher"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

type Simulation struct {
	ctx         *progctx.ProgCtx
	cfg         *Config
	nodes       map[NodeId]*Node
	d           *dispatcher.Dispatcher
	vis         visualize.Visualizer
	cmdRunner   CmdRunner
	rawMode     bool
	networkInfo visualize.NetworkInfo
}

func NewSimulation(ctx *progctx.ProgCtx, cfg *Config, dispatcherCfg *dispatcher.Config) (*Simulation, error) {
	s := &Simulation{
		ctx:         ctx,
		cfg:         cfg,
		nodes:       map[NodeId]*Node{},
		rawMode:     cfg.RawMode,
		networkInfo: visualize.DefaultNetworkInfo(),
	}
	s.networkInfo.Real = cfg.Real

	// start the event_dispatcher for virtual time
	if dispatcherCfg == nil {
		dispatcherCfg = dispatcher.DefaultConfig()
	}

	dispatcherCfg.Speed = cfg.Speed
	dispatcherCfg.Real = cfg.Real
	dispatcherCfg.Host = cfg.DispatcherHost
	dispatcherCfg.Port = cfg.DispatcherPort
	dispatcherCfg.DumpPackets = cfg.DumpPackets
	dispatcherCfg.UnitDistance = cfg.UnitDistance

	s.d = dispatcher.NewDispatcher(s.ctx, dispatcherCfg, s)
	s.vis = s.d.GetVisualizer()
	if err := s.removeTmpDir(); err != nil {
		simplelogger.Panicf("remove tmp directory failed: %+v", err)
	}

	return s, nil
}

func (s *Simulation) AddNode(cfg *NodeConfig) (*Node, error) {
	if cfg == nil {
		cfg = DefaultNodeConfig()
	}

	nodeid := cfg.ID
	if nodeid <= 0 {
		nodeid = s.genNodeId()
	}

	if s.nodes[nodeid] != nil {
		return nil, errors.Errorf("node %d already exists", nodeid)
	}

	node, err := newNode(s, nodeid, cfg)
	if err != nil {
		simplelogger.Errorf("simulation add node failed: %v", err)
		return nil, err
	}

	s.nodes[nodeid] = node

	simplelogger.Infof("simulation:CtrlAddNode: %+v, rawMode=%v", cfg, s.rawMode)
	s.d.AddNode(nodeid, cfg)

	node.detectVirtualTimeUART()

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

func (s *Simulation) Run() {
	s.ctx.WaitAdd("simulation", 1)
	defer s.ctx.WaitDone("simulation")
	defer simplelogger.Debugf("simulation exit.")

	defer s.Stop()

	s.d.Run()
}

func (s *Simulation) Nodes() map[NodeId]*Node {
	return s.nodes
}

func (s *Simulation) NetworkKey() string {
	return s.cfg.NetworkKey
}

func (s *Simulation) Panid() uint16 {
	return s.cfg.Panid
}

func (s *Simulation) Channel() int {
	return s.cfg.Channel
}

func (s *Simulation) Stop() {
	if s.IsStopped() {
		return
	}

	simplelogger.Infof("stopping simulation ...")
	for _, node := range s.nodes {
		_ = node.Exit()
	}

	s.nodes = nil

	s.d.Stop()
}

func (s *Simulation) SetVisualizer(vis visualize.Visualizer) {
	simplelogger.AssertNotNil(vis)
	s.vis = vis
	s.d.SetVisualizer(vis)
	vis.SetController(NewSimulationController(s))

	s.vis.SetNetworkInfo(s.GetNetworkInfo())
}

func (s *Simulation) OnNodeFail(nodeid NodeId) {
	node := s.nodes[nodeid]
	simplelogger.AssertNotNil(node)
}

func (s *Simulation) OnNodeRecover(nodeid NodeId) {
	node := s.nodes[nodeid]
	simplelogger.AssertNotNil(node)
}

// OnUartWrite notifies the simulation that a node has received some data from UART.
// It is part of implementation of dispatcher.CallbackHandler.
func (s *Simulation) OnUartWrite(nodeid NodeId, data []byte) {
	node := s.nodes[nodeid]
	if node == nil {
		return
	}

	node.onUartWrite(data)
}

func (s *Simulation) PostAsync(trivial bool, f func()) {
	s.d.PostAsync(trivial, f)
}

func (s *Simulation) Dispatcher() *dispatcher.Dispatcher {
	return s.d
}

func (s *Simulation) VisitNodesInOrder(cb func(node *Node)) {
	var nodeids []NodeId
	for nodeid := range s.nodes {
		nodeids = append(nodeids, nodeid)
	}
	sort.Ints(nodeids)
	for _, nodeid := range nodeids {
		cb(s.nodes[nodeid])
	}
}

func (s *Simulation) MoveNodeTo(nodeid NodeId, x, y int) {
	dn := s.d.GetNode(nodeid)
	if dn == nil {
		simplelogger.Errorf("node not found: %d", nodeid)
		return
	}
	s.d.SetNodePos(nodeid, x, y)
}

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

func (s *Simulation) SetNodeFailed(id NodeId, failed bool) {
	s.d.SetNodeFailed(id, failed)
}

func (s *Simulation) ShowDemoLegend(x int, y int, title string) {
	s.vis.ShowDemoLegend(x, y, title)
}

func (s *Simulation) SetSpeed(speed float64) {
	s.d.SetSpeed(speed)
}

func (s *Simulation) GetSpeed() float64 {
	return s.Dispatcher().GetSpeed()
}

func (s *Simulation) CountDown(duration time.Duration, text string) {
	s.vis.CountDown(duration, text)
}

func (s *Simulation) Go(duration time.Duration) <-chan struct{} {
	return s.d.Go(duration)
}

func (s *Simulation) removeTmpDir() error {
	// tmp directory is used by nodes for saving *.flash files. Need to be removed when simulation started
	return os.RemoveAll("tmp")
}

// IsStopped returns if the simulation is already stopped.
func (s *Simulation) IsStopped() bool {
	return s.nodes == nil
}

func (s *Simulation) SetTitleInfo(titleInfo visualize.TitleInfo) {
	s.vis.SetTitle(titleInfo)
}

func (s *Simulation) SetCmdRunner(cmdRunner CmdRunner) {
	simplelogger.AssertTrue(s.cmdRunner == nil)
	s.cmdRunner = cmdRunner
}

func (s *Simulation) GetNetworkInfo() visualize.NetworkInfo {
	return s.networkInfo
}

func (s *Simulation) SetNetworkInfo(networkInfo visualize.NetworkInfo) {
	s.networkInfo = networkInfo
	s.vis.SetNetworkInfo(networkInfo)
}
