# OTNS CLI Reference

The OTNS CLI exposes configuration and management APIs via a
command line interface. Use the CLI to control OTNS or use it
directly in additional application code. For example, the OTNS
Python libraries use the CLI to manage simulations.

## OTNS command list

* [add](#add)
* [autogo](#autogo)
* [coaps](#coaps)
* [counters](#counters)
* [cv](#cv)
* [del](#del)
* [energy](#energy)
* [exe](#exe)
* [exit](#exit)
* [go](#go)
* [help](#help)
* [joins](#joins)
* [log](#log)
* [move](#move)
* [netinfo](#netinfo)
* [node](#node)
* [nodes](#nodes)
* [partitions](#partitions)
* [ping](#ping)
* [pings](#pings)
* [plr](#plr)
* [radio](#radio)
* [radiomodel](#radiomodel)
* [radioparam](#radioparam)
* [rxsens](#rxsens)
* [scan](#scan)
* [speed](#speed)
* [time](#time)
* [title](#title)
* [unwatch](#unwatch)
* [watch](#watch)
* [web](#web)

## OTNS CLI command reference


### add

Add a node to the simulation and get the node ID.

```shell
add <type> [x <x>] [y <y>] [rr <radio-range>] [id <node-id>] [restore] [exe <path>] [v11|v12|v13|v131]
```

Node ID can be specified, otherwise OTNS assigns the next available 
one. If the `restore` option is specified, the node restores its network configuration from persistent storage.

The (advanced) `exe` option can be used to specify a node executable for the new node; either a name only which is 
then located in the default search paths, or a full abs or rel pathname pointing to the executable to use.

The options `v11`, `v12`, `v13` and `v131` are a quick way to add a Thread v1.x node. This uses the binaries 
prebuilt for these nodes the `ot-rfsim` submodule, `ot-versions` directory. 
See [GUIDE.md](../GUIDE.md) for details on this.

```bash
> add router
1
Done
> add fed x 100 y 100
2
Done
> add med x 100 y 200 rr 200
3
Done
> add sed x 200 y 200 rr 400
4
Done
> add sed x 200 y 200 restore
5
Done
> add fed x 200 y 200 id 25
25
Done
> add router v11
6
Done
> add router exe "ot-cli-ftd_nologs"
7
Done
> add router exe "/home/user/my/path/to/ot-cli-ftd"
8
Done
```

### autogo

Get or set the simulation's `autogo` property. 

```shell
autogo [ 1 | 0 ]
```

Use without parameter to get the property's value. If true (1), then autogo is enabled and the  
simulation automatically runs with the current speed. If false (0), the simulation does not automatically run and 
requires an explicit `go` command to advance a particular time period. Use with a parameter to set the value.

```bash
> autogo
1
Done
> autogo 0
Done
> autogo
0
Done
>
```

### coaps

Enable collecting info on CoAP messages or show collected info in YAML.

#### coaps enable

Enable collecting info of CoAP messages. CoAP message transmission and reception is detected through the special 
"coap" OTNS push events sent from the OT node binary to the simulator.

```bash
> coaps enable
Done
```

#### coaps

Show info of collected CoAP messages in YAML format.

```bash
> coaps
- {time: 57019000, src: 2, id: 25421, type: 0, code: 2, uri: a/as, dst_addr: 'fdde:ad00:beef:0:0:ff:fe00:f000', dst_port: 61631, receivers: [{time: 57019961, dst: 1, src_addr: 'fdde:ad00:beef:0:0:ff:fe00:f001', src_port: 61631}]}
- {time: 57019961, src: 1, id: 25421, type: 2, code: 68, dst_addr: 'fdde:ad00:beef:0:0:ff:fe00:f001', dst_port: 61631, receivers: [{time: 57021242, dst: 2, src_addr: 'fdde:ad00:beef:0:0:ff:fe00:f000', src_port: 61631}]}
Done
```

### counters

Display runtime counters of the simulation.

```bash
> counters
AlarmEvents                              95983
RadioEvents                              1674
StatusPushEvents                         47
UartWriteEvents                          182322
CollisionEvents                          0
DispatchByExtAddrSucc                    239
DispatchByExtAddrFail                    0
DispatchByShortAddrSucc                  188
DispatchByShortAddrFail                  0
DispatchAllInRange                       290
Done
```

### cv

Configure visualization options.

```shell
cv [<option> on|off] ...
```

Visualization Options:
- `bro`: broadcast message
- `uni`: unicast message
- `ack`: ACK message
- `rtb`: router table
- `ctb`: child table

```bash
> cv
bro=on
uni=on
ack=off
rtb=on
ctb=on
Done
> cv bro off
bro=off
uni=on
ack=off
rtb=on
ctb=on
Done
> cv bro on uni on ack on rtb on ctb on
bro=on
uni=on
ack=on
rtb=on
ctb=on
Done
```

### del

Delete nodes by ID.

```shell
del <node-id> [<node-id> ...]
```

```bash
> del 1
Done
> del 1 2 3
Done
``` 

### energy
To be documented (TODO) - saves energy use information of nodes to file.

```shell
energy [save] "<filename>"
```

### exe

List, or change, OT versions/executables used per node type.

#### exe: list OT executables used per node type 

Use 'exe' without arguments to list the OpenThread (OT) executables, or shell scripts, that are preconfigured for each 
of the node types
FTD (Full Thread Device), MTD (Minimal Thread Device) and BR (Thread Border Router). When a new node is created the
executable currently in this list is used to start a node instance of the respective node type.

NOTE: the `br` (Border Router) node type is currently not supported (functionality is under construction).

The line `Executables search path` lists the paths where the executable of that given name will be searched first.
Finally, the line `Detected FTD path` lists the final detected path where the `ftd` executable has been found. This 
is provided as a sanity check (for the FTD case only) that the right executable has been detected for future OT nodes.

```bash
> exe
ftd: ot-cli-ftd
mtd: ot-cli-ftd
br : ot-br.sh
Executables search path: [".", "./ot-rfsim/ot-versions"]
Detected FTD path      : ./ot-rfsim/ot-versions/ot-cli-ftd
Done
>  
```

#### exe: Set OT executable for all node types

```shell
exe (default | v11 | v12 | v13)
```

Set all OpenThread (OT) executables, or shell scripts, for all node types to particular defaults. Value `default` will 
use the OTNS default executables which is OpenThread as built by the user and placed in the `.` directory from 
where the simulator is run. Values starting with `v` will use the pre-built binary of the specific indicated Thread 
version, e.g. `v12` denotes Thread v1.2.x. 

```bash
> exe default
ftd: ./ot-cli-ftd
mtd: ./ot-cli-ftd
br : ./otbr-sim.sh
Done
>
```

#### exe: Change OT executable for particular node type

```shell
exe ( ftd | mtd | br ) ["<path-to-executable>"]
```

Change the OpenThread (OT) executable, or shell script, for a particular node types as provided in the first 
argument (ftd, mtd, or br). The path-to-executable is provided in the second argument and will replace the current 
default executable for that node type. If only the first argument is given, the current executable for this node 
type will be listed.

Note that the default executable is used when normally adding a node using the GUI or a command such as 
```add router x 200 y 200``` where the executable is not explictly specified. The "exe" argument of the "add" command 
will however override the default executable always, for example as in the command 
```add router x 200 y 200 exe "./my-override-ot-cli-ftd"``` .

```bash
> exe ftd "./my-ot-cli-ftd"
Done
> exe br "./br-script.sh"
Done
> exe
ftd: ./my-ot-cli-ftd
mtd: ./ot-cli-ftd
br : ./br-script.sh
Executables search path: [".", "./ot-rfsim/ot-versions"]
Detected FTD path      : ./my-ot-cli-ftd
Done
> exe mtd
mtd: ./ot-cli-ftd
Done
```

### exit

Exit OTNS, if in main context (no node selected). If in a node context (node selected on the CLI), exits the node context.

```bash
node 3> exit
Done
> exit
Done
<EOF>
```

### go

Simulate for a specified time.

```shell
go <duration> [speed <particular-speed>]
 ```

Simulate for a specified time in seconds or indefinitely (duration=`ever`). It is required in `-autogo=false` mode to
advance the simulation. In `-autogo=true` mode, it can be optionally used to advance the simulation quickly 
by the given time. For example, in a paused simulation to quickly advance 64 us, 1 ms, 10 seconds, or an hour.
The optional `speed` argument can be given to do the simulation at that speed e.g. to see the animations 
and log output better. 
The `duration` argument can optionally end with a time unit suffix: 
`us`, `ms`, `s`, `m`, or `h`.

```bash
> go 1
Done
> go 10
Done
> go 0.003
Done
> go 5 speed 0.1
Done
> go 64us
Done
> go 20m
Done
> go ever
<NEVER FINISHES>
```

### help

Show help text for specific, or all, OTNS CLI commands.

```shell
help [ <command> ]
```

### joins

Displays finished joiner sessions.

```bash
> joins
node=2    join=4.899s session=5.000s
Done
```

### log
Inspect current log level, or set a new log level.

```shell
log [ debug | info | warn | error ]
```

The default is taken from the command line argument,
or 'warn' if nothing is specified yet. Use 'debug' to see detailed log messages.
Log level 'info' or lower is needed to see any OT node's stack + application log messages.

```bash
> log
warn
Done
> log debug
Done
> log
debug
Done
```

### move

Move a node to the 2D target position (x,y).

```shell
move <node-id> <x> <y>
```

```bash
> move 1 200 300
Done
```

### netinfo

Set network nodes info.

```shell
netinfo [version "<string>"] [commit "<string>"] [real y|n]
```

Sets information about OpenThread version and Commit used for simulation nodes, as well as whether nodes are real 
or simulated. This information is then shown in the GUI. This temporarily overrides the information that OTNS already sets by 
default, whenever a node is added or deleted. When a node is added or deleted again, OTNS will automatically set the 
version and commit information again based on current nodes in the simulation.

In the GUI, when the version/commit message is clicked, a web browser tab will be opened with the Github code for 
the particular version/commit. This only works currently if there's a single version/commit shared by all nodes in the 
simulation.

NOTE: setting `real` enabled (y) will disable some of the GUI controls of the simulation, such as speed/pause.

```bash
> netinfo version "Latest"
Done
> netinfo version "Latest" commit "b49ee08"
Done
> netinfo real y
Done
```

### node

Switch CLI context to specific node, or run command on node.

#### node: switch CLI context to specific node

```shell
node <node-id>
```

From within this new context, regular OT CLI commands (e.g. "help" or "state") can be 
used to directly interact with the node's CLI. The command 'exit' or 'node 0' can then be used again to exit the node 
context and return the CLI to global (OTNS) command context.

```bash
> node 3
Done
node 3> state
router
Done
node 3> exit
Done
>
```

While in a node context, there is a shortcut to execute global-scope commands instead of node-specific OT CLI 
commands. This is adding the exclamation mark '!' character before the command. This is useful to avoid frequently 
changing between global and node contexts.

```bash
> node 2
Done
node 2> state
leader
Done
node 2> !nodes
id=1    extaddr=da7bb222abc9c806        rloc16=a400     x=149   y=1176  state=router    failed=false
id=2    extaddr=0a5b1645b5dfdd73        rloc16=1c00     x=163   y=1175  state=leader    failed=false
id=3    extaddr=0638ac1ab9072dea        rloc16=d800     x=170   y=1176  state=router    failed=false
Done
node 2>  
```

#### node: run CLI command on specific node

Run an OpenThread CLI command on a specific node.

```shell
node <node-id> "<command>"
```

```bash
> node 1 "state"
leader
Done
```

### nodes

List current nodes in the simulation and key status information. The attribute 'failed' represents whether the 
node is currently in a simulated radio failure (true), or not (false).

```bash
> nodes
id=1	extaddr=62cfcf3c5556ac7c	rloc16=c000	x=200	y=300	failed=false
id=2	extaddr=6a7d9d31e3511147	rloc16=3000	x=278	y=708	failed=false
id=3	extaddr=266db93fad653782	rloc16=2800	x=207	y=666	failed=false
Done
```

### partitions

List Thread partitions in the current simulation.

```
> partitions
partition=4683661d	nodes=4,1,3
partition=7cb22d3b	nodes=2
Done
> pts
partition=7cb22d3b	nodes=2
partition=4683661d	nodes=1,3,4
Done
```

### ping

Request ping from source node to a destination (other node, or IPv6 address).

```shell
ping <src-id> <dst-id> [<addr-type>] [datasize <sz>] [count <cnt>] [interval <intval>] [hoplimit <hoplim>]
ping <src-id> "<dst-addr>" [datasize <sz>] [count <cnt>] [interval <intval>] [hoplimit <hoplim>]
```

NOTE: Sleepy End Devices (SEDs) typically don't respond to a ping request, while Synchronized Sleepy End Devices
(SSEDs) do. A regular SED can be turned into a SSED by using the `csl period` command on the SED node.

```bash
> ping 1 2 
Done
> ping 1 2 rloc
Done
> ping 1 2 mleid
Done
> ping 1 "fdde:ad00:beef:0:31d6:8873:f685:9c40"
Done
> ping 1 2 datasize 10 count 3 interval 1 hoplimit 10
Done
```

### pings

Display finished ping sessions. 

```bash
> ping 1 2 count 3
Done
> pings
node=1    dst=fdde:ad00:beef:0:31d6:8873:f685:9c40     datasize=4   delay=0.322ms
node=1    dst=fdde:ad00:beef:0:31d6:8873:f685:9c40     datasize=4   delay=2.242ms
node=1    dst=fdde:ad00:beef:0:31d6:8873:f685:9c40     datasize=4   delay=1.282ms
Done
```

### plr

Get or set global packet loss ratio (PLR) for the simulation. 

```shell
plr [<loss-value>]
```

Value `0` means no random packet loss, `0.5` means 
50% of packets are randomly lost, while `1.0` means 100% of packets are lost.

Note that packets can be lost even if PLR is 0, for example if the RSSI of a frame is below the receiver's 
detection threshold, or if it has been interfered by another transmission. The PLR defines just an additional 
mechanism of purely random loss.

```bash
> plr 
0
Done
> plr 0.5
0.5
Done
```

### pts

Synonym for `partitions` command. See [partitions](#partitions).

### radio

Set the node radio on/off/fail time parameters.

```shell
radio <node-id> [<node-id> ...] [on | off | ft <fail-duration> <fail-interval>]
```

All `ft` parameters are in seconds (float).
While a node's radio is off/failed, a red cross will be shown over the node in the Web GUI.

`ft 10 60` means the nodes' radio will be non-functional for a single window of 10 seconds, on average once every
60 seconds.

```bash
> radio 1 off
Done
> radio 1 on
Done
> radio 1 2 3 off
Done
> radio 1 2 3 on
Done
> radio 1 2 3 ft 10 60
Done
> radio 3 ft 0.364 10.0
Done
```

### radiomodel

Get or set current radiomodel (RF propagation).

```shell
radiomodel [<modelName>]
```

Use without parameter to get the name of the currently used radiomodel (RF propagation model and radio chip 
characteristics applicable to all nodes). Or set the model to another model by providing the name or an alias of 
the model. Current models supported:

* `Ideal` (alias `I` or `1`) - has perfect radio reception within disc radius with constant good RSSI. CCA always finds the channel clear. 
  There can be infinite parallel transmissions over the RF channel. If the OT node would request a transmission while one 
 is already ongoing, it would be granted.
* `Ideal_Rssi` (alias `IR` or `2`) - has perfect radio reception within disc radius with decreasing RSSI over distance. CCA is like
  in the Ideal model.
* `MutualInterference` (alias `M` or `MI` or `3`) - has good to reasonable radio reception within disc radius with decreasing 
 RSSI over distance. Outside the disc radius, there is still RF reception but of poor quality (Link Quality 0 or 1). CCA 
 will consider nearby transmitting nodes, and will fail if energy is detected above CCA Threshold (which is configurable 
 on the OT node on a per-node basis using the `ccathreshold` CLI command.)  Concurrent transmissions will interfere and 
 if the interferer signal is sufficiently strong, it will fail the radio frame transmission with FCS error. Only one 
 transmission can occur at a time by a given node; if an additional transmission is requested by OT then the radio will 
 report the ABORT failure. Also CCA failure is reported if transmit is requested while the radio is receiving a frame.
* `MIDisc` (alias `MID` or `4`) - same as `MutualInterference` but limits transmissions/interference to a disc range 
 equal to the node's radio-range parameter.
* `Outdoor` (alias `5`) - experimental outdoor propagation model. It assumes Line-of-Sight (LoS).

```bash
> radiomodel
Ideal
Done
> radiomodel MutualInterference
MutualInterference
Done
> radiomodel
MutualInterference
Done
> radiomodel 1
Ideal
Done
> radiomodel IR
Ideal_Rssi
Done
>
```

### radioparam

Get or set parameters of the current radiomodel.

```shell
radioparam [param-name] [new-value]
```

Use without the optional arguments to get a list of all current 
radiomodel parameters. Add the `param-name` to get only the value of that parameter. If both `param-name` and `new-value` 
are provided, the parameter value is set to `new-value`. It has to be a numeric value (float).

Note: How the parameter is used by the radiomodel may differ per radiomodel. Some parameters may not be used.

```bash
> radioparam
MeterPerUnit         0.1
IsDiscLimit          0
RssiMinDbm           -126
RssiMaxDbm           126
ExponentDb           17.3
FixedLossDb          40
NlosExponentDb       38.3
NlosFixedLossDb      26.77
NoiseFloorDbm        -95
SnrMinThresholdDb    -4
ShadowFadingSigmaDb  8.03
Done
> radioparam MeterPerUnit
0.1
Done
> radioparam MeterPerUnit 0.5
Done
> radioparam MeterPerUnit
0.5
Done
> 
```

### rxsens

Get or set receiver sensitivity (dBm) for a node.

```shell
rxsens <node-id> [sensitivity-value]
```

Values range from -126 to 126. For correct radio 
operation, the receiver sensitivity MUST be kept lower than the current CCA ED threshold. The latter can be set 
using the OT node CLI command `ccathreshold`.

Note: this command may get replaced in the future by a generic command for setting OT node radio parameters.

```bash
> add router
1
Done
> rxsens 1
-100 dBm
Done
> rxsens 1 -85
Done
> rxsens 1
-85 dBm
Done
>  
```

### scan

Perform a network scan by the indicated node.

```shell
scan <node-id>
```

This simply calls the `scan` CLI command on the indicated node and outputs results.

```bash
> scan 2
| J | Network Name     | Extended PAN     | PAN  | MAC Address      | Ch | dBm | LQI |
+---+------------------+------------------+------+------------------+----+-----+-----+
| 0 | OpenThread       | dead00beef00cafe | face | 66c6bfef495534af | 11 | -20 |   0 |
Done
```

### speed

Get or set the simulation speed.

```shell
speed [ <speed> | max | inf ] 
```

```bash
> speed
8
Done
> speed 10
Done
> speed
10
Done
```

Use `inf` or `max` to set maximum simulation speed.

```bash
> speed max
Done
> speed
1e+06
Done
> speed inf
Done
> speed
1e+06
Done
```

### time

Display current simulation time in us. 

The below shows an example of a paused simulation, that is advanced by exactly 100 microseconds using the `go` 
command.

```bash
> time
312560
Done
> go 100us
Done
> time
312660
Done
> 
```

### title

Set simulation title. This is displayed in the GUI.

```shell
title "<string>" [x <int>] [y <int>] [fs <font-size-integer>]
```

```bash
> title "Example"
Done
> title "Another Example" x 100 y 200
Done
> title "Example with font size 30" fs 30
Done
```

### unwatch

Disable detailed logging for selected node(s).

```shell
unwatch all | <node-id> [<node-id> ...]
```

With node number parameter(s), it disables the watch status for one or more nodes. 
Using the `all` parameter will disable the watch status for all nodes. See [watch](watch) for details.

### watch

Enable additional, detailed logging on selected node(s).

```shell
watch [<node-id>] [<node-id> ...]
watch <node-id> [<node-id> ...] [<LogLevel>]
watch all [<LogLevel>]
watch default [<LogLevel>]
```

This can be useful for interactive debugging or 
inspection of a node's behavior. 

* To see all nodes currently being watched, use "watch" without parameters.
* With the below examples, watching a node will only display OT stack log messages from level Info (I) or up. To see Debug (D) 
  messages, or only Warn (W) or Error/Critical (C) messages, use the `<LogLevel>` parameter as shown further down.

```bash
> watch 1
Done
> watch 3 5 6
Done
> watch
1 3 5 6
> unwatch 1 3 5
> watch
6
> watch 3 5
Done
> unwatch all
Done
> watch

Done
>
```

#### watch with \<LogLevel\>
An advanced use of the watch command uses the LogLevel option. Adding the `<LogLevel>` optional parameter will  
cause OT stack log messages from indicated log level, or higher (more important), to be shown. By default, only the 
Info (I) level or up is shown. Setting the level can be useful for interactive debugging or inspection of a node's behavior 
including the operation of its simulated radio.

* Valid long-form LogLevels are "trace", "debug", "info", "note", "warn", "error", or "crit" (same as "error").
* Valid short-form LogLevels that are named like in the OT stack log output are "D", "I", "N", "W", "C"; with 
 additionally "T" for trace or "E" for error/critical available.
* This command can also be used to change the LogLevel of one or more nodes being already watched, to a new  
 LogLevel.

```bash
> watch 1 debug
Done
> watch 3 5 6 warn
Done
> watch
1 3 5 6
> watch 3 5 trace
Done
> watch 3 5 6 D
Done
> watch 3 5 6 I
Done
>
```

#### watch all \[\<LogLevel\>\]

Using `all` enables the watch status for all nodes. 

```bash
> watch all
Done
> watch all debug
Done
> 
```

#### watch default \[\<LogLevel\>\]

Using `default` sets the default watch status and `LogLevel` for all newly created nodes. 

* Use `off` for disabling the default watch on new nodes. This also sets the watch LogLevel to `default` in case a 
 manual watch is set later on without specifying a LogLevel parameter.
* Omit the `LogLevel` argument to see current default.

```bash
> watch default debug
Done
> watch default
debug
Done
> watch default off
Done
> 
```

### web

Open a web browser (tab) for visualization. 

```shell
web [ <TabName> ]
```

The optional `TabName` indicates which OTNS2 tab to open: 

* if not provided, or "main", the default main simulation window will open.
* if "stats", the stats-viewer will be opened.
* if "energy", the energy-viewer will be opened.

NOTE: multiple web browser tabs/windows of the same type may be opened for the same simulation.

```bash
> web
Done
> web energy
Done
> web stats
Done
```
