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

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/radiomodel"
	"github.com/openthread/ot-ns/simulation"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/openthread/ot-ns/web"
)

const (
	Prompt = "> "
)

type CommandContext struct {
	context.Context
	*Command
	rt              *CmdRunner
	err             error
	output          io.Writer
	isBackgroundCmd bool
}

func (cc *CommandContext) outputStr(msg string) {
	_, _ = fmt.Fprint(cc.output, msg)
}

func (cc *CommandContext) outputStrArray(msg []string) {
	for _, line := range msg {
		_, _ = fmt.Fprint(cc.output, line+"\n")
	}
}

func (cc *CommandContext) outputErr(err error) {
	s := err.Error()
	if strings.HasPrefix(s, "Error") { // OT CLI errors already prepend this.
		cc.outputf("%s\n", s)
	} else {
		cc.outputf("Error: %s\n", s) // OTNS generated errors not.
	}
}

func (cc *CommandContext) outputf(format string, args ...interface{}) {
	_, _ = fmt.Fprintf(cc.output, format, args...)
}

func (cc *CommandContext) errorf(format string, args ...interface{}) {
	cc.error(errors.Errorf(format, args...))
}

func (cc *CommandContext) error(err error) {
	if err != nil {
		if cc.Err() != nil { // if previous error stored, print it now and keep the last.
			cc.outputErr(cc.Err())
		}
		cc.err = err
	}
}

// Err returns the last error that occurred during command execution.
func (cc *CommandContext) Err() error {
	return cc.err
}

func (cc *CommandContext) outputItemsAsYaml(items interface{}) {
	var itemsYaml yaml.Node

	err := itemsYaml.Encode(items)
	logger.PanicIfError(err)

	for _, content := range itemsYaml.Content {
		content.Style = yaml.FlowStyle
	}

	data, err := yaml.Marshal(&itemsYaml)
	logger.PanicIfError(err)

	_, err = cc.output.Write(data)
	logger.PanicIfError(err)
}

type CmdRunner struct {
	sim           *simulation.Simulation
	ctx           *progctx.ProgCtx
	contextNodeId NodeId
	help          Help
}

func NewCmdRunner(ctx *progctx.ProgCtx, sim *simulation.Simulation) *CmdRunner {
	cr := &CmdRunner{
		ctx:           ctx,
		sim:           sim,
		contextNodeId: InvalidNodeId,
		help:          newHelp(),
	}
	sim.SetCmdRunner(cr)
	return cr
}

func (rt *CmdRunner) RunCommand(cmdline string, output io.Writer) error {
	if rt.ctx.Err() == nil {
		// if character '!' is used to invoke no-node (global) context, remove it.
		if len(cmdline) > 1 && cmdline[0] == '!' {
			cmdline = cmdline[1:]
		}
		// run the OTNS-CLI command without node context
		cmd := Command{}

		if err := parseBytes([]byte(cmdline), &cmd); err != nil {
			if _, err := fmt.Fprintf(output, "Error: %v\n", err); err != nil {
				return err
			}
		} else {
			rt.execute(&cmd, output)
		}
	}
	return rt.ctx.Err()
}

func (rt *CmdRunner) HandleCommand(cmdline string, output io.Writer) error {
	if rt.ctx.Err() == nil {
		if rt.contextNodeId != InvalidNodeId && !isContextlessCommand(cmdline) {
			// run the command in node context
			cmd := Command{
				Node: &NodeCmd{
					Node:    NodeSelector{Id: rt.contextNodeId},
					Command: &cmdline,
				},
			}
			rt.execute(&cmd, output)
		} else {
			// run the command without node-specific context
			return rt.RunCommand(cmdline, output)
		}
	}
	return rt.ctx.Err()
}

func (rt *CmdRunner) GetPrompt() string {
	if rt.contextNodeId == InvalidNodeId {
		return Prompt
	} else {
		return fmt.Sprintf("node %d%s", rt.contextNodeId, Prompt)
	}
}

