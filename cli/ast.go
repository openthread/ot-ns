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

// This file defines the format of all CLI commands and their flags.

package cli

import (
	"strconv"

	"github.com/alecthomas/participle"
)

type command struct {
	Add        *AddCmd        `  @@` //nolint
	CountDown  *CountDownCmd  `| @@` //nolint
	Counters   *CountersCmd   `| @@` //nolint
	Debug      *DebugCmd      `| @@` //nolint
	Del        *DelCmd        `| @@` //nolint
	DemoLegend *DemoLegendCmd `| @@` //nolint
	Exit       *ExitCmd       `| @@` //nolint
	Go         *GoCmd         `| @@` //nolint
	Joins      *JoinsCmd      `| @@` //nolint
	Move       *MoveCmd       `| @@` //nolint
	Node       *NodeCmd       `| @@` //nolint
	Nodes      *NodesCmd      `| @@` //nolint
	Partitions *PartitionsCmd `| @@` //nolint
	Ping       *PingCmd       `| @@` //nolint
	Pings      *PingsCmd      `| @@` //nolint
	Plr        *PlrCmd        `| @@` //nolint
	Radio      *RadioCmd      `| @@` //nolint
	Scan       *ScanCmd       `| @@` //nolint
	Speed      *SpeedCmd      `| @@` //nolint
	Web        *WebCmd        `| @@` //nolint
}

// DebugCmd defines the `debug` command format.
type DebugCmd struct {
	Cmd  struct{} `"debug"`            //nolint
	Fail *string  `[ @"fail" ]`        //nolint
	Echo *string  `[ "echo" @String ]` //nolint
}

// GoCmd defines the `go` command format.
type GoCmd struct {
	Cmd     struct{}  `"go"`                      //nolint
	Seconds float64   `( (@Int|@Float)`           //nolint
	Ever    *EverFlag `| @@ )`                    //nolint
	Speed   *float64  `[ "speed" (@Int|@Float) ]` //nolint
}

// NodeSelector defines the node selector format.
type NodeSelector struct {
	Id int `@Int` //nolint
}

func (ns *NodeSelector) String() string {
	return strconv.Itoa(ns.Id)
}

// Ipv6Address defines the IPv6 address format.
type Ipv6Address struct {
	Addr string `@String` //nolint
}

// AddrType defines the `address type` flag format for specifying address type.
type AddrType struct {
	Type string `@( "any" | "mleid" | "rloc" | "aloc" | "linklocal" )` //nolint
}

// DataSizeFlag defines the `datasize` flag format for specifying data size.
type DataSizeFlag struct {
	Val int `("datasize"|"ds") @Int` //nolint
}

// IntervalFlag defines the `interval` flag format for specifying time interval.
type IntervalFlag struct {
	Val int `("interval"|"itv") @Int` //nolint
}

// CountFlag defines the `count` flag format for specifying count.
type CountFlag struct {
	Val int `("count" | "c") @Int` //nolint
}

// HopLimitFlag defines the `hop limit` flag format for specifying hop limit.
type HopLimitFlag struct {
	Val int `("hoplimit" | "hl") @Int` //nolint
}

// PingCmd defines the `ping` command format.
type PingCmd struct {
	Cmd      struct{}      `"ping"`   //nolint
	Src      NodeSelector  `@@`       //nolint
	Dst      *NodeSelector `( @@`     //nolint
	AddrType *AddrType     `  [ @@ ]` //nolint
	DstAddr  *Ipv6Address  `| @@)`    //nolint
	DataSize *DataSizeFlag `( @@`     //nolint
	Count    *CountFlag    `| @@`     //nolint
	Interval *IntervalFlag `| @@`     //nolint
	HopLimit *HopLimitFlag `| @@ )*`  //nolint
}

// NodeCmd defines `node` command format.
type NodeCmd struct {
	Cmd     struct{}     `"node"`      //nolint
	Node    NodeSelector `@@`          //nolint
	Command *string      `[ @String ]` //nolint
}

// DemoLegendCmd defines the `demo_legend` command format.
type DemoLegendCmd struct {
	Cmd   struct{} `"demo_legend"` //nolint
	Title string   `@String`       //nolint
	X     int      `@Int`          //nolint
	Y     int      `@Int`          //nolint
}

// CountDownCmd defines the `countdown` command format.
type CountDownCmd struct {
	Cmd     struct{} `"countdown"` //nolint
	Seconds int      `@Int`        //nolint
	Text    *string  `[ @String ]` //nolint
}

