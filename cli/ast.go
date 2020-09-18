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
	"strconv"

	. "github.com/openthread/ot-ns/types"

	"github.com/alecthomas/participle"
)

//noinspection GoStructTag
type Command struct {
	Add                 *AddCmd                 `  @@` //nolint
	ConfigVisualization *ConfigVisualizationCmd `| @@` //nolint
	CountDown           *CountDownCmd           `| @@` //nolint
	Counters            *CountersCmd            `| @@` //nolint
	Debug               *DebugCmd               `| @@` //nolint
	Del                 *DelCmd                 `| @@` //nolint
	DemoLegend          *DemoLegendCmd          `| @@` //nolint
	Exit                *ExitCmd                `| @@` //nolint
	Go                  *GoCmd                  `| @@` //nolint
	Joins               *JoinsCmd               `| @@` //nolint
	Move                *Move                   `| @@` //nolint
	NetInfo             *NetInfoCmd             `| @@` //nolint
	Node                *NodeCmd                `| @@` //nolint
	Nodes               *NodesCmd               `| @@` //nolint
	Now                 *NowCmd                 `| @@` //nolint
	Partitions          *PartitionsCmd          `| @@` //nolint
	Ping                *PingCmd                `| @@` //nolint
	Pings               *PingsCmd               `| @@` //nolint
	Plr                 *PlrCmd                 `| @@` //nolint
	Radio               *RadioCmd               `| @@` //nolint
	Scan                *ScanCmd                `| @@` //nolint
	Speed               *SpeedCmd               `| @@` //nolint
	Title               *TitleCmd               `| @@` //nolint
	Web                 *WebCmd                 `| @@` //nolint
}

//noinspection GoStructTag
type FullScreen struct {
	FullScreen struct{} `"fs"` //nolint
}

//noinspection GoStructTag
type RadioRange struct {
	Val int `"rr" @Int` //nolint
}

//noinspection GoStructTag
type FieldWidth struct {
	Val int `"fw" @Int` //nolint
}

//noinspection GoStructTag
type FieldHeight struct {
	Val int `"fh" @Int` //nolint
}

//noinspection GoStructTag
type VisualizeArg struct {
	Flag struct{}  `"v"`    //nolint
	None *NoneFlag `( @@`   //nolint
	Gui  *GuiFlag  `| @@ )` //nolint
}

//noinspection GoStructTag
type DebugCmd struct {
	Cmd  struct{} `"debug"`            //nolint
	Fail *string  `[ @"fail" ]`        //nolint
	Echo *string  `[ "echo" @String ]` //nolint
}

//noinspection GoStructTag
type GoCmd struct {
	Cmd     struct{}    `"go"`                      //nolint
	Seconds float64     `( (@Int|@Float)`           //nolint
	Ever    *EverFlag   `| @@ )`                    //nolint
	Speed   *float64    `( ("speed" (@Int|@Float))` //nolint
	NodeId  *NodeIdFlag `| @@ )*`                   //nolint
}

//noinspection GoStructTag
type NodeIdFlag struct {
	Flag struct{} `"node"` //nolint
	Id   int      `@Int`   //nolint
}

//noinspection GoStructTag
type NodeSelector struct {
	Id int `@Int` //nolint
}

func (ns *NodeSelector) String() string {
	return strconv.Itoa(ns.Id)
}

//noinspection GoStructTag
type Ipv6Address struct {
	Addr string `@String` //nolint
}

//noinspection GoStructTag
type AddrTypeFlag struct {
	Type AddrType `@( "any" | "mleid" | "rloc" | "aloc" | "linklocal" )` //nolint
}

//noinspection GoStructTag
type DataSizeFlag struct {
	Val int `("datasize"|"ds") @Int` //nolint
}

//noinspection GoStructTag
type IntervalFlag struct {
	Val int `("interval"|"itv") @Int` //nolint
}

//noinspection GoStructTag
type CountFlag struct {
	Val int `("count" | "c") @Int` //nolint
}

//noinspection GoStructTag
type HopLimitFlag struct {
	Val int `("hoplimit" | "hl") @Int` //nolint
}