func (rt *CmdRunner) execute(cmd *Command, output io.Writer) {
	cc := &CommandContext{
		Command:         cmd,
		rt:              rt,
		output:          output,
		isBackgroundCmd: isBackgroundCommand(cmd),
	}

	defer func() {
		if rt.ctx.Err() != nil && cmd.Exit == nil {
			cc.outputErr(simulation.CommandInterruptedError)
		} else if cc.Err() != nil {
			cc.outputErr(cc.Err())
		} else if !cc.isBackgroundCmd {
			cc.outputf("Done\n")
		}
	}()

	defer func() {
		rerr := recover()
		if rerr != nil {
			if err, ok := rerr.(error); ok {
				cc.err = errors.Wrapf(err, "panic: %v", err)
			} else {
				cc.err = errors.Errorf("panic: %v", rerr)
			}
		}
	}()

	if cmd.Move != nil {
		rt.executeMoveNode(cc, cc.Move)
	} else if cmd.Radio != nil {
		rt.executeRadio(cc, cc.Radio)
	} else if cmd.Go != nil {
		rt.executeGo(cc, cmd.Go)
	} else if cmd.Nodes != nil {
		rt.executeLsNodes(cc, cc.Nodes)
	} else if cmd.Partitions != nil {
		rt.executeLsPartitions(cc)
	} else if cmd.Add != nil {
		rt.executeAddNode(cc, cmd.Add)
	} else if cmd.Del != nil {
		rt.executeDelNode(cc, cmd.Del)
	} else if cmd.Ping != nil {
		rt.executePing(cc, cmd.Ping)
	} else if cmd.Node != nil {
		rt.executeNode(cc, cmd.Node)
	} else if cmd.CountDown != nil {
		rt.executeCountDown(cc, cmd.CountDown)
	} else if cmd.Speed != nil {
		rt.executeSpeed(cc, cmd.Speed)
	} else if cmd.Plr != nil {
		rt.executePlr(cc, cc.Plr)
	} else if cmd.Pings != nil {
		rt.executeCollectPings(cc, cc.Pings)
	} else if cmd.Counters != nil {
		rt.executeCounters(cc, cc.Counters)
	} else if cmd.Joins != nil {
		rt.executeCollectJoins(cc, cc.Joins)
	} else if cmd.Coaps != nil {
		rt.executeCoaps(cc, cc.Coaps)
	} else if cmd.Scan != nil {
		rt.executeScan(cc, cc.Scan)
	} else if cmd.ConfigVisualization != nil {
		rt.executeConfigVisualization(cc, cc.ConfigVisualization)
	} else if cmd.Debug != nil {
		rt.executeDebug(cc, cmd.Debug)
	} else if cmd.Title != nil {
		rt.executeTitle(cc, cmd.Title)
	} else if cmd.DemoLegend != nil {
		rt.executeDemoLegend(cc, cmd.DemoLegend)
	} else if cmd.Exit != nil {
		rt.executeExit(cc, cmd.Exit)
	} else if cmd.Web != nil {
		rt.executeWeb(cc, cc.Web)
	} else if cmd.NetInfo != nil {
		rt.executeNetInfo(cc, cc.NetInfo)
	} else if cmd.RadioModel != nil {
		rt.executeRadioModel(cc, cc.RadioModel)
	} else if cmd.RadioParam != nil {
		rt.executeRadioParam(cc, cc.RadioParam)
	} else if cmd.RfSim != nil {
		rt.executeRfSim(cc, cc.RfSim)
	} else if cmd.Energy != nil {
		rt.executeEnergy(cc, cc.Energy)
	} else if cmd.LogLevel != nil {
		rt.executeLogLevel(cc, cc.LogLevel)
	} else if cmd.Watch != nil {
		rt.executeWatch(cc, cmd.Watch)
	} else if cmd.Unwatch != nil {
		rt.executeUnwatch(cc, cmd.Unwatch)
	} else if cmd.Time != nil {
		rt.executeTime(cc, cmd.Time)
	} else if cmd.Help != nil {
		rt.executeHelp(cc, cmd.Help)
	} else if cmd.Exe != nil {
		rt.executeExe(cc, cmd.Exe)
	} else if cmd.AutoGo != nil {
		rt.executeAutoGo(cc, cmd.AutoGo)
	} else if cmd.Kpi != nil {
		rt.executeKpi(cc, cmd.Kpi)
	} else if cmd.Load != nil {
		rt.executeLoad(cc, cmd.Load)
	} else if cmd.Save != nil {
		rt.executeSave(cc, cmd.Save)
	} else {
		logger.Panicf("unimplemented command: %#v", cmd)
	}
}

func (rt *CmdRunner) executeGo(cc *CommandContext, cmd *GoCmd) {
	// determine duration and desired speed of the Go simulation period.
	timeDurToGo, err := time.ParseDuration(cmd.Time)
	if cmd.Ever == nil && err != nil {
		timeDurToGo, err = time.ParseDuration(cmd.Time + "s") // try parsing as seconds
		if err != nil {
			cc.errorf("could not parse time duration: %s", cmd.Time)
			return
		}
	}
	speed := rt.sim.GetSpeed()
	if cmd.Speed != nil {
		speed = *cmd.Speed
	} else if rt.sim.AutoGo() {
		// when in AutoGo mode, 'go' command used to quickly jump time.
		speed = dispatcher.MaxSimulateSpeed
	}
	if speed == 0 { // when paused or silly 'speed' param, assume 'go' is used to quickly jump time.
		speed = dispatcher.MaxSimulateSpeed
	}

	// execute the Go
	var done <-chan error
	if cmd.Ever == nil {
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			done = sim.GoAtSpeed(timeDurToGo, speed)
		})
		cc.err = <-done // block for the simulation period.
	} else {
		for { // run forever but stop if rt.ctx.Err indicates "done"
			rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
				sim.SetSpeed(speed) // permanent speed update
				done = sim.Go(time.Hour)
			})
			cc.err = <-done

			if rt.ctx.Err() != nil || cc.Err() != nil {
				break
			}
		}
	}
}

func (rt *CmdRunner) executeAutoGo(cc *CommandContext, cmd *AutoGoCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		if cmd.Val == nil {
			autoGoInt := 0
			if sim.AutoGo() {
				autoGoInt = 1
			}
			cc.outputf("%d\n", autoGoInt)
		} else {
			sim.SetAutoGo(cmd.Val.Yes != nil)
		}
	})
}

func (rt *CmdRunner) executeSpeed(cc *CommandContext, cmd *SpeedCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		if cmd.Speed == nil && cmd.Max == nil {
			cc.outputf("%v\n", sim.GetSpeed())
		} else if cmd.Max != nil {
			sim.SetSpeed(dispatcher.MaxSimulateSpeed)
		} else {
			sim.SetSpeed(*cmd.Speed)
		}
	})
}

