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

package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/openthread/ot-ns/web"

	"github.com/openthread/ot-ns/progctx"

	"github.com/openthread/ot-ns/dispatcher"

	"github.com/openthread/ot-ns/simulation"
	. "github.com/openthread/ot-ns/types"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
)

const (
	Prompt = "> "
)

type CommandContext struct {
	context.Context
	*Command
	rt  *CmdRunner
	err error
}

func (cc *CommandContext) outputf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (cc *CommandContext) errorf(format string, args ...interface{}) {
	cc.err = errors.Errorf(format, args...)
}

func (cc *CommandContext) error(err error) {
	cc.err = err
}

func (cc *CommandContext) Err() error {
	return cc.err
}

type CmdRunner struct {
	sim           *simulation.Simulation
	ctx           *progctx.ProgCtx
	contextNodeId NodeId
}

func (rt *CmdRunner) HandleCommand(cmdline string, output io.Writer) error {
	if rt.contextNodeId != InvalidNodeId && !isContextlessCommand(cmdline) {
		// run the command in node context
		cmd := &Command{
			Node: &NodeCmd{
				Node:    NodeSelector{Id: rt.contextNodeId},
				Command: &cmdline,
			},
		}
		cc := rt.Execute(cmd)

		if cc.Err() != nil {
			if _, err := fmt.Fprintf(output, "Error: %v\n", cc.Err()); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(output, "Done\n"); err != nil {
				return err
			}
		}
	} else {
		// run the OTNS-CLI command
		cmd := &Command{}
		if err := ParseBytes([]byte(cmdline), cmd); err != nil {
			if _, err := fmt.Fprintf(os.Stdout, "Error: %v\n", err); err != nil {
				return err
			}
		} else {
			cc := rt.Execute(cmd)

			if cc.Err() != nil {
				if _, err := fmt.Fprintf(os.Stdout, "Error: %v\n", cc.Err()); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(os.Stdout, "Done\n"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (rt *CmdRunner) GetPrompt() string {
	if rt.contextNodeId == InvalidNodeId {
		return Prompt
	} else {
		return fmt.Sprintf("node %d%s", rt.contextNodeId, Prompt)
	}
}

func (rt *CmdRunner) Execute(cmd *Command) (cc *CommandContext) {
	cc = &CommandContext{
		Command: cmd,
		rt:      rt,
	}

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
	} else if cmd.Scan != nil {
		rt.executeScan(cc, cc.Scan)
	} else if cmd.ConfigVisualization != nil {
		rt.executeConfigVisualization(cc, cc.ConfigVisualization)
	} else if cmd.Debug != nil {
		rt.executeDebug(cc, cmd.Debug)
	} else if cmd.DemoLegend != nil {
		rt.executeDemoLegend(cc, cmd.DemoLegend)
	} else if cmd.Exit != nil {
		rt.executeExit(cc, cmd.Exit)
	} else if cmd.Web != nil {
		rt.executeWeb(cc, cc.Web)
	} else {
		panic("not implemented")
	}
	return
}

func (rt *CmdRunner) executeGo(cc *CommandContext, cmd *GoCmd) {
	if cmd.Speed != nil {
		rt.postAsyncWait(func(sim *simulation.Simulation) {
			sim.SetSpeed(*cmd.Speed)
		})
	}
	var done <-chan struct{}

	if cmd.Ever == nil {
		rt.postAsyncWait(func(sim *simulation.Simulation) {
			done = sim.Go(time.Duration(float64(time.Second) * cmd.Seconds))
		})

		<-done
	} else {
		for {
			rt.postAsyncWait(func(sim *simulation.Simulation) {
				done = sim.Go(time.Hour) // run for ever
			})
			<-done

			if rt.ctx.Err() != nil {
				break
			}
		}
	}
}

func (rt *CmdRunner) executeSpeed(cc *CommandContext, cmd *SpeedCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		if cmd.Speed == nil && cmd.Max == nil {
			cc.outputf("%v\n", sim.GetSpeed())
		} else if cmd.Max != nil {
			sim.SetSpeed(dispatcher.MaxSimulateSpeed)
		} else {
			sim.SetSpeed(*cmd.Speed)
		}
	})
}

func (rt *CmdRunner) postAsyncWait(f func(sim *simulation.Simulation)) {
	done := make(chan struct{})
	rt.sim.PostAsync(false, func() {
		f(rt.sim)
		close(done)
	})
	<-done
}

func (rt *CmdRunner) executeAddNode(cc *CommandContext, cmd *AddCmd) {
	simplelogger.Infof("Add: %#v", *cmd)
	cfg := simulation.DefaultNodeConfig()
	if cmd.X != nil {
		cfg.X = *cmd.X
	}
	if cmd.Y != nil {
		cfg.Y = *cmd.Y
	}

	if cmd.Type.Val == "router" {
		cfg.IsRouter = true
		cfg.IsMtd = false
		cfg.RxOffWhenIdle = false
	} else if cmd.Type.Val == "fed" {
		cfg.IsRouter = false
		cfg.IsMtd = false
		cfg.RxOffWhenIdle = false
	} else if cmd.Type.Val == "med" {
		cfg.IsRouter = false
		cfg.IsMtd = true
		cfg.RxOffWhenIdle = false
	} else if cmd.Type.Val == "sed" {
		cfg.IsRouter = false
		cfg.IsMtd = true
		cfg.RxOffWhenIdle = true
	} else {
		panic("wrong node type")
	}

	if cmd.Id != nil {
		cfg.ID = cmd.Id.Val
	}

	if cmd.RadioRange != nil {
		cfg.RadioRange = cmd.RadioRange.Val
	}

	rt.postAsyncWait(func(sim *simulation.Simulation) {
		node, err := sim.AddNode(cfg)
		if err != nil {
			cc.error(err)
			return
		}

		cc.outputf("%d\n", node.Id)
	})
}

func (rt *CmdRunner) executeDelNode(cc *CommandContext, cmd *DelCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		for _, sel := range cmd.Nodes {
			node, _ := rt.getNode(sim, sel)
			if node == nil {
				cc.errorf("node %v not found", sel)
				continue
			}

			cc.error(sim.DeleteNode(node.Id))
		}
	})
}

