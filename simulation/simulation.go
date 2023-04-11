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

package simulation

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"time"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/energy"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/radiomodel"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

type Simulation struct {
	ctx            *progctx.ProgCtx
	err            error
	cfg            *Config
	nodes          map[NodeId]*Node
	d              *dispatcher.Dispatcher
	vis            visualize.Visualizer
	cmdRunner      CmdRunner
	rawMode        bool
	networkInfo    visualize.NetworkInfo
	energyAnalyser *energy.EnergyAnalyser
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
	dispatcherCfg.DumpPackets = cfg.DumpPackets

	s.d = dispatcher.NewDispatcher(s.ctx, dispatcherCfg, s)
	s.d.SetRadioModel(radiomodel.Create(cfg.RadioModel))
	s.vis = s.d.GetVisualizer()
	if err := s.createTmpDir(); err != nil {
		simplelogger.Panicf("creating ./tmp/ directory failed: %+v", err)
	}
	if err := s.cleanTmpDir(cfg.Id); err != nil {
		simplelogger.Panicf("cleaning ./tmp/ directory files '%d_*.*' failed: %+v", cfg.Id, err)
	}

	//TODO add a flag to turn on/off the energy analyzer
	s.energyAnalyser = energy.NewEnergyAnalyser()
	s.d.SetEnergyAnalyser(s.energyAnalyser)
	s.vis.SetEnergyAnalyser(s.energyAnalyser)

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

	// creation of the sim/dispatcher nodes
	simplelogger.Infof("simulation:AddNode: %+v, rawMode=%v", cfg, s.rawMode)
	s.d.AddNode(nodeid, cfg) // ensure dispatcher-node is present before OT process starts.
	node, err := newNode(s, nodeid, cfg)
	if err != nil {
		simplelogger.Errorf("simulation add node failed: %v", err)
		s.d.DeleteNode(nodeid)
		return nil, err
	}
	s.nodes[nodeid] = node

	// init of the sim/dispatcher nodes
	node.uartType = NodeUartTypeVirtualTime
	s.d.RecvEvents() // allow new node to connect, and to receive its startup events.
	if s.d.IsAlive(nodeid) {
		simplelogger.Fatalf("simulation AddNode: new node %d did not respond", nodeid)
	}
	node.setupMode()
	if !s.rawMode {
		initScript := cfg.InitScript
		if len(initScript) == 0 {
			initScript = s.cfg.InitScript // use default, if indicated so in cfg.
		}
		node.SetupNetworkParameters(initScript)
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

// Returns the last error that occurred in the simulation run, or nil if none.
func (s *Simulation) Error() error {
	return s.err
}

func (s *Simulation) Nodes() map[NodeId]*Node {
	return s.nodes
}

func (s *Simulation) GetNodes() []NodeId {
	keys := make([]NodeId, len(s.nodes))
	i := 0
	for key := range s.nodes {
		keys[i] = key
		i++
	}
	return keys
}

func (s *Simulation) AutoGo() bool {
	return s.cfg.AutoGo
}

func (s *Simulation) Stop() {
	if s.IsStopped() {
		return
	}

	simplelogger.Infof("stopping simulation and exiting nodes ...")
	for _, node := range s.nodes {
		_ = node.Exit()
	}
	s.nodes = nil
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

func (s *Simulation) OnNodeProcessFailure(node *Node) {
	if s.ctx.Err() != nil {
		// ignore any node errors when simulation is done.
		simplelogger.Debugf("While exiting simulation, Node %v process failed: %s", node.Id, node.err)
		return
	}
	s.err = node.err
	simplelogger.Errorf("Node %v process failed: %s", node.Id, node.err)
	s.PostAsync(false, func() {
		simplelogger.Infof("Deleting node %v due to process failure.", node.Id)
		_ = s.DeleteNode(node.Id)
	})
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
	return s.d.GetSpeed()
}

func (s *Simulation) CountDown(duration time.Duration, text string) {
	s.vis.CountDown(duration, text)
}

// Run simulation for duration at Dispatcher's set speed.
func (s *Simulation) Go(duration time.Duration) <-chan struct{} {
	return s.d.Go(duration)
}

// Stop any ongoing (previous) 'go' period and then run simulation for duration at given speed.
func (s *Simulation) GoAtSpeed(duration time.Duration, speed float64) <-chan struct{} {
	simplelogger.AssertTrue(speed > 0)
	_ = s.d.GoCancel()
	return s.d.GoAtSpeed(duration, speed)
}

func (s *Simulation) cleanTmpDir(simulationId int) error {
	// tmp directory is used by nodes for saving *.flash files. Need to be cleaned when simulation started
	err := RemoveAllFiles(fmt.Sprintf("tmp/%d_*.flash", simulationId))
	if err != nil {
		return err
	}
	err = RemoveAllFiles(fmt.Sprintf("tmp/%d_*.log", simulationId))
	return err
}

func (s *Simulation) createTmpDir() error {
	// tmp directory is used by nodes for saving *.flash files. Need to be present when simulation started
	err := os.Mkdir("tmp", 0775)
	if errors.Is(err, fs.ErrExist) {
		return nil // ok, already present
	}
	return err
}

// IsStopped returns if the simulation is already stopped.
func (s *Simulation) IsStopped() bool {
	return s.nodes == nil
}

func (s *Simulation) SetTitleInfo(titleInfo visualize.TitleInfo) {
	s.vis.SetTitle(titleInfo)
	s.energyAnalyser.SetTitle(titleInfo.Title)
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

func (s *Simulation) GetEnergyAnalyser() *energy.EnergyAnalyser {
	return s.energyAnalyser
}
