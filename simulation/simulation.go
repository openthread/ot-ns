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

package simulation

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"time"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/energy"
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/radiomodel"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
)

type Simulation struct {
	Started chan struct{}
	Exited  chan struct{}

	ctx            *progctx.ProgCtx
	stopped        bool
	cfg            *Config
	nodes          map[NodeId]*Node
	d              *dispatcher.Dispatcher
	vis            visualize.Visualizer
	cmdRunner      CmdRunner
	rawMode        bool
	autoGo         bool
	autoGoChange   chan bool
	networkInfo    visualize.NetworkInfo
	energyAnalyser *energy.EnergyAnalyser
	nodePlacer     *NodeAutoPlacer
}

func NewSimulation(ctx *progctx.ProgCtx, cfg *Config, dispatcherCfg *dispatcher.Config) (*Simulation, error) {
	s := &Simulation{
		Started:      make(chan struct{}),
		Exited:       make(chan struct{}),
		ctx:          ctx,
		cfg:          cfg,
		nodes:        map[NodeId]*Node{},
		rawMode:      cfg.RawMode,
		autoGo:       cfg.AutoGo,
		autoGoChange: make(chan bool, 1),
		networkInfo:  visualize.DefaultNetworkInfo(),
		nodePlacer:   NewNodeAutoPlacer(),
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
		logger.Panicf("creating ./tmp/ directory failed: %+v", err)
	}
	if err := s.cleanTmpDir(cfg.Id); err != nil {
		logger.Panicf("cleaning ./tmp/ directory files '%d_*.*' failed: %+v", cfg.Id, err)
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
		cfg.ID = nodeid
	}

	if s.nodes[nodeid] != nil {
		return nil, errors.Errorf("node %d already exists", nodeid)
	}

	// node position may use the nodePlacer
	if cfg.IsAutoPlaced {
		cfg.X, cfg.Y, cfg.Z = s.nodePlacer.NextNodePosition(cfg.IsMtd || !cfg.IsRouter)
	} else {
		s.nodePlacer.UpdateReference(cfg.X, cfg.Y, cfg.Z)
	}

	// selection of Executable by simulation's policy, based on cfg
	if len(cfg.Version) > 0 {
		cfg.ExecutablePath = s.cfg.ExeConfigDefault.FindExecutableBasedOnConfig(cfg)
	} else {
		cfg.ExecutablePath = s.cfg.ExeConfig.FindExecutableBasedOnConfig(cfg)
	}

	// creation of the dispatcher and simulation nodes
	logger.Debugf("simulation:AddNode: %+v, rawMode=%v", cfg, s.rawMode)
	dnode := s.d.AddNode(nodeid, cfg)
	node, err := newNode(s, nodeid, cfg, dnode)
	if err != nil {
		logger.Errorf("simulation add node failed: %v", err)
		_ = node.exit()        // delete the simulation node
		s.d.DeleteNode(nodeid) // delete dispatcher node again.
		s.nodePlacer.ReuseNextNodePosition()
		return nil, err
	}
	s.nodes[nodeid] = node

	// init of the sim/dispatcher nodes
	node.uartType = NodeUartTypeVirtualTime
	logger.AssertTrue(s.d.IsAlive(nodeid))
	evtCnt := s.d.RecvEvents() // allow new node to connect, and to receive its startup events.

	if s.IsStopping() { // stop early when exiting the simulation.
		_ = s.DeleteNode(nodeid)
		return nil, CommandInterruptedError
	}

	if !dnode.IsConnected() {
		_ = s.DeleteNode(nodeid)
		s.nodePlacer.ReuseNextNodePosition()
		return nil, errors.Errorf("simulation AddNode: new node %d did not respond (evtCnt=%d)", nodeid, evtCnt)
	}
	logger.AssertFalse(s.d.IsAlive(nodeid))

	// run setup and script(s) for the node
	node.Logger.Debugf("start setup of node (version/commit, mode, init script)")
	ver := node.GetVersion()
	threadVer := node.GetThreadVersion()
	nodeInfo := visualize.NetworkInfo{
		Real:          s.networkInfo.Real,
		Version:       ver,
		Commit:        getCommitFromOtVersion(ver),
		NodeId:        nodeid,
		ThreadVersion: threadVer,
	}
	s.vis.SetNetworkInfo(nodeInfo)

	err = node.CommandResult()
	if err == nil {
		node.setupMode()
		err = node.CommandResult()
	}
	if !s.rawMode && err == nil {
		err = node.runScript(cfg.InitScript)
	}

	if s.IsStopping() { // stop here when exiting the simulation.
		_ = s.DeleteNode(nodeid)
		return nil, CommandInterruptedError
	}

	if err != nil {
		node.Logger.Errorf("simulation node init failed, deleting node - %v", err)
		_ = s.DeleteNode(node.Id)
		s.nodePlacer.ReuseNextNodePosition()
		return nil, err
	}

	node.onStart()
	node.DisplayPendingLogEntries()
	node.DisplayPendingLines()

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
	defer logger.Debugf("simulation exit.")

	if s.cfg.AutoGo {
		s.autoGoChange <- true
	}

	// run dispatcher in current thread, until exit.
	s.ctx.WaitAdd("dispatcher", 1)
	close(s.Started)
	s.d.Run()
	s.ctx.Cancel("simulation-run")
	s.Stop()
	close(s.Exited)
	s.d.Stop()
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
	return s.autoGo
}

