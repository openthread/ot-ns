// Copyright (c) 2026, The OTNS Authors.
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
	"context"
	"encoding/binary"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openthread/ot-ns/dispatcher"
	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/prng"
	"github.com/openthread/ot-ns/progctx"
	. "github.com/openthread/ot-ns/types"
)

const testHelperModeEnv = "OTNS_SIM_TEST_HELPER_MODE"

func TestMain(m *testing.M) {
	if mode := os.Getenv(testHelperModeEnv); mode != "" {
		runTestHelperNode(mode)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func serializeNodeInfoEvent(nodeId int) []byte {
	payload := make([]byte, 4)
	binary.LittleEndian.PutUint32(payload, uint32(nodeId))
	ev := &event.Event{
		Type: event.EventTypeNodeInfo,
		Data: payload,
	}
	return ev.Serialize()
}

func runTestHelperNode(mode string) {
	if len(os.Args) < 3 {
		os.Exit(2)
	}
	nodeId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(2)
	}
	sockName := os.Args[2]

	conn, err := net.Dial("unix", sockName)
	if err != nil {
		os.Exit(2)
	}
	defer conn.Close()

	// Send EventTypeNodeInfo first, just like otSysInit().
	if _, err = conn.Write(serializeNodeInfoEvent(nodeId)); err != nil {
		return
	}

	switch mode {
	case "hang_after_nodeinfo":
		// Stay connected without sending EventTypeAlarmFired so dnode.IsConnected() is true
		// while s.d.IsAlive(nodeid) remains true at deadline.
		buf := make([]byte, 256)
		_, _ = conn.Read(buf)
		return
	case "delay_after_nodeinfo":
		// Delay past a single dispatcher.DefaultReadTimeout window before finishing startup.
		time.Sleep(dispatcher.DefaultReadTimeout + 300*time.Millisecond)
	case "fast":
		// Normal fast startup.
	}

	extAddr := make([]byte, 8)
	binary.BigEndian.PutUint64(extAddr, uint64(0x1020304050600000)+uint64(nodeId))
	extEv := &event.Event{
		Type: event.EventTypeExtAddr,
		Data: extAddr,
	}
	if _, err = conn.Write(extEv.Serialize()); err != nil {
		return
	}

	sleepEv := &event.Event{
		Delay: 1000000,
		Type:  event.EventTypeAlarmFired,
		MsgId: 0,
	}
	if _, err = conn.Write(sleepEv.Serialize()); err != nil {
		return
	}

	// Minimal virtual-time UART responder for AddNode's version/threadversion queries.
	buf := make([]byte, 4096)
	var pending []byte
	for {
		n, readErr := conn.Read(buf)
		if n > 0 {
			pending = append(pending, buf[:n]...)
			for len(pending) > 0 {
				inEv := &event.Event{}
				used := inEv.Deserialize(pending)
				if used == 0 {
					break
				}
				pending = pending[used:]

				if inEv.Type == event.EventTypeUartWrite {
					cmd := strings.TrimSpace(string(inEv.Data))
					lines := []string{cmd}
					switch cmd {
					case "version":
						lines = append(lines, "OPENTHREAD/test-mock; RFSIM; Jan  1 2026 00:00:00", "Done")
					case "thread version":
						lines = append(lines, "5", "Done")
					default:
						lines = append(lines, "Done")
					}
					for _, l := range lines {
						uartOut := &event.Event{
							Type:  event.EventTypeUartWrite,
							MsgId: inEv.MsgId,
							Data:  []byte(l + "\r\n"),
						}
						if _, err = conn.Write(uartOut.Serialize()); err != nil {
							return
						}
					}
				}

				doneSleep := &event.Event{
					Delay: 1000000,
					Type:  event.EventTypeAlarmFired,
					MsgId: inEv.MsgId,
				}
				if _, err = conn.Write(doneSleep.Serialize()); err != nil {
					return
				}
			}
		}
		if readErr != nil {
			return
		}
	}
}

func newTestSimulation(t *testing.T, simId int) (*Simulation, func()) {
	t.Helper()
	prng.Init(1)
	outDir := t.TempDir()
	ctx := progctx.New(context.Background())

	cfg := DefaultConfig()
	cfg.Id = simId
	cfg.OutputDir = outDir
	cfg.AutoGo = false
	cfg.LogLevel = logger.WarnLevel
	cfg.LogFileLevel = logger.OffLevel

	dispCfg := dispatcher.DefaultConfig()
	dispCfg.SimulationId = simId
	dispCfg.OutputDir = outDir
	dispCfg.PcapEnabled = false

	sim, err := NewSimulation(ctx, cfg, dispCfg)
	require.NoError(t, err)

	cleanup := func() {
		sim.Stop()
		sim.d.Stop()
	}
	return sim, cleanup
}

func testRawNodeConfig(t *testing.T, nodeId int) *NodeConfig {
	t.Helper()
	exePath, err := os.Executable()
	require.NoError(t, err)

	ncfg := DefaultNodeConfig()
	ncfg.ID = nodeId
	ncfg.Type = MED
	ncfg.IsMtd = true
	ncfg.IsRouter = false
	ncfg.IsRaw = true
	ncfg.NodeLogFile = false
	ncfg.ExecutablePath = exePath
	return &ncfg
}

func TestAddNodeSlowStartupAfterNodeInfo(t *testing.T) {
	sim, cleanup := newTestSimulation(t, 971)
	defer cleanup()

	t.Setenv(testHelperModeEnv, "delay_after_nodeinfo")
	node, err := sim.AddNode(testRawNodeConfig(t, 1))
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.False(t, sim.Dispatcher().IsAlive(1))
	assert.Equal(t, 1, len(sim.Nodes()))

	require.NoError(t, sim.DeleteNode(1))
	assert.Equal(t, 0, len(sim.Nodes()))
}

func TestAddNodeIncompleteStartupReturnsErrorAndCleansUp(t *testing.T) {
	sim, cleanup := newTestSimulation(t, 972)
	defer cleanup()

	// Limit to a single RecvEvents() window so the test completes quickly.
	sim.nodeStartTimeout = time.Millisecond * 100

	t.Setenv(testHelperModeEnv, "hang_after_nodeinfo")
	var node *Node
	var err error
	assert.NotPanics(t, func() {
		node, err = sim.AddNode(testRawNodeConfig(t, 1))
	})
	require.Error(t, err)
	assert.Nil(t, node)
	assert.Contains(t, err.Error(), "did not finish startup")
	assert.Equal(t, 0, len(sim.Nodes()))
	assert.Nil(t, sim.Dispatcher().GetNode(1))
	assert.False(t, sim.Dispatcher().IsAlive(1))

	// Verify the same node ID can be added cleanly immediately afterwards.
	sim.nodeStartTimeout = NodeStartTimeout
	t.Setenv(testHelperModeEnv, "fast")
	node, err = sim.AddNode(testRawNodeConfig(t, 1))
	require.NoError(t, err)
	require.NotNil(t, node)
	assert.Equal(t, 1, len(sim.Nodes()))
}