func (rt *CmdRunner) postAsyncWait(cc *CommandContext, f func(sim *simulation.Simulation)) {
	done := make(chan struct{})
	if rt.sim.PostAsync(func() {
		defer close(done) // even if f() fails execution, 'done' should be closed.
		f(rt.sim)         // executing task (later) may set cc.Err() status if error occurs.
	}) {
		<-done // only block-wait if task was accepted.
	} else {
		cc.error(simulation.CommandInterruptedError) // report cc error if not accepted.
	}
}

func (rt *CmdRunner) executeAddNode(cc *CommandContext, cmd *AddCmd) {
	logger.Debugf("Add: %#v", *cmd)
	simCfg := cc.rt.sim.GetConfig()
	cfg := simCfg.NewNodeConfig // copy current new-node config for simulation, and modify it.

	cfg.Type = cmd.Type.Val
	if cmd.X != nil {
		cfg.X = *cmd.X
		cfg.IsAutoPlaced = false
	}
	if cmd.Y != nil {
		cfg.Y = *cmd.Y
		cfg.IsAutoPlaced = false
	}
	if cmd.Z != nil {
		cfg.Z = *cmd.Z
		cfg.IsAutoPlaced = false
	}
	if cmd.Id != nil {
		cfg.ID = cmd.Id.Val
	}
	if cmd.RadioRange != nil {
		cfg.RadioRange = cmd.RadioRange.Val
	}
	cfg.Restore = cmd.Restore != nil
	if cmd.Version != nil {
		cfg.Version = cmd.Version.Val
	}
	if cmd.Executable != nil {
		cfg.ExecutablePath = cmd.Executable.Path
	}

	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		sim.NodeConfigFinalize(&cfg)
		node, err := sim.AddNode(&cfg)
		if err != nil {
			cc.error(err)
			return
		}

		cc.outputf("%d\n", node.Id)
	})
}

func (rt *CmdRunner) executeDelNode(cc *CommandContext, cmd *DelCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		for _, sel := range cmd.Nodes {
			node, _ := rt.getNode(sim, sel)
			if node == nil {
				cc.outputf("Warn: node %d not found, skipping\n", sel.Id)
				continue
			}

			err := sim.DeleteNode(node.Id)
			if err != nil {
				cc.errorf("node %d, %+v", sel.Id, err)
			}
		}
	})
}

func (rt *CmdRunner) executeExit(cc *CommandContext, cmd *ExitCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		if rt.enterNodeContext(InvalidNodeId) {
			return
		}
		sim.Stop()
	})
}

func (rt *CmdRunner) executePing(cc *CommandContext, cmd *PingCmd) {
	logger.Debugf("ping %#v", cmd)
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		src, _ := rt.getNode(sim, cmd.Src)
		if src == nil {
			cc.errorf("src node not found")
			return
		}

		var dstaddr string
		if cmd.Dst != nil {
			dst, _ := rt.getNode(sim, *cmd.Dst)

			if dst == nil {
				cc.errorf("dst node not found")
				return
			}
			dstaddrs := rt.getAddrs(dst, cmd.AddrType)
			if len(dstaddrs) <= 0 {
				cc.errorf("dst addr not found")
				return
			}
			dstaddr = dstaddrs[0]
		} else {
			dstaddr = cmd.DstAddr.Addr
		}

		datasize := 4 // Note: must be at least 4 otherwise OTNS will ignore ping req/resp for stats.
		count := 1
		interval := 10
		hopLimit := 64

		if cmd.DataSize != nil {
			datasize = cmd.DataSize.Val
			if datasize < 4 {
				logger.Warnf("Ping with datasize < 4 is ignored by OT-NS statistics code.")
			}
		}

		if cmd.Count != nil {
			count = cmd.Count.Val
		}

		if cmd.Interval != nil {
			interval = cmd.Interval.Val
		}

		if cmd.HopLimit != nil {
			hopLimit = cmd.HopLimit.Val
		}

		src.Ping(dstaddr, datasize, count, interval, hopLimit)
	})
}

func (rt *CmdRunner) getNode(sim *simulation.Simulation, sel NodeSelector) (*simulation.Node, *dispatcher.Node) {
	if sel.Id > 0 {
		var dnode *dispatcher.Node
		node := sim.Nodes()[sel.Id]
		if node != nil {
			dnode = node.DNode
		}
		return node, dnode
	}

	return nil, nil
}

func (rt *CmdRunner) getAddrs(node *simulation.Node, addrType *AddrTypeFlag) []string {
	if node == nil {
		return nil
	}

	var addrs []string
	if (addrType == nil || addrType.Type == AddrTypeAny) || addrType.Type == AddrTypeMleid {
		addrs = append(addrs, node.GetIpAddrMleid()...)
	}
	if len(addrs) > 0 {
		return addrs
	}

	if (addrType == nil || addrType.Type == AddrTypeAny) || addrType.Type == AddrTypeRloc {
		addrs = append(addrs, node.GetIpAddrRloc()...)
	}
	if len(addrs) > 0 {
		return addrs
	}

	if (addrType == nil || addrType.Type == AddrTypeAny) || addrType.Type == AddrTypeSlaac {
		addrs = append(addrs, node.GetIpAddrSlaac()...)
	}
	if len(addrs) > 0 {
		return addrs
	}

	if (addrType == nil || addrType.Type == AddrTypeAny) || addrType.Type == AddrTypeLinkLocal {
		addrs = append(addrs, node.GetIpAddrLinkLocal()...)
	}

	return addrs
}