//noinspection GoStructTag
type PingCmd struct {
	Cmd      struct{}      `"ping"`   //nolint
	Src      NodeSelector  `@@`       //nolint
	Dst      *NodeSelector `( @@`     //nolint
	AddrType *AddrTypeFlag `  [ @@ ]` //nolint
	DstAddr  *Ipv6Address  `| @@)`    //nolint
	DataSize *DataSizeFlag `( @@`     //nolint
	Count    *CountFlag    `| @@`     //nolint
	Interval *IntervalFlag `| @@`     //nolint
	HopLimit *HopLimitFlag `| @@ )*`  //nolint
}

//noinspection GoStructTag
type NetInfoCmd struct {
	Cmd     struct{}     `"netinfo" (`         //nolint
	Version *string      `  "version" @String` //nolint
	Commit  *string      `| "commit" @String`  //nolint
	Real    *YesOrNoFlag `| "real" @@ )+`      //nolint
}

//noinspection GoStructTag
type NodeCmd struct {
	Cmd     struct{}     `"node"`      //nolint
	Node    NodeSelector `@@`          //nolint
	Command *string      `[ @String ]` //nolint
}

//noinspection GoStructTag
type DemoLegendCmd struct {
	Cmd   struct{} `"demo_legend"` //nolint
	Title string   `@String`       //nolint
	X     int      `@Int`          //nolint
	Y     int      `@Int`          //nolint
}

//noinspection GoStructTag
type ConfigVisualizationCmd struct {
	Cmd              struct{}            `"cv"`    //nolint
	BroadcastMessage *CVBroadcastMessage `( @@`    //nolint
	UnicastMessage   *CVUnicastMessage   `| @@`    //nolint
	AckMessage       *CVAckMessage       `| @@`    //nolint
	RouterTable      *CVRouterTable      `| @@`    //nolint
	ChildTable       *CVChildTable       `| @@ )*` //nolint
}

