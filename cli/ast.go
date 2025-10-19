// Copyright (c) 2020-2025, The OTNS Authors.
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
	"strings"

	"github.com/alecthomas/participle"

	. "github.com/openthread/ot-ns/types"
)

// noinspection GoVetStructTag
type Command struct {
	Add                 *AddCmd                 `  @@` //nolint
	AutoGo              *AutoGoCmd              `| @@` //nolint
	Coaps               *CoapsCmd               `| @@` //nolint
	ConfigVisualization *ConfigVisualizationCmd `| @@` //nolint
	CountDown           *CountDownCmd           `| @@` //nolint
	Counters            *CountersCmd            `| @@` //nolint
	Debug               *DebugCmd               `| @@` //nolint
	Del                 *DelCmd                 `| @@` //nolint
	DemoLegend          *DemoLegendCmd          `| @@` //nolint
	Energy              *EnergyCmd              `| @@` //nolint
	Exe                 *ExeCmd                 `| @@` //nolint
	Exit                *ExitCmd                `| @@` //nolint
	Go                  *GoCmd                  `| @@` //nolint
	Help                *HelpCmd                `| @@` //nolint
	Host                *HostCmd                `| @@` //nolint
	Joins               *JoinsCmd               `| @@` //nolint
	Kpi                 *KpiCmd                 `| @@` //nolint
	Load                *LoadCmd                `| @@` //nolint
	LogLevel            *LogLevelCmd            `| @@` //nolint
	Move                *MoveCmd                `| @@` //nolint
	NetInfo             *NetInfoCmd             `| @@` //nolint
	Node                *NodeCmd                `| @@` //nolint
	Nodes               *NodesCmd               `| @@` //nolint
	Partitions          *PartitionsCmd          `| @@` //nolint
	Ping                *PingCmd                `| @@` //nolint
	Pings               *PingsCmd               `| @@` //nolint
	Plr                 *PlrCmd                 `| @@` //nolint
	Radio               *RadioCmd               `| @@` //nolint
	RadioModel          *RadioModelCmd          `| @@` //nolint
	RadioParam          *RadioParamCmd          `| @@` //nolint
	RfSim               *RfSimCmd               `| @@` //nolint
	Save                *SaveCmd                `| @@` //nolint
	Scan                *ScanCmd                `| @@` //nolint
	Send                *SendCmd                `| @@` //nolint
	Speed               *SpeedCmd               `| @@` //nolint
	Time                *TimeCmd                `| @@` //nolint
	Title               *TitleCmd               `| @@` //nolint
	Unwatch             *UnwatchCmd             `| @@` //nolint
	Watch               *WatchCmd               `| @@` //nolint
	Web                 *WebCmd                 `| @@` //nolint
}

// noinspection GoVetStructTag
type FullScreen struct {
	FullScreen struct{} `"fs"` //nolint
}

// noinspection GoVetStructTag
type RadioRange struct {
	Val int `"rr" @Int` //nolint
}

// noinspection GoVetStructTag
type FieldWidth struct {
	Val int `"fw" @Int` //nolint
}

// noinspection GoVetStructTag
type FieldHeight struct {
	Val int `"fh" @Int` //nolint
}

// noinspection GoVetStructTag
type VisualizeArg struct {
	Flag struct{}  `"v"`    //nolint
	None *NoneFlag `( @@`   //nolint
	Gui  *GuiFlag  `| @@ )` //nolint
}

// noinspection GoVetStructTag
type DebugCmd struct {
	Cmd  struct{} `"debug"`            //nolint
	Fail *string  `[ @"fail" ]`        //nolint
	Echo *string  `[ "echo" @String ]` //nolint
}

// noinspection GoVetStructTag
type AutoGoCmd struct {
	Cmd struct{}     `"autogo"` //nolint
	Val *YesOrNoFlag `[ @@ ]`   // nolint
}