func (rt *CmdRunner) executeDebug(cc *CommandContext, cmd *DebugCmd) {
	logger.Infof("debug %#v", *cmd)

	if cmd.Echo != nil {
		cc.outputf("%s\n", *cmd.Echo)
	}

	if cmd.Fail != nil {
		cc.errorf("debug failed")
	}
}

func (rt *CmdRunner) executeNode(cc *CommandContext, cmd *NodeCmd) {
	contextNodeId := InvalidNodeId
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		node, _ := rt.getNode(sim, cmd.Node)
		if node == nil {
			if cmd.Node.Id == 0 && rt.contextNodeId != InvalidNodeId && rt.enterNodeContext(InvalidNodeId) {
				// the 'node 0' command will exit node context, only when inside a node-context.
				return
			}
			cc.errorf("node %d not found", cmd.Node.Id)
			return
		}

		defer func() {
			rerr := recover()
			if rerr != nil {
				cc.errorf("%+v", rerr)
			}
		}()

		if cmd.Command != nil {
			var output []string
			prefix := ""
			if cc.isBackgroundCmd {
				output = node.CommandNoDone(*cmd.Command, simulation.DefaultCommandTimeout)
				prefix = "  "
			} else {
				output = node.Command(*cmd.Command, simulation.DefaultCommandTimeout)
			}

			err := node.CommandResult()
			node.DisplayPendingLogEntries()
			node.DisplayPendingLines()
			if cc.isBackgroundCmd && err == nil {
				cc.outputf("Started\n")
			}
			for _, line := range output {
				cc.outputf("%s%s\n", prefix, line)
			}

			if err != nil {
				cc.error(err)
			}
		} else {
			contextNodeId = node.Id
		}

		if contextNodeId != InvalidNodeId {
			// enter node context
			rt.enterNodeContext(contextNodeId)
		}
	})
}

func (rt *CmdRunner) executeDemoLegend(cc *CommandContext, cmd *DemoLegendCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		sim.ShowDemoLegend(cmd.X, cmd.Y, cmd.Title)
	})
}

func (rt *CmdRunner) executeCountDown(cc *CommandContext, cmd *CountDownCmd) {
	title := "%v"
	if cmd.Text != nil {
		title = *cmd.Text
	}
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		sim.CountDown(time.Duration(cmd.Seconds)*time.Second, title)
	})
}

func (rt *CmdRunner) executeRadio(cc *CommandContext, radio *RadioCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		for _, sel := range radio.Nodes {
			node, dnode := rt.getNode(sim, sel)
			if node == nil {
				cc.errorf("node %d not found", sel.Id)
				continue
			}

			if radio.On != nil {
				sim.SetNodeFailed(node.Id, false)
			} else if radio.Off != nil {
				sim.SetNodeFailed(node.Id, true)
			} else if radio.FailTime != nil {
				if radio.FailTime.FailDuration > 0 && radio.FailTime.FailInterval > radio.FailTime.FailDuration {
					dnode.SetFailTime(dispatcher.FailTime{
						FailDuration: uint64(radio.FailTime.FailDuration * 1000000),
						FailInterval: uint64(radio.FailTime.FailInterval * 1000000),
					})
				} else if radio.FailTime.FailInterval <= radio.FailTime.FailDuration {
					cc.errorf("ft parameter: fail-duration must be < fail-interval")
				} else {
					dnode.SetFailTime(dispatcher.NonFailTime)
				}
			}
		}
	})
}

func (rt *CmdRunner) executeMoveNode(cc *CommandContext, cmd *MoveCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		cc.error(sim.MoveNodeTo(cmd.Target.Id, cmd.X, cmd.Y, cmd.Z))
	})
}

func (rt *CmdRunner) executeLsNodes(cc *CommandContext, cmd *NodesCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		for _, nodeid := range sim.GetNodes() {
			snode := sim.Nodes()[nodeid]
			dnode := sim.Dispatcher().GetNode(nodeid)
			var line strings.Builder
			line.WriteString(fmt.Sprintf("id=%d\ttype=%-6s  extaddr=%016x  rloc16=%04x  x=%-2d\ty=%-3d\tz=%-3d\tstate=%s\tfailed=%v", nodeid, dnode.Type, dnode.ExtAddr, dnode.Rloc16,
				dnode.X, dnode.Y, dnode.Z, dnode.Role, dnode.IsFailed()))
			line.WriteString(fmt.Sprintf("\texe=%s", snode.GetExecutableName()))
			cc.outputf("%s\n", line.String())
		}
	})
}

func (rt *CmdRunner) executeLsPartitions(cc *CommandContext) {
	pars := map[uint32][]NodeId{}

	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		for _, dnode := range sim.Dispatcher().Nodes() {
			parid := dnode.PartitionId
			pars[parid] = append(pars[parid], dnode.Id)
		}
	})

	for parid, nodeids := range pars {
		cc.outputf("partition=%08x\tnodes=", parid)
		for i, nodeid := range nodeids {
			if i > 0 {
				cc.outputf(",")
			}
			cc.outputf("%d", nodeid)
		}
		cc.outputf("\n")
	}
}

