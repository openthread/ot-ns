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
	Started        chan struct{}
	ctx            *progctx.ProgCtx
	stopped        bool
	cfg            *Config
	nodes          map[NodeId]*Node
	d              *dispatcher.Dispatcher
	vis            visualize.Visualizer
	cmdRunner      CmdRunner
	rawMode        bool
	networkInfo    visualize.NetworkInfo
	energyAnalyser *energy.EnergyAnalyser
	nodePlacer     *NodeAutoPlacer
	logLevel       WatchLogLevel
}

func NewSimulation(ctx *progctx.ProgCtx, cfg *Config, dispatcherCfg *dispatcher.Config) (*Simulation, error) {
	s := &Simulation{
		Started:     make(chan struct{}),
		ctx:         ctx,
		cfg:         cfg,
		nodes:       map[NodeId]*Node{},
		rawMode:     cfg.RawMode,
		networkInfo: visualize.DefaultNetworkInfo(),
		nodePlacer:  NewNodeAutoPlacer(),
	}
	s.SetLogLevel(cfg.LogLevel)
	s.networkInfo.Real = cfg.Real

	// start the dispatcher for virtual time
	if dispatcherCfg == nil {
		dispatcherCfg = dispatcher.DefaultConfig()
	}

	dispatcherCfg.Speed = cfg.Speed
	dispatcherCfg.Real = cfg.Real
	dispatcherCfg.DumpPackets = cfg.DumpPackets

	s.d = dispatcher.NewDispatcher(s.ctx, dispatcherCfg, s)
	s.d.SetRadioModel(radiomodel.NewRadioModel(cfg.RadioModel))
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
	nodeid := cfg.ID
	if nodeid <= 0 {
		nodeid = s.genNodeId()
	}

	if s.nodes[nodeid] != nil {
		return nil, errors.Errorf("node %d already exists", nodeid)
	}

	// node position may use the nodePlacer
	if cfg.IsAutoPlaced {
		cfg.X, cfg.Y = s.nodePlacer.NextNodePosition(cfg.IsMtd || !cfg.IsRouter)
	} else {
		s.nodePlacer.UpdateReference(cfg.X, cfg.Y)
	}

	// auto-selection of Executable by simulation's policy, in case not defined by cfg.
	if len(cfg.ExecutablePath) == 0 {
		cfg.ExecutablePath = s.cfg.ExeConfig.DetermineExecutableBasedOnConfig(cfg)
	}

	// creation of the dispatcher and simulation nodes
	simplelogger.Debugf("simulation:AddNode: %+v, rawMode=%v", cfg, s.rawMode)
	dnode := s.d.AddNode(nodeid, cfg) // ensure dispatcher-node is present before OT process starts.
	node, err := newNode(s, nodeid, cfg)
	if err != nil {
		simplelogger.Errorf("simulation add node failed: %v", err)
		s.d.DeleteNode(nodeid) // delete dispatcher node again.
		s.nodePlacer.ReuseNextNodePosition()
		return nil, err
	}
	s.nodes[nodeid] = node

	// init of the sim/dispatcher nodes
	node.uartType = NodeUartTypeVirtualTime
	simplelogger.AssertTrue(s.d.IsAlive(nodeid))
	evtCnt := s.d.RecvEvents() // allow new node to connect, and to receive its startup events.
	ts := s.d.CurTime

	node.DisplayPendingLogEntries(ts)
	if s.ctx.Err() != nil { // stop early when exiting the simulation.
		return nil, CommandInterruptedError
	}

	simplelogger.AssertFalse(s.d.IsAlive(nodeid))
	if !dnode.IsConnected() {
		_ = s.DeleteNode(nodeid)
		s.nodePlacer.ReuseNextNodePosition()
		node.DisplayPendingLogEntries(ts)
		return nil, errors.Errorf("simulation AddNode: new node %d did not respond (evtCnt=%d)", nodeid, evtCnt)
	}
	simplelogger.Debugf("start setup of new node (mode, init script)")
	node.setupMode()
	err = node.CommandResult()

	if !s.rawMode && err == nil {
		err = node.runInitScript(cfg.InitScript)
	}

	node.DisplayPendingLogEntries(ts)
	if s.ctx.Err() != nil { // stop early when exiting the simulation.
		return nil, CommandInterruptedError
	}

	if err != nil {
		node.logError(fmt.Errorf("simulation node init failed, deleting node - %v", err))
		_ = s.DeleteNode(node.Id)
		s.nodePlacer.ReuseNextNodePosition()
		node.DisplayPendingLogEntries(ts)
		return nil, err
	}

	node.onStart()
	node.DisplayPendingLogEntries(ts)
	return node, err
}