// ScanCmd defines the `scan` command format.
type ScanCmd struct {
	Cmd  struct{}     `"scan"` //nolint
	Node NodeSelector `@@`     // nolint
}

// SpeedCmd defines the `speed` command format.
type SpeedCmd struct {
	Cmd   struct{}      `"speed"`               //nolint
	Max   *MaxSpeedFlag `( @@`                  //nolint
	Speed *float64      `| [ (@Int|@Float) ] )` //nolint
}

// AddCmd defines the `add` command format.
type AddCmd struct {
	Cmd        struct{}        `"add"`                //nolint
	Type       NodeType        `@@`                   //nolint
	X          *int            `( "x" (@Int|@Float) ` //nolint
	Y          *int            `| "y" (@Int|@Float) ` //nolint
	Id         *AddNodeId      `| @@`                 //nolint
	RadioRange *RadioRangeFlag `|@@ )*`               //nolint
}

// RadioRangeFlag defines the `radio range` flag format.
type RadioRangeFlag struct {
	Val int `"rr" @Int` //nolint
}

// MaxSpeedFlag defines the `max speed` flag format.
type MaxSpeedFlag struct {
	Dummy struct{} `( "max" | "inf")` //nolint
}

// NodeType defines the `node type` flag for specifying node types.
type NodeType struct {
	Val string `@("router"|"fed"|"med"|"sed")` //nolint
}

// AddNodeId defines the `id` flag format for specifying node ID.
type AddNodeId struct {
	Val int `"id" @Int` //nolint
}

// DelCmd defines the `del` command format.
type DelCmd struct {
	Cmd   struct{}       `"del"`   //nolint
	Nodes []NodeSelector `( @@ )+` //nolint
}

// EverFlag defines the `ever` flag format.
type EverFlag struct {
	Dummy struct{} `"ever"` //nolint
}

// ExitCmd defines the `exit` command format.
type ExitCmd struct {
	Cmd struct{} `"exit"` //nolint
}

// WebCmd defines the `web` command format.
type WebCmd struct {
	Cmd struct{} `"web"` //nolint
}

// RadioCmd defines the `radio` command format.
type RadioCmd struct {
	Cmd      struct{}        `"radio"` //nolint
	Nodes    []NodeSelector  `( @@ )+` //nolint
	On       *OnFlag         `( @@`    //nolint
	Off      *OffFlag        `| @@`    //nolint
	FailTime *FailTimeParams `| @@ )`  //nolint
}

// OnFlag defines the `on` flag format.
type OnFlag struct {
	Dummy struct{} `"on"` //nolint
}

// OffFlag defines the `off` flag format.
type OffFlag struct {
	Dummy struct{} `"off"` //nolint
}

// MoveCmd defines the `move` command format.
type MoveCmd struct {
	Cmd    struct{}     `"move"` //nolint
	Target NodeSelector `@@`     //nolint
	X      int          `@Int`   //nolint
	Y      int          `@Int`   //nolint
}

// NodesCmd defines the `nodes` command format.
type NodesCmd struct {
	Cmd struct{} `"nodes"` //nolint
}

// PartitionsCmd defines the `partitions` command format.
type PartitionsCmd struct {
	Cmd struct{} `( "partitions" | "pts")` //nolint
}

// PingsCmd defines the `pings` command format.
type PingsCmd struct {
	Cmd struct{} `"pings"` //nolint
}

// JoinsCmd defines the `joins` command format.
type JoinsCmd struct {
	Cmd struct{} `"joins"` //nolint
}

// CountersCmd defies the `counters` command format.
type CountersCmd struct {
	Cmd struct{} `"counters"` //nolint
}

// PlrCmd defines the `plr` command format.
type PlrCmd struct {
	Cmd struct{} `"plr"`             //nolint
	Val *float64 `[ (@Int|@Float) ]` //nolint
}

// FailTimeParams defines the fail time parameters format.
type FailTimeParams struct {
	Dummy        struct{} `"ft"`          //nolint
	FailDuration float64  `(@Int|@Float)` //nolint
	FailInterval float64  `(@Int|@Float)` //nolint
}

var (
	commandParser = participle.MustBuild(&command{})
)

func parseCmdBytes(b []byte, cmd *command) error {
	err := commandParser.ParseBytes(b, cmd)
	return err
}