func (rt *CmdRunner) executeCollectPings(cc *CommandContext, pings *PingsCmd) {
	allPings := make(map[NodeId][]*dispatcher.PingResult)
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		d := sim.Dispatcher()
		for _, node := range d.Nodes() {
			pings := node.CollectPings()
			if len(pings) > 0 {
				allPings[node.Id] = pings
			}
		}
	})

	for nodeid, pings := range allPings {
		for _, ping := range pings {
			cc.outputf("node=%-4d dst=%-40s datasize=%-3d delay=%.3fms\n", nodeid, ping.Dst, ping.DataSize, float64(ping.Delay)/1000)
		}
	}
}

func (rt *CmdRunner) executeCollectJoins(cc *CommandContext, joins *JoinsCmd) {
	allJoins := make(map[NodeId][]*dispatcher.JoinResult)

	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		d := sim.Dispatcher()
		for _, node := range d.Nodes() {
			joins := node.CollectJoins()
			if len(joins) > 0 {
				allJoins[node.Id] = joins
			}
		}
	})

	for nodeid, joins := range allJoins {
		for _, join := range joins {
			cc.outputf("node=%-4d join=%.3fs session=%.3fs\n", nodeid, float64(join.JoinDuration)/1000000, float64(join.SessionDuration)/1000000)
		}
	}
}

func (rt *CmdRunner) executeCounters(cc *CommandContext, counters *CountersCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		d := sim.Dispatcher()
		countersVal := reflect.ValueOf(d.Counters)
		countersTyp := reflect.TypeOf(d.Counters)
		for i := 0; i < countersVal.NumField(); i++ {
			fname := countersTyp.Field(i).Name
			fval := countersVal.Field(i)
			cc.outputf("%-40s %v\n", fname, fval.Uint())
		}
	})
}

func (rt *CmdRunner) executeWeb(cc *CommandContext, webcmd *WebCmd) {
	tabResource := ""
	if *webcmd.TabName == "" {
		tabResource = web.MainTab
	} else {
		switch *webcmd.TabName {
		case "main":
			tabResource = web.MainTab
		case "stats":
			tabResource = web.StatsTab
		case "energy":
			tabResource = web.EnergyTab
		default:
			cc.errorf("unrecognized web tab identifier: %s", *webcmd.TabName)
			return
		}
	}
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		if err := web.OpenWeb(rt.ctx, tabResource); err != nil {
			cc.error(err)
		}
	})
}

func (rt *CmdRunner) executeRadioModel(cc *CommandContext, cmd *RadioModelCmd) {
	var name string
	if len(cmd.Model) == 0 {
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			name = sim.Dispatcher().GetRadioModel().GetName()
		})
		cc.outputf("%v\n", name)
	} else {
		name = cmd.Model
		ok := false
		var model radiomodel.RadioModel = nil
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			model = radiomodel.NewRadioModel(name)
			ok = model != nil
			if ok {
				sim.Dispatcher().SetRadioModel(model)
			}
		})
		if ok {
			cc.outputf("%v\n", model.GetName())
		} else {
			cc.errorf("radiomodel '%v' is not defined", name)
		}
	}
}

func displayRadioParam(val *reflect.Value) (string, bool) {
	if val.CanFloat() {
		f := val.Float()
		if f == radiomodel.UndefinedDbValue {
			return "undefined", false
		}
		return strconv.FormatFloat(f, 'f', -1, 64), true
	} else if val.Bool() {
		return "1", true
	}
	return "0", true
}

func (rt *CmdRunner) executeRadioParam(cc *CommandContext, cmd *RadioParamCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		rm := sim.Dispatcher().GetRadioModel()
		rp := rm.GetParameters()
		rpVal := reflect.ValueOf(rp).Elem()
		rpTyp := reflect.TypeOf(rp).Elem()

		// variant: radioparam
		if len(cmd.Param) == 0 {
			for i := 0; i < rpVal.NumField(); i++ {
				fname := rpTyp.Field(i).Name
				fval := rpVal.Field(i)
				s, isDefined := displayRadioParam(&fval)
				if isDefined {
					cc.outputf("%-20s %s\n", fname, s)
				}
			}
			return
		}

		// variant: radioparam <param-name>
		// variant: radioparam <param-name> <param-value>
		var fval reflect.Value
		_, ok := rpTyp.FieldByName(cmd.Param)
		if !ok {
			cc.errorf("unknown radiomodel parameter: %s", cmd.Param)
			return
		}

		fval = rpVal.FieldByName(cmd.Param)
		isFloat := fval.CanFloat()
		if cmd.Val == nil { // show value of single parameter
			s, _ := displayRadioParam(&fval)
			cc.outputf("%s\n", s)
			return
		}

		// set new parameter value
		newVal := *cmd.Val
		if cmd.Sign == "-" {
			newVal = -newVal
		}

		isChanged := false
		if !isFloat { // if we're setting a bool parameter, use <=0 -> false ; >0 -> true
			newValBool := newVal > 0
			oldValBool := fval.Bool()
			isChanged = oldValBool != newValBool
			fval.SetBool(newValBool)
		} else {
			oldVal := fval.Float()
			isChanged = oldVal != newVal
			fval.SetFloat(newVal)
		}
		if isChanged {
			rm.OnParametersModified()
		}
	})
}