func (s *Simulation) SetAutoGo(isAuto bool) {
	if s.autoGo != isAuto {
		s.autoGoChange <- isAuto
		s.autoGo = isAuto
	}
}

func (s *Simulation) AutoGoRoutine(ctx *progctx.ProgCtx, sim *Simulation) {
	defer ctx.WaitDone("autogo")

	for {
	loop2:
		// First for block waits until autogo is enabled.
		for {
			select {
			case isAutoGo := <-sim.autoGoChange:
				if isAutoGo {
					break loop2
				}
			case <-ctx.Done():
				return
			}
		}

	loop:
		// Second for block executes Go() until autogo is disabled.
		for {
			select {
			case isAutoGo := <-sim.autoGoChange:
				if !isAutoGo {
					break loop
				}
			case <-ctx.Done():
				return
			default:
				<-sim.Go(time.Second)
			}
		}
	}
}

func (s *Simulation) IsStopping() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

func (s *Simulation) Stop() {
	if s.stopped {
		return
	}

	logger.Infof("stopping simulation and exiting nodes ...")
	s.stopped = true

	// for faster process, signal node exit first in parallel.
	for _, node := range s.nodes {
		_ = node.signalExit()
	}

	// then clean up and wait for each node process to stop, sequentially.
	s.ctx.Cancel("simulation-stop")
	for _, node := range s.nodes {
		_ = node.exit()
	}

	logger.Debugf("all simulation nodes exited.")
}

func (s *Simulation) SetVisualizer(vis visualize.Visualizer) {
	logger.AssertNotNil(vis)
	s.vis = vis
	s.d.SetVisualizer(vis)
	vis.SetController(NewSimulationController(s))

	s.vis.SetNetworkInfo(s.GetNetworkInfo())
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

func (s *Simulation) OnNextEventTime(nextTs uint64) {
	// display the pending log messages of nodes. Nodes are sorted by id.
	s.VisitNodesInOrder(func(node *Node) {
		node.processUartData()
		node.DisplayPendingLogEntries()
		node.DisplayPendingLines()
	})
}

func (s *Simulation) OnRfSimEvent(nodeid NodeId, evt *event.Event) {
	node := s.nodes[nodeid]
	if node == nil {
		return
	}

	switch evt.Type {
	case event.EventTypeRadioRfSimParamRsp:
		node.pendingEvents <- evt
	default:
		break
	}
}

// PostAsync will post an asynchronous simulation task in the queue for execution
// @return true when post was successful, false if not (e.g. when sim exited)
func (s *Simulation) PostAsync(f func()) bool {
	select {
	case <-s.Exited:
		return false
	case <-s.ctx.Done():
		return false
	default:
		s.d.PostAsync(f)
		return true
	}
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

func (s *Simulation) MoveNodeTo(nodeid NodeId, x, y int, z *int) error {
	dn := s.d.GetNode(nodeid)
	if dn == nil {
		err := fmt.Errorf("node %d not found", nodeid)
		return err
	}
	zNew := dn.Z
	if z != nil { // only adapt z-coordinate if provided (!= nil)
		zNew = *z
	}
	s.d.SetNodePos(nodeid, x, y, zNew)
	s.nodePlacer.UpdateReference(x, y, zNew)
	return nil
}

func (s *Simulation) DeleteNode(nodeid NodeId) error {
	node := s.nodes[nodeid]
	if node == nil {
		err := fmt.Errorf("node %d not found", nodeid)
		return err
	}
	s.d.NotifyCommand(nodeid) // sets node alive, as we expect a NodeExit event as final one in queue.
	err := node.exit()
	s.d.RecvEvents()
	s.d.DeleteNode(nodeid)
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
	logger.AssertTrue(speed > 0)
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
	logger.AssertTrue(s.cmdRunner == nil)
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

func (s *Simulation) GetLogLevel() logger.Level {
	return logger.GetLevel()
}

func (s *Simulation) SetLogLevel(level logger.Level) {
	logger.SetLevel(level)
}
