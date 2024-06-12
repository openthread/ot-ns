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
	"io"
	"regexp"
	"strings"

	. "github.com/openthread/ot-ns/types"
)

var (
	CommandInterruptedError = fmt.Errorf("command interrupted due to simulation exit")
)

var (
	doneOrErrorRegexp = regexp.MustCompile(`(Done|Error \d+: .*)`)
)

type NodeUartType int

const (
	nodeUartTypeUndefined   NodeUartType = iota
	nodeUartTypeRealTime    NodeUartType = iota
	nodeUartTypeVirtualTime NodeUartType = iota
)

// CmdRunner will point to an external package that can run a user's CLI commands.
type CmdRunner interface {
	// RunCommand will let the CmdRunner execute/run the user's CLI cmd. CLI output is sent to 'output'.
	RunCommand(cmd string, output io.Writer) error

	// GetNodeContext returns the current CLI context as selected by the user, or InvalidNodeId if none.
	GetNodeContext() NodeId
}

// NodeCounters keeps track of a node's internal diagnostic counters.
type NodeCounters map[string]int

// YamlConfigFile is the complete YAML structure for a config file for load/save.
type YamlConfigFile struct {
	ScriptConfig  YamlScriptConfig  `yaml:"script"`
	NetworkConfig YamlNetworkConfig `yaml:"network"`
	NodesList     []YamlNodeConfig  `yaml:"nodes"`
}

// YamlScriptConfig defines startup scripts for nodes, depending on node type.
type YamlScriptConfig struct {
	Mtd string `yaml:"mtd"`
	Ftd string `yaml:"ftd"`
	Br  string `yaml:"br"`
	All string `yaml:"all"`
}

// YamlNetworkConfig is a global network config that can be loaded/saved in YAML.
type YamlNetworkConfig struct {
	Position   [3]int `yaml:"pos-shift,flow"`        // provides an optional 3D position shift of all nodes.
	RadioRange *int   `yaml:"radio-range,omitempty"` // provides optional default radio-range.
	BaseId     *int   `yaml:"base-id,omitempty"`     // provides an optional node ID base (offset) for all nodes.
}

// YamlNodeConfig is a node config that can be loaded/saved in YAML.
type YamlNodeConfig struct {
	ID         int     `yaml:"id"`
	Type       string  `yaml:"type"`              // Node type (router, sed, fed, br, etc.)
	Version    *string `yaml:"version,omitempty"` // Thread version string or "" for default
	Position   [3]int  `yaml:"pos,flow"`
	RadioRange *int    `yaml:"radio-range,omitempty"`
}

func (yc *YamlConfigFile) MinNodeId() NodeId {
	var m NodeId = 0
	for _, n := range yc.NodesList {
		if n.ID < m || m == 0 {
			m = n.ID
		}
	}
	return m
}

func (ys *YamlScriptConfig) BuildMtdScript() []string {
	script := ys.Mtd + "\n" + ys.All
	return strings.Split(script, "\n")
}

func (ys *YamlScriptConfig) BuildFtdScript() []string {
	script := ys.Ftd + "\n" + ys.All
	return strings.Split(script, "\n")
}

func (ys *YamlScriptConfig) BuildBrScript() []string {
	script := ys.Ftd + "\n" + ys.Br + "\n" + ys.All
	return strings.Split(script, "\n")
}