func (s *Simulation) genNodeId() NodeId {
	nodeid := 1
	for s.nodes[nodeid] != nil {
		nodeid += 1
	}
	return nodeid
}

func (s *Simulation) Run() {
	defer simplelogger.Debugf("simulation exit.")
	defer s.d.Stop()
	defer s.Stop()

	// run dispatcher in current thread, until exit.
	s.ctx.WaitAdd("dispatcher", 1)
	close(s.Started)
	s.d.Run()
}

func (s *Simulation) Nodes() map[NodeId]*Node {
	return s.nodes
}

// GetNodes returns a sorted array of NodeIds.
func (s *Simulation) GetNodes() []NodeId {
	keys := make([]NodeId, len(s.nodes))
	i := 0
	for key := range s.nodes {
		keys[i] = key
		i++
	}
	sort.Ints(keys)
	return keys
}

func (s *Simulation) AutoGo() bool {
	return s.cfg.AutoGo
}

func (s *Simulation) Stop() {
	if s.stopped {
		return
	}

	simplelogger.Infof("stopping simulation and exiting nodes ...")
	s.stopped = true

	// for faster process, signal node exit first in parallel.
	for _, node := range s.nodes {
		_ = node.SignalExit()
	}

	// then clean up and wait for each node process to stop, sequentially.
	s.ctx.Cancel("simulation-stop")
	for _, node := range s.nodes {
		_ = node.Exit()
	}

	simplelogger.Debugf("all simulation nodes exited.")
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
	node.uartReader <- data
}

// OnLogMessage notifies the simulation of a new node-related log message from Dispatcher.
func (s *Simulation) OnLogMessage(logEntry LogEntry) {
	node := s.nodes[logEntry.NodeId]
	if node == nil {
		PrintLog(logEntry.Level, fmt.Sprintf("(Unknown Node %d) %s", logEntry.NodeId, logEntry.Msg))
		return
	}
	node.logEntries <- logEntry
}

func (s *Simulation) OnNextEventTime(ts uint64, nextTs uint64) {
	// display the pending log messages of nodes. Nodes are sorted by id.
	s.VisitNodesInOrder(func(node *Node) {
		node.processUartData()
		node.DisplayPendingLogEntries(ts)
		node.DisplayPendingLines(ts)
	})
	s.VisitNodesInOrder(func(node *Node) {
		simplelogger.AssertEqual(0, len(node.logEntries))
	})
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

func (s *Simulation) MoveNodeTo(nodeid NodeId, x, y int) error {
	dn := s.d.GetNode(nodeid)
	if dn == nil {
		err := fmt.Errorf("node not found: %d", nodeid)
		return err
	}
	s.d.SetNodePos(nodeid, x, y)
	s.nodePlacer.UpdateReference(x, y)
	return nil
}

func (s *Simulation) DeleteNode(nodeid NodeId) error {
	node := s.nodes[nodeid]
	if node == nil {
		err := fmt.Errorf("node not found: %d", nodeid)
		return err
	}
	simplelogger.AssertFalse(s.Dispatcher().IsAlive(nodeid))
	s.d.NotifyCommand(nodeid) // sets node alive, as we expect a NodeExit event as final one in queue.
	err := node.Exit()
	s.d.RecvEvents()
	s.d.DeleteNode(nodeid)
	node.DisplayPendingLogEntries(s.d.CurTime)
	delete(s.nodes, nodeid)
	return err
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

// Go runs the simulation for duration at Dispatcher's set speed.
func (s *Simulation) Go(duration time.Duration) <-chan error {
	return s.d.Go(duration)
}

// GoAtSpeed stops any ongoing (previous) 'go' period and then runs simulation for duration at given speed.
func (s *Simulation) GoAtSpeed(duration time.Duration, speed float64) <-chan error {
	simplelogger.AssertTrue(speed > 0)
	s.d.GoCancel()
	return s.d.GoAtSpeed(duration, speed)
}

func (s *Simulation) cleanTmpDir(simulationId int) error {
	// tmp directory is used by nodes for saving *.flash files. Need to be cleaned when simulation started
	err := removeAllFiles(fmt.Sprintf("tmp/%d_*.flash", simulationId))
	if err != nil {
		return err
	}
	err = removeAllFiles(fmt.Sprintf("tmp/%d_*.log", simulationId))
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

func (s *Simulation) GetConfig() *Config {
	return s.cfg
}

func (s *Simulation) GetLogLevel() WatchLogLevel {
	return s.logLevel
}

func (s *Simulation) SetLogLevel(level WatchLogLevel) {
	s.logLevel = level
	simplelogger.SetLevel(GetSimpleloggerLevel(level))
}