func (rt *CmdRunner) executeRfSim(cc *CommandContext, cmd *RfSimCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		node, _ := rt.getNode(sim, cmd.Id)
		if node == nil {
			cc.errorf("node not found")
			return
		}

		defer node.DisplayPendingLogEntries()

		// variant: rfsim <nodeid>
		if len(cmd.Param) == 0 {
			for i := 0; i < len(RfSimParamsList); i++ {
				value := node.GetRfSimParam(RfSimParamsList[i])
				unit := RfSimParamUnitsList[i]
				if node.CommandResult() != nil {
					cc.error(node.CommandResult())
					return
				}
				cc.outputf("%-20s %d (%s)\n", RfSimParamNamesList[i], value, unit)
			}
			return
		}

		param := ParseRfSimParam(cmd.Param)
		if param == ParamUnknown {
			cc.errorf("parameter '%s' not found", cmd.Param)
			return
		}

		// variant: rfsim <nodeid> <param>
		if cmd.Val == nil {
			value := node.GetRfSimParam(param)
			cc.outputf("%d\n", value)
			return
		}

		// variant: rfsim <nodeid> <param> <new-value>
		newVal := *cmd.Val
		if cmd.Sign == "-" {
			newVal = -newVal
		}

		node.SetRfSimParam(param, RfSimParamValue(newVal))
		if node.CommandResult() != nil {
			cc.error(node.CommandResult())
		}
	})
}

func (rt *CmdRunner) executeLogLevel(cc *CommandContext, cmd *LogLevelCmd) {
	if cmd.Level == "" {
		cc.outputf("%v\n", logger.GetLevelString(rt.sim.GetLogLevel()))
	} else {
		lev, err := logger.ParseLevelString(cmd.Level)
		if err == nil {
			rt.sim.SetLogLevel(lev)
		} else {
			cc.error(err)
		}
	}
}

func (rt *CmdRunner) executeWatch(cc *CommandContext, cmd *WatchCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		levelStr := ""
		var level = logger.DefaultLevel
		var err error
		if len(cmd.Level) > 0 {
			levelStr = cmd.Level
			level, err = logger.ParseLevelString(levelStr)
			if err != nil {
				cc.error(err)
				return
			}
		}
		nodesToWatch := cmd.Nodes

		if len(cmd.Nodes) == 0 && len(cmd.All) == 0 && len(cmd.Default) == 0 && len(cmd.Level) == 0 {
			// variant: 'watch'
			watchedList := strings.Trim(fmt.Sprintf("%v", sim.Dispatcher().GetWatchingNodes()), "[]")
			cc.outputf("%v\n", watchedList)
			return
		} else if len(cmd.Nodes) == 0 && len(cmd.All) == 0 && len(cmd.Default) > 0 && len(cmd.Level) > 0 {
			// variant: 'watch default <level>'
			sim.Dispatcher().GetConfig().DefaultWatchOn = cmd.Level != logger.OffLevelString && cmd.Level != logger.NoneLevelString
			sim.Dispatcher().GetConfig().DefaultWatchLevel = cmd.Level
			return
		} else if len(cmd.Nodes) == 0 && len(cmd.All) == 0 && len(cmd.Default) > 0 && len(cmd.Level) == 0 {
			// variant: 'watch default'
			watchLevelDefault := logger.DefaultLevelString
			if sim.Dispatcher().GetConfig().DefaultWatchOn {
				watchLevelDefault = sim.Dispatcher().GetConfig().DefaultWatchLevel
			}
			cc.outputf("%s\n", watchLevelDefault)
			return
		} else if len(cmd.Nodes) == 0 && len(cmd.All) > 0 && len(cmd.Default) == 0 {
			// variant: 'watch all [<level>]'
			for nodeid := range sim.Nodes() {
				nodesToWatch = append(nodesToWatch, NodeSelector{Id: nodeid})
			}
		} else if len(cmd.Nodes) > 0 && len(cmd.All) == 0 && len(cmd.Default) == 0 {
			// variant: 'watch <nodeid> [<nodeid> ...] [<level>]'
			// Do nothing here. Will iterate over nodes below.
		} else if len(cmd.Nodes) == 0 && len(cmd.All) == 0 && len(cmd.Default) == 0 && len(cmd.Level) > 0 {
			// variant: 'watch <level>'
			// Do nothing here. <level> was processed above already.
		} else {
			cc.errorf("watch: unsupported combination of command options")
			return
		}

		for _, sel := range nodesToWatch {
			node, _ := rt.getNode(sim, sel)
			if node == nil {
				cc.errorf("node %d not found", sel.Id)
				continue
			}
			sim.Dispatcher().WatchNode(node.Id, level)
		}
	})
}

func (rt *CmdRunner) executeUnwatch(cc *CommandContext, cmd *UnwatchCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		// if no node-number(s) given, unwatch all.
		if len(cmd.Nodes) == 0 {
			for _, n := range sim.Dispatcher().GetWatchingNodes() {
				sim.Dispatcher().UnwatchNode(n)
			}
		} else {
			for _, sel := range cmd.Nodes {
				node, _ := rt.getNode(sim, sel)
				if node == nil {
					cc.outputf("Warn: node %d not found, skipping\n", sel.Id)
					continue
				}
				sim.Dispatcher().UnwatchNode(node.Id)
			}
		}
	})
}