// noinspection GoVetStructTag
type GoCmd struct {
	Cmd   struct{}  `"go"`                                     //nolint
	Time  string    `( @((Int|Float)["h"|"us"|"m"|"ms"|"s"]) ` //nolint
	Ever  *EverFlag `| @@ )`                                   //nolint
	Speed *float64  `[ "speed" (@Int|@Float) ]`                //nolint
}

// noinspection GoVetStructTag
type NodeSelector struct {
	Id      int     `( @Int`       //nolint
	All     *string `| @"all" )`   //nolint
	IdRange int     `[ "-" @Int ]` //nolint
}

type NodeSelectorSlice []NodeSelector

func (ns *NodeSelector) String() string {
	if ns.All != nil {
		return "all"
	}
	if ns.IdRange > 0 {
		return strconv.Itoa(ns.Id) + "-" + strconv.Itoa(ns.IdRange)
	}
	return strconv.Itoa(ns.Id)
}

// String creates a string of NodeID numbers from a []NodeSelector.
func (ns NodeSelectorSlice) String() string {
	var line strings.Builder
	for _, n := range ns {
		line.WriteString(n.String())
		line.WriteRune(' ')
	}
	return line.String()
}

// noinspection GoVetStructTag
type Ipv6Address struct {
	Addr string `@String` //nolint
}

// noinspection GoVetStructTag
type AddrTypeFlag struct {
	Type AddrType `@( "any" | "mleid" | "rloc" | "slaac" | "linklocal" )` //nolint
}

// noinspection GoVetStructTag
type DataSizeFlag struct {
	Val int `("datasize"|"ds") @Int` //nolint
}

// noinspection GoVetStructTag
type IntervalFlag struct {
	Val float64 `("interval"|"itv") (@Int|@Float)` //nolint
}

// noinspection GoVetStructTag
type CountFlag struct {
	Val int `("count" | "c") @Int` //nolint
}

// noinspection GoVetStructTag
type HopLimitFlag struct {
	Val int `("hoplimit" | "hl") @Int` //nolint
}

// noinspection GoVetStructTag
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

// noinspection GoVetStructTag
type NetInfoCmd struct {
	Cmd     struct{} `"netinfo" (`           //nolint
	Version *string  `  "version" @String`   //nolint
	Commit  *string  `| "commit" @String )+` //nolint
}

// noinspection GoVetStructTag
type NodeCmd struct {
	Cmd     struct{}     `"node"`      //nolint
	Node    NodeSelector `@@`          //nolint
	Command *string      `[ @String ]` //nolint
}

// noinspection GoVetStructTag
type DemoLegendCmd struct {
	Cmd   struct{} `"demo_legend"` //nolint
	Title string   `@String`       //nolint
	X     int      `@Int`          //nolint
	Y     int      `@Int`          //nolint
}

// noinspection GoVetStructTag
type ConfigVisualizationCmd struct {
	Cmd              struct{}            `"cv"`    //nolint
	BroadcastMessage *CVBroadcastMessage `( @@`    //nolint
	UnicastMessage   *CVUnicastMessage   `| @@`    //nolint
	AckMessage       *CVAckMessage       `| @@`    //nolint
	RouterTable      *CVRouterTable      `| @@`    //nolint
	ChildTable       *CVChildTable       `| @@ )*` //nolint
}