func (rt *CmdRunner) executeExit(cc *CommandContext, cmd *ExitCmd) {
	if rt.enterNodeContext(InvalidNodeId) {
		return
	}

	rt.postAsyncWait(func(sim *simulation.Simulation) {
		sim.Stop()
	})
	rt.ctx.Cancel("exit")
}

func (rt *CmdRunner) executePing(cc *CommandContext, cmd *PingCmd) {
	simplelogger.Debugf("ping %#v", cmd)
	rt.postAsyncWait(func(sim *simulation.Simulation) {
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

		datasize := 4
		count := 1
		interval := 1
		hopLimit := 64

		if cmd.DataSize != nil {
			datasize = cmd.DataSize.Val
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
		return sim.Nodes()[sel.Id], sim.Dispatcher().Nodes()[sel.Id]
	}

	panic("node selector not implemented")
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

	if (addrType == nil || addrType.Type == AddrTypeAny) || addrType.Type == AddrTypeLinkLocal {
		addrs = append(addrs, node.GetIpAddrLinkLocal()...)
	}

	return addrs
}

func (rt *CmdRunner) executeDebug(cc *CommandContext, cmd *DebugCmd) {
	simplelogger.Infof("debug %#v", *cmd)

	if cmd.Echo != nil {
		cc.outputf("%s\n", *cmd.Echo)
	}

	if cmd.Fail != nil {
		cc.errorf("debug failed")
	}
}

func (rt *CmdRunner) executeNode(cc *CommandContext, cmd *NodeCmd) {
	contextNodeId := InvalidNodeId
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		node, _ := rt.getNode(sim, cmd.Node)
		if node == nil {
			cc.errorf("node not found")
			return
		}

		defer func() {
			err := recover()
			if err != nil {
				cc.errorf("%+v", err)
			}
		}()

		if cmd.Command != nil {
			output := node.Command(*cmd.Command, simulation.DefaultCommandTimeout)
			for _, line := range output {
				cc.outputf("%s\n", line)
			}
		} else {
			contextNodeId = node.Id
		}
	})

	if contextNodeId != InvalidNodeId {
		// enter node context
		rt.enterNodeContext(contextNodeId)
	}
}

func (rt *CmdRunner) executeDemoLegend(cc *CommandContext, cmd *DemoLegendCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		sim.ShowDemoLegend(cmd.X, cmd.Y, cmd.Title)
	})
}

func (rt *CmdRunner) executeCountDown(cc *CommandContext, cmd *CountDownCmd) {
	title := "%v"
	if cmd.Text != nil {
		title = *cmd.Text
	}
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		sim.CountDown(time.Duration(cmd.Seconds)*time.Second, title)
	})
}