func (rt *CmdRunner) executePlr(cc *CommandContext, cmd *PlrCmd) {
	if cmd.Val == nil {
		// get PLR
		var plr float64

		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			plr = sim.Dispatcher().GetGlobalMessageDropRatio()
		})

		cc.outputf("%v\n", plr)
	} else {
		// set PLR
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			sim.Dispatcher().SetGlobalPacketLossRatio(*cmd.Val)
			*cmd.Val = sim.Dispatcher().GetGlobalMessageDropRatio()
		})
		cc.outputf("%v\n", *cmd.Val)
	}
}

func (rt *CmdRunner) executeScan(cc *CommandContext, cmd *ScanCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		node, _ := rt.getNode(sim, cmd.Node)
		if node == nil {
			cc.errorf("node %d not found", cmd.Node.Id)
			return
		}
		output := node.CommandNoDone("scan", simulation.DefaultCommandTimeout)
		err := node.CommandResult()
		cc.error(err)
		if err == nil {
			cc.outputf("Started\n")
		}
		cc.outputStrArray(output)
	})
}

func (rt *CmdRunner) executeConfigVisualization(cc *CommandContext, cmd *ConfigVisualizationCmd) {
	var opts dispatcher.VisualizationOptions
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		opts = sim.Dispatcher().GetVisualizationOptions()

		if cmd.BroadcastMessage != nil {
			opts.BroadcastMessage = cmd.BroadcastMessage.OnOrOff.On != nil
		}

		if cmd.UnicastMessage != nil {
			opts.UnicastMessage = cmd.UnicastMessage.OnOrOff.On != nil
		}

		if cmd.AckMessage != nil {
			opts.AckMessage = cmd.AckMessage.OnOrOff.On != nil
		}

		if cmd.RouterTable != nil {
			opts.RouterTable = cmd.RouterTable.OnOrOff.On != nil
		}

		if cmd.ChildTable != nil {
			opts.ChildTable = cmd.ChildTable.OnOrOff.On != nil
		}

		sim.Dispatcher().SetVisualizationOptions(opts)
	})

	bool_to_onoroff := func(on bool) string {
		if on {
			return "on"
		} else {
			return "off"
		}
	}
	cc.outputf("bro=%s\n", bool_to_onoroff(opts.BroadcastMessage))
	cc.outputf("uni=%s\n", bool_to_onoroff(opts.UnicastMessage))
	cc.outputf("ack=%s\n", bool_to_onoroff(opts.AckMessage))
	cc.outputf("rtb=%s\n", bool_to_onoroff(opts.RouterTable))
	cc.outputf("ctb=%s\n", bool_to_onoroff(opts.ChildTable))
}

func (rt *CmdRunner) enterNodeContext(nodeid NodeId) bool {
	logger.AssertTrue(nodeid == InvalidNodeId || nodeid > 0)
	if rt.contextNodeId == nodeid {
		return false
	}

	rt.contextNodeId = nodeid
	return true
}

func (rt *CmdRunner) executeTitle(cc *CommandContext, cmd *TitleCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		titleInfo := visualize.DefaultTitleInfo()

		titleInfo.Title = cmd.Title
		if cmd.X != nil {
			titleInfo.X = *cmd.X
		}
		if cmd.Y != nil {
			titleInfo.Y = *cmd.Y
		}
		if cmd.FontSize != nil {
			titleInfo.FontSize = *cmd.FontSize
		}

		sim.SetTitleInfo(titleInfo)
	})
}

func (rt *CmdRunner) executeTime(cc *CommandContext, cmd *TimeCmd) {
	var dispTime uint64
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		dispTime = sim.Dispatcher().CurTime
	})
	cc.outputf("%d\n", dispTime)
}

func (rt *CmdRunner) executeNetInfo(cc *CommandContext, cmd *NetInfoCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		netinfo := sim.GetNetworkInfo()
		if cmd.Version != nil {
			netinfo.Version = *cmd.Version
		}
		if cmd.Commit != nil {
			netinfo.Commit = *cmd.Commit
		}
		if cmd.Real != nil {
			netinfo.Real = cmd.Real.Yes != nil
		}
		sim.SetNetworkInfo(netinfo)
	})
}

func (rt *CmdRunner) executeCoaps(cc *CommandContext, cmd *CoapsCmd) {
	if cmd.Enable != nil {
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			sim.Dispatcher().EnableCoaps()
		})
	} else {
		var coapMessages []*dispatcher.CoapMessage
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			coapMessages = sim.Dispatcher().CollectCoapMessages()
		})

		cc.outputItemsAsYaml(coapMessages)
	}
}

func (rt *CmdRunner) executeEnergy(cc *CommandContext, energy *EnergyCmd) {
	if energy.Save != nil {
		rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
			sim.GetEnergyAnalyser().SaveEnergyDataToFile(energy.Name, sim.Dispatcher().CurTime)
		})
	} else {
		cc.outputf("energy <command>\n")
		cc.outputf("\tsave [output name]\n")
	}
}