//noinspection GoStructTag
type CVBroadcastMessage struct {
	Flag    struct{}    `"bro"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

//noinspection GoStructTag
type CVUnicastMessage struct {
	Flag    struct{}    `"uni"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

//noinspection GoStructTag
type CVAckMessage struct {
	Flag    struct{}    `"ack"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

//noinspection GoStructTag
type CVRouterTable struct {
	Flag    struct{}    `"rtb"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

//noinspection GoStructTag
type CVChildTable struct {
	Flag    struct{}    `"ctb"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

//noinspection GoStructTag
type CountDownCmd struct {
	Cmd     struct{} `"countdown"` //nolint
	Seconds int      `@Int`        //nolint
	Text    *string  `[ @String ]` //nolint
}

//noinspection GoStructTag
type ScanCmd struct {
	Cmd  struct{}     `"scan"` //nolint
	Node NodeSelector `@@`     // nolint
}

//noinspection GoStructTag
type SpeedCmd struct {
	Cmd   struct{}      `"speed"`               //nolint
	Max   *MaxSpeedFlag `( @@`                  //nolint
	Speed *float64      `| [ (@Int|@Float) ] )` //nolint
}

//noinspection GoStructTag
type TitleCmd struct {
	Cmd      struct{} `"title"`              //nolint
	Title    string   `@String`              //nolint
	X        *int     `( "x" (@Int|@Float) ` //nolint
	Y        *int     `| "y" (@Int|@Float) ` //nolint
	FontSize *int     `| "fs" @Int )*`       //nolint
}

//noinspection GoStructTag
type AddCmd struct {
	Cmd        struct{}        `"add"`                //nolint
	Type       NodeType        `@@`                   //nolint
	X          *int            `( "x" (@Int|@Float) ` //nolint
	Y          *int            `| "y" (@Int|@Float) ` //nolint
	Id         *AddNodeId      `| @@`                 //nolint
	RadioRange *RadioRangeFlag `| @@`                 //nolint
	Restore    *RestoreFlag    `| @@`                 //nolint
	Executable *ExecutableFlag `| @@`                 //nolint
	UartType   *UartTypeFlag   `| @@ )*`              //nolint
}

//noinspection GoStructTag
type RadioRangeFlag struct {
	Val int `"rr" @Int` //nolint
}

//noinspection GoStructTag
type RestoreFlag struct {
	Dummy struct{} `"restore"` //nolint
}

//noinspection GoStructTag
type ExecutableFlag struct {
	Dummy struct{} `"exe"`   //nolint
	Path  string   `@String` //nolint
}

//noinspection GoStructTag
type UartTypeFlag struct {
	Dummy struct{} `"uart"`                     //nolint
	Val   string   `@("virtual"|"real"|"auto")` //nolint
}

//noinspection MaxSpeedFlag
type MaxSpeedFlag struct {
	Dummy struct{} `( "max" | "inf")` //nolint
}

//noinspection GoStructTag
type NodeType struct {
	Val string `@("router"|"fed"|"med"|"sed"|"raw")` //nolint
}

//noinspection GoStructTag
type AddNodeId struct {
	Val int `"id" @Int` //nolint
}

//noinspection GoStructTag
type DelCmd struct {
	Cmd   struct{}       `"del"`   //nolint
	Nodes []NodeSelector `( @@ )+` //nolint
}

//noinspection GoStructTag
type EverFlag struct {
	Dummy struct{} `"ever"` //nolint
}

//noinspection GoStructTag
type Empty struct {
	Empty struct{} `""` //nolint
}

//noinspection GoStructTag
type ExitCmd struct {
	Cmd struct{} `"exit"` //nolint
}

//noinspection GoStructTag
type WebCmd struct {
	Cmd struct{} `"web"` //nolint
}

//noinspection GoStructTag
type RadioCmd struct {
	Cmd      struct{}        `"radio"` //nolint
	Nodes    []NodeSelector  `( @@ )+` //nolint
	On       *OnFlag         `( @@`    //nolint
	Off      *OffFlag        `| @@`    //nolint
	FailTime *FailTimeParams `| @@ )`  //nolint
}

//noinspection GoStructTag
type OnFlag struct {
	Dummy struct{} `"on"` //nolint
}

//noinspection GoStructTag
type OffFlag struct {
	Dummy struct{} `"off"` //nolint
}

//noinspection GoStructTag
type OnOrOffFlag struct {
	On  *OnFlag  `( @@`   //nolint
	Off *OffFlag `| @@ )` //nolint
}

//noinspection GoStructTag
type YesFlag struct {
	Dummy struct{} `("y"|"yes"|"true"|"1")` //nolint
}

//noinspection GoStructTag
type NoFlag struct {
	Dummy struct{} `("n"|"no"|"false"|"0")` //nolint
}

//noinspection GoStructTag
type YesOrNoFlag struct {
	Yes *YesFlag `( @@`   //nolint
	No  *NoFlag  `| @@ )` //nolint
}

//noinspection GoStructTag
type Move struct {
	Cmd    struct{}     `"move"` //nolint
	Target NodeSelector `@@`     //nolint
	X      int          `@Int`   //nolint
	Y      int          `@Int`   //nolint
}

//noinspection GoStructTag
type NodesCmd struct {
	Cmd struct{} `"nodes"` //nolint
}

//noinspection GoStructTag
type NowCmd struct {
	Cmd struct{} `"now"` //nolint
}

//noinspection GoStructTag
type PartitionsCmd struct {
	Cmd struct{} `( "partitions" | "pts")` //nolint
}

//noinspection GoStructTag
type PingsCmd struct {
	Cmd struct{} `"pings"` //nolint
}

//noinspection GoStructTag
type JoinsCmd struct {
	Cmd struct{} `"joins"` //nolint
}

//noinspection GoStructTag
type CountersCmd struct {
	Cmd struct{} `"counters"` //nolint
}

//noinspection GoStructTag
type PlrCmd struct {
	Cmd struct{} `"plr"`             //nolint
	Val *float64 `[ (@Int|@Float) ]` //nolint
}

//noinspection GoStructTag
type FailTimeParams struct {
	Dummy        struct{} `"ft"`          //nolint
	FailDuration float64  `(@Int|@Float)` //nolint
	FailInterval float64  `(@Int|@Float)` //nolint
}

//noinspection GoStructTag
type NoneFlag struct {
	Dummy struct{} `"none"` //nolint
}

//noinspection GoStructTag
type GuiFlag struct {
	Dummy struct{} `"gui"` //nolint
}

var (
	commandParser = participle.MustBuild(&Command{})
)

func ParseBytes(b []byte, cmd *Command) error {
	err := commandParser.ParseBytes(b, cmd)
	return err
}