func (rt *CmdRunner) executeRadio(cc *CommandContext, radio *RadioCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
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
				if radio.FailTime.FailInterval > 0 && radio.FailTime.FailDuration > 0 {
					dnode.SetFailTime(dispatcher.FailTime{
						FailDuration: uint64(radio.FailTime.FailDuration * 1000000),
						FailInterval: uint64(radio.FailTime.FailInterval * 1000000),
					})
				} else {
					dnode.SetFailTime(dispatcher.NonFailTime)
				}
			}
		}
	})
}

func (rt *CmdRunner) executeMoveNode(cc *CommandContext, cmd *Move) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		sim.MoveNodeTo(cmd.Target.Id, cmd.X, cmd.Y)
	})
}

func (rt *CmdRunner) executeLsNodes(cc *CommandContext, cmd *NodesCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		for nodeid := range sim.Nodes() {
			dnode := sim.Dispatcher().GetNode(nodeid)
			var line strings.Builder
			line.WriteString(fmt.Sprintf("id=%d\textaddr=%016x\trloc16=%04x\tx=%d\ty=%d\tfailed=%v", nodeid, dnode.ExtAddr, dnode.Rloc16,
				dnode.X, dnode.Y, dnode.IsFailed()))
			cc.outputf("%s\n", line.String())
		}
	})
}

func (rt *CmdRunner) executeLsPartitions(cc *CommandContext) {
	pars := map[uint32][]NodeId{}

	rt.postAsyncWait(func(sim *simulation.Simulation) {
		for nodeid, dnode := range sim.Dispatcher().Nodes() {
			parid := dnode.PartitionId
			pars[parid] = append(pars[parid], nodeid)
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
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		d := sim.Dispatcher()
		for nodeid, node := range d.Nodes() {
			pings := node.CollectPings()
			if len(pings) > 0 {
				allPings[nodeid] = pings
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

	rt.postAsyncWait(func(sim *simulation.Simulation) {
		d := sim.Dispatcher()
		for nodeid, node := range d.Nodes() {
			joins := node.CollectJoins()
			if len(joins) > 0 {
				allJoins[nodeid] = joins
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
	rt.postAsyncWait(func(sim *simulation.Simulation) {
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
	if err := web.OpenWeb(rt.ctx); err != nil {
		cc.error(err)
	}
}

func (rt *CmdRunner) executePlr(cc *CommandContext, cmd *PlrCmd) {
	if cmd.Val == nil {
		// get PLR
		var plr float64

		rt.postAsyncWait(func(sim *simulation.Simulation) {
			plr = sim.Dispatcher().GetGlobalMessageDropRatio()
		})

		cc.outputf("%v\n", plr)
	} else {
		// set PLR
		rt.postAsyncWait(func(sim *simulation.Simulation) {
			sim.Dispatcher().SetGlobalPacketLossRatio(*cmd.Val)
			*cmd.Val = sim.Dispatcher().GetGlobalMessageDropRatio()
		})
		cc.outputf("%v\n", *cmd.Val)
	}
}

func (rt *CmdRunner) executeScan(cc *CommandContext, cmd *ScanCmd) {
	rt.postAsyncWait(func(sim *simulation.Simulation) {
		node, _ := rt.getNode(sim, cmd.Node)
		if node == nil {
			cc.errorf("node not found")
			return
		}

		node.CommandExpectNone("scan", simulation.DefaultCommandTimeout)
	})

	timeout := time.Millisecond * 600 // FIXME: hardcoding
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		rt.postAsyncWait(func(sim *simulation.Simulation) {
			node, _ := rt.getNode(sim, cmd.Node)
			if node == nil {
				return
			}
			node.AssurePrompt()
		})
	}
}

func (rt *CmdRunner) executeConfigVisualization(cc *CommandContext, cmd *ConfigVisualizationCmd) {
	var opts dispatcher.VisualizationOptions
	rt.postAsyncWait(func(sim *simulation.Simulation) {
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
	simplelogger.AssertTrue(nodeid == InvalidNodeId || nodeid > 0)
	if rt.contextNodeId == nodeid {
		return false
	}

	rt.contextNodeId = nodeid
	return true
}

func NewCmdRunner(ctx *progctx.ProgCtx, sim *simulation.Simulation) *CmdRunner {
	r := &CmdRunner{
		ctx:           ctx,
		sim:           sim,
		contextNodeId: InvalidNodeId,
	}
	return r
}