func (rt *CmdRunner) executeExe(cc *CommandContext, cmd *ExeCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		cfg := sim.GetConfig()
		ec := &cfg.ExeConfig
		isSetDefault := cmd.Default != nil
		isSetNodeType := len(cmd.NodeType.Val) > 0
		isSetVersion := len(cmd.Version.Val) > 0
		isSetPath := len(cmd.Path) > 0

		if isSetNodeType {
			// get or set the exe per individual node type.
			switch cmd.NodeType.Val {
			case FTD, ROUTER, REED, FED:
				if isSetPath {
					ec.Ftd = cmd.Path
				}
				cc.outputf("ftd: %s\n", ec.Ftd)
			case MTD, MED, SED, SSED:
				if isSetPath {
					ec.Mtd = cmd.Path
				}
				cc.outputf("mtd: %s\n", ec.Mtd)
			case BR:
				if isSetPath {
					ec.Br = cmd.Path
				}
				cc.outputf("br : %s\n", ec.Br)
			}
			return
		} else if isSetDefault && !isSetPath && !isSetNodeType && !isSetVersion {
			// set defaults for all node types.
			cfg.ExeConfig = cfg.ExeConfigDefault
		} else if isSetVersion && !isSetPath {
			// set executables to that of a named version for all node types except br.
			ec.SetVersion(cmd.Version.Val, &cfg.ExeConfigDefault)
		} else if !isSetDefault && !isSetNodeType && !isSetVersion && !isSetPath {
			// only display the exe output list.
		} else {
			cc.errorf("exe: unsupported combination of command options")
			return
		}

		cc.outputf("ftd: %s\n", ec.Ftd)
		cc.outputf("mtd: %s\n", ec.Mtd)
		cc.outputf("br : %s\n", ec.Br)
		cc.outputf("Executables search path: %s\n", ec.SearchPathsString())
		cc.outputf("Detected FTD path      : %s\n", ec.FindExecutable(ec.Ftd))
		cc.outputf("Detected MTD path      : %s\n", ec.FindExecutable(ec.Mtd))
		cc.outputf("Detected BR path       : %s\n", ec.FindExecutable(ec.Br))
	})
}

func (rt *CmdRunner) executeHelp(cc *CommandContext, cmd *HelpCmd) {
	if len(cmd.HelpTopic) > 0 {
		cc.outputStr(rt.help.outputCommandHelp(cmd.HelpTopic))
	} else {
		cc.outputStr(rt.help.outputGeneralHelp())
	}
}

func (rt *CmdRunner) executeKpi(cc *CommandContext, cmd *KpiCmd) {
	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		if len(cmd.Operation) == 0 {
			isRunning := sim.GetKpiManager().IsRunning()
			if isRunning {
				cc.outputf("on\n")
			} else {
				cc.outputf("off\n")
			}
			return
		}

		switch cmd.Operation {
		case "start":
			sim.GetKpiManager().Start()
		case "stop":
			sim.GetKpiManager().Stop()
		case "save":
			if len(cmd.Filename) == 0 {
				sim.GetKpiManager().SaveDefaultFile()
			} else {
				sim.GetKpiManager().SaveFile(cmd.Filename)
			}
		}
	})
}

func (rt *CmdRunner) executeSave(cc *CommandContext, cmd *SaveCmd) {
	var rootYaml yaml.Node

	// test filename if valid
	_, err := os.Stat(cmd.Filename)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		cc.errorf("Invalid save file name: %s", cmd.Filename)
		return
	}

	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		networkConfig := sim.ExportNetwork()
		nodesConfig := sim.ExportNodes(&networkConfig)

		root := simulation.YamlConfigFile{
			NetworkConfig: networkConfig,
			NodesList:     nodesConfig,
		}
		err = rootYaml.Encode(root)
		logger.PanicIfError(err)

		var data []byte
		data, err = yaml.Marshal(&rootYaml)
		err = os.WriteFile(cmd.Filename, data, 0644)
	})

	if err != nil {
		cc.errorf("Error writing file '%s': %v", cmd.Filename, err)
	}
}

func (rt *CmdRunner) executeLoad(cc *CommandContext, cmd *LoadCmd) {
	// test filename if valid
	fileInfo, err := os.Stat(cmd.Filename)
	if err != nil || fileInfo.IsDir() {
		cc.errorf("Invalid load file name: %s", cmd.Filename)
		return
	}

	rt.postAsyncWait(cc, func(sim *simulation.Simulation) {
		b, err := os.ReadFile(cmd.Filename)
		if err != nil {
			cc.errorf("Could not load file '%s': %v", cmd.Filename, err)
			return
		}
		cfgFile := simulation.YamlConfigFile{}
		err = yaml.Unmarshal(b, &cfgFile)
		if err != nil {
			cc.errorf("Error in YAML file: %v", err)
			return
		}
		if len(cfgFile.NodesList) == 0 {
			cc.errorf("No nodes defined in YAML file")
			return
		}

		if cmd.Add != nil {
			yamlMinNodeId := cfgFile.MinNodeId()
			nodeIdOffset := sim.MaxNodeId() + 1 - yamlMinNodeId
			cfgFile.NetworkConfig.BaseId = &nodeIdOffset
		}

		err = sim.ImportNodes(cfgFile.NetworkConfig, cfgFile.NodesList)
		if err != nil {
			cc.outputf("Warning: %v\n", err)
		}
	})
}