// noinspection GoVetStructTag
type CVBroadcastMessage struct {
	Flag    struct{}    `"bro"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

// noinspection GoVetStructTag
type CVUnicastMessage struct {
	Flag    struct{}    `"uni"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

// noinspection GoVetStructTag
type CVAckMessage struct {
	Flag    struct{}    `"ack"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

// noinspection GoVetStructTag
type CVRouterTable struct {
	Flag    struct{}    `"rtb"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

// noinspection GoVetStructTag
type CVChildTable struct {
	Flag    struct{}    `"ctb"` //nolint
	OnOrOff OnOrOffFlag `@@`    //nolint
}

// noinspection GoVetStructTag
type CountDownCmd struct {
	Cmd     struct{} `"countdown"` //nolint
	Seconds int      `@Int`        //nolint
	Text    *string  `[ @String ]` //nolint
}

// noinspection GoVetStructTag
type ScanCmd struct {
	Cmd  struct{}     `"scan"` //nolint
	Node NodeSelector `@@`     // nolint
}

// noinspection GoVetStructTag
type SpeedCmd struct {
	Cmd   struct{}      `"speed"`             //nolint
	Max   *MaxSpeedFlag `[ ( @@`              //nolint
	Speed *float64      `| (@Int|@Float) ) ]` //nolint
}

// noinspection GoVetStructTag
type TimeCmd struct {
	Cmd struct{} `"time"` //nolint
}

// noinspection GoVetStructTag
type TitleCmd struct {
	Cmd      struct{} `"title"`              //nolint
	Title    string   `@String`              //nolint
	X        *int     `( "x" (@Int|@Float) ` //nolint
	Y        *int     `| "y" (@Int|@Float) ` //nolint
	FontSize *int     `| "fs" @Int )*`       //nolint
}

// noinspection GoVetStructTag
type AddCmd struct {
	Cmd        struct{}        `"add"`                //nolint
	Type       NodeTypeOrRole  `@@`                   //nolint
	X          *int            `( "x" (@Int|@Float) ` //nolint
	Y          *int            `| "y" (@Int|@Float) ` //nolint
	Z          *int            `| "z" (@Int|@Float) ` //nolint
	Id         *AddNodeId      `| @@`                 //nolint
	RadioRange *RadioRangeFlag `| @@`                 //nolint
	Restore    *RestoreFlag    `| @@`                 //nolint
	Version    *ThreadVersion  `| @@`                 //nolint
	Raw        *RawFlag        `| @@`                 //nolint
	Executable *ExecutableFlag `| @@ )*`              //nolint
}

// noinspection GoVetStructTag
type NodeTypeOrRole struct {
	Val string `@("router"|"reed"|"fed"|"med"|"sed"|"ssed"|"br"|"mtd"|"ftd"|"wifi")` //nolint
}

// noinspection GoVetStructTag
type AddNodeId struct {
	Val int `"id" @Int` //nolint
}

// noinspection GoVetStructTag
type RadioRangeFlag struct {
	Val int `"rr" @Int` //nolint
}

// noinspection GoVetStructTag
type RestoreFlag struct {
	Dummy struct{} `"restore"` //nolint
}

// noinspection GoVetStructTag
type ThreadVersion struct {
	Val string `@("v11"|"v12"|"v13"|"v14"|"ccm")` //nolint
}

// noinspection GoVetStructTag
type ExecutableFlag struct {
	Dummy struct{} `"exe"`   //nolint
	Path  string   `@String` //nolint
}

// noinspection GoVetStructTag
type RawFlag struct {
	Dummy struct{} `"raw"` //nolint
}

// noinspection GoVetStructTag
type MaxSpeedFlag struct {
	Dummy struct{} `( "max" | "inf")` //nolint
}

// noinspection GoVetStructTag
type AddFlag struct {
	Dummy struct{} `"add"` //nolint
}

// noinspection GoVetStructTag
type CoapsCmd struct {
	Cmd    struct{}    `"coaps"` //nolint
	Enable *EnableFlag `@@ ?`    //nolint
}

// noinspection GoVetStructTag
type EnableFlag struct {
	Dummy struct{} `"enable"` //nolint
}

// noinspection GoVetStructTag
type DelCmd struct {
	Cmd   struct{}       `"del"`   //nolint
	Nodes []NodeSelector `( @@ )+` //nolint
}

// noinspection GoVetStructTag
type EverFlag struct {
	Dummy struct{} `"ever"` //nolint
}

// noinspection GoVetStructTag
type Empty struct {
	Empty struct{} `""` //nolint
}

// noinspection GoVetStructTag
type ExitCmd struct {
	Cmd struct{} `"exit"` //nolint
}

// noinspection GoVetStructTag
type WebCmd struct {
	Cmd     struct{} `"web"`                            //nolint
	TabName *string  `@[ "main" | "energy" | "stats" ]` //nolint
}

// noinspection GoVetStructTag
type EnergyCmd struct {
	Cmd  struct{}  `"energy"` //nolint
	Save *SaveFlag `( @@ )?`  //nolint
	Name string    `@String?` //nolint
}

// noinspection GoVetStructTag
type SaveFlag struct {
	Dummy struct{} `"save"` //nolint
}

// noinspection GoVetStructTag
type ExeCmd struct {
	Cmd      struct{}       `"exe"`       //nolint
	NodeType NodeTypeOrRole `( @@`        //nolint
	Default  *DefaultFlag   `| @@`        //nolint
	Version  ThreadVersion  `| @@ )?`     //nolint
	Path     string         `[ @String ]` //nolint
}

// noinspection GoVetStructTag
type DefaultFlag struct {
	Dummy struct{} `"default"` //nolint
}

// noinspection GoVetStructTag
type RadioCmd struct {
	Cmd      struct{}        `"radio"` //nolint
	Nodes    []NodeSelector  `( @@ )+` //nolint
	On       *OnFlag         `( @@`    //nolint
	Off      *OffFlag        `| @@`    //nolint
	FailTime *FailTimeParams `| @@ )`  //nolint
}

// noinspection GoVetStructTag
type OnFlag struct {
	Dummy struct{} `"on"` //nolint
}

// noinspection GoVetStructTag
type OffFlag struct {
	Dummy struct{} `"off"` //nolint
}

// noinspection GoVetStructTag
type OnOrOffFlag struct {
	On  *OnFlag  `( @@`   //nolint
	Off *OffFlag `| @@ )` //nolint
}

// noinspection GoVetStructTag
type YesFlag struct {
	Dummy struct{} `("y"|"yes"|"true"|"1")` //nolint
}

// noinspection GoVetStructTag
type NoFlag struct {
	Dummy struct{} `("n"|"no"|"false"|"0")` //nolint
}

// noinspection GoVetStructTag
type YesOrNoFlag struct {
	Yes *YesFlag `( @@`   //nolint
	No  *NoFlag  `| @@ )` //nolint
}

// noinspection GoVetStructTag
type MoveCmd struct {
	Cmd    struct{}     `"move"`   //nolint
	Target NodeSelector `@@`       //nolint
	X      int          `@Int`     //nolint
	Y      int          `@Int`     //nolint
	Z      *int         `[ @Int ]` //nolint
}

// noinspection GoVetStructTag
type NodesCmd struct {
	Cmd struct{} `"nodes"` //nolint
}

// noinspection GoVetStructTag
type PartitionsCmd struct {
	Cmd struct{} `( "partitions" | "pts")` //nolint
}

// noinspection GoVetStructTag
type PingsCmd struct {
	Cmd struct{} `"pings"` //nolint
}

// noinspection GoVetStructTag
type JoinsCmd struct {
	Cmd struct{} `"joins"` //nolint
}

// noinspection GoVetStructTag
type CountersCmd struct {
	Cmd struct{} `"counters"` //nolint
}

// noinspection GoVetStructTag
type PlrCmd struct {
	Cmd struct{} `"plr"`             //nolint
	Val *float64 `[ (@Int|@Float) ]` //nolint
}

// noinspection GoVetStructTag
type RadioModelCmd struct {
	Cmd   struct{} `"radiomodel"`    //nolint
	Model string   `[(@Ident|@Int)]` //nolint
}

// noinspection GoVetStructTag
type RadioParamCmd struct {
	Cmd   struct{} `"radioparam"`      //nolint
	Param string   `[@Ident]`          //nolint
	Sign  string   `[@("-"|"+")]`      //nolint
	Val   *float64 `[ (@Int|@Float) ]` //nolint
}

// noinspection GoVetStructTag
type RfSimCmd struct {
	Cmd     struct{}      `"rfsim"`      //nolint
	Default *DefaultFlag  `[@@|`         //nolint
	Id      *NodeSelector `@@]`          //nolint
	Param   string        `[@Ident]`     //nolint
	Sign    string        `[@("-"|"+")]` //nolint
	Val     *int          `[ @Int ]`     //nolint
}

// noinspection GoVetStructTag
type LogLevelCmd struct {
	Cmd   struct{} `"log"`                                                                        //nolint
	Level string   `[@( "micro"|"trace"|"debug"|"info"|"warn"|"error"|"D"|"I"|"N"|"W"|"C"|"E" )]` //nolint
}

// noinspection GoVetStructTag
type WatchCmd struct {
	Cmd     struct{}       `"watch"`                                                                                             //nolint
	Default string         `[ @("default"|"def") ]`                                                                              //nolint
	Nodes   []NodeSelector `[ ( @@ )+ ]`                                                                                         //nolint
	Level   string         `[@( "trace"|"debug"|"info"|"note"|"warn"|"error"|"crit"|"off"|"none"|"T"|"D"|"I"|"N"|"W"|"E"|"C" )]` //nolint
}

// noinspection GoVetStructTag
type UnwatchCmd struct {
	Cmd   struct{}       `"unwatch"` //nolint
	Nodes []NodeSelector `( @@ )+`   //nolint
}

// noinspection GoVetStructTag
type HelpCmd struct {
	Cmd       struct{} `"help"`       //nolint
	HelpTopic string   `[ (@Ident) ]` //nolint
}

// noinspection GoVetStructTag
type KpiCmd struct {
	Cmd       struct{} `"kpi"`                        //nolint
	Operation string   `[ @("start"|"stop"|"save") ]` //nolint
	Filename  string   `[ @String ]`                  //nolint
}

// noinspection GoVetStructTag
type LoadCmd struct {
	Cmd      struct{} `"load"`  //nolint
	Filename string   `@String` //nolint
	Add      *AddFlag `[ @@ ]`  //nolint
}

// noinspection GoVetStructTag
type SaveCmd struct {
	Cmd       struct{} `"save"`                   //nolint
	Filename  string   `@String`                  //nolint
	Operation string   `[ @("all"|"topo"|"py") ]` //nolint
}

// noinspection GoVetStructTag
type SendCmd struct {
	Cmd        struct{}       `"send"`                        //nolint
	Protocol   string         `@("udp"|"tcp"|"coap"|"reset")` //nolint
	ProtoParam string         `[ @("non"|"con") ]?`           //nolint
	SrcId      NodeSelector   `@@`                            //nolint
	DstId      []NodeSelector `( @@ )*`                       //nolint
	AddrType   *AddrTypeFlag  `[ @@ ]`                        //nolint
	DataSize   *DataSizeFlag  `[ @@ ]`                        //nolint
}

// noinspection GoVetStructTag
type HostCmd struct {
	Cmd        struct{}     `"host"`                //nolint
	SubCmd     string       `@("add"|"del"|"list")` //nolint
	Hostname   string       `[ @String ]`           //nolint
	IpAddr     *Ipv6Address `[ @@ ]`                //nolint
	Port       uint16       `[ @Int ]`              //nolint
	PortMapped uint16       `[ @Int ]`              //nolint
}

// noinspection GoVetStructTag
type FailTimeParams struct {
	Dummy        struct{} `"ft"`          //nolint
	FailDuration float64  `(@Int|@Float)` //nolint
	FailInterval float64  `(@Int|@Float)` //nolint
}

// noinspection GoVetStructTag
type NoneFlag struct {
	Dummy struct{} `"none"` //nolint
}

// noinspection GoVetStructTag
type GuiFlag struct {
	Dummy struct{} `"gui"` //nolint
}

var (
	commandParser = participle.MustBuild(&Command{})
)

func parseBytes(b []byte, cmd *Command) error {
	err := commandParser.ParseBytes(b, cmd)
	return err
}
