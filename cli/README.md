# OTNS CLI Reference

The OTNS CLI exposes configuration and management APIs via a command line interface. Use the CLI to control OTNS or use it directly in additional application code. For example, the OTNS Python libraries use the CLI to manage simulations.

## OTNS command list

- [add](#add)
- [autogo](#autogo)
- [coaps](#coaps)
- [counters](#counters)
- [cv](#cv)
- [debug](#debug)
- [del](#del)
- [energy](#energy)
- [exe](#exe)
- [exit](#exit)
- [go](#go)
- [help](#help)
- [joins](#joins)
- [kpi](#kpi)
- [load](#load)
- [log](#log)
- [move](#move)
- [netinfo](#netinfo)
- [node](#node)
- [nodes](#nodes)
- [partitions](#partitions)
- [ping](#ping)
- [pings](#pings)
- [plr](#plr)
- [radio](#radio)
- [radiomodel](#radiomodel)
- [radioparam](#radioparam)
- [rfsim](#rfsim)
- [save](#save)
- [scan](#scan)
- [send](#send)
- [speed](#speed)
- [time](#time)
- [title](#title)
- [unwatch](#unwatch)
- [watch](#watch)
- [web](#web)

## OTNS CLI command reference

NOTE: the below sections including header and contents are automatically read and parsed during the build of OTNS, and used for providing inline help in the CLI via the `help` command. Specific syntax is followed in headers and in the triple backtick marked code segments in the Markdown source of this document. So, the help provided here or interactively in the OTNS CLI is exactly the same.

### add

Add a node to the simulation and get the node ID.

```shell
add <type> [x <x>] [y <y>] [rr <radio-range>] [id <node-id>] [restore] [exe <path>] [v11|v12|v13|v14]
```

The `<type>` can be `router`, `fed`, `med`, `sed`, `ssed`, `br` (Border Router), or `wifi` (for a Wi-Fi interferer node). Node ID can be specified using the `id` parameter, otherwise OTNS assigns the next available one. If the `restore` option is specified, the node restores its network configuration from persistent storage.

The (advanced) `exe` option can be used to specify a node executable for the new node; either a name only which is then located in the default search paths, or a full abs or rel pathname pointing to the executable to use.

The options `v11`, `v12`, `v13` and `v14` are a quick way to add a Thread v1.x node. This uses the binaries prebuilt for these nodes the `ot-rfsim` submodule, `ot-versions` directory. See [GUIDE.md](../GUIDE.md) for details on this.

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
simulation automatically runs with the current speed. If false (0), the simulation does not automatically run and requires an explicit `go` command to advance a particular time period. Use with a parameter to set the value.

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

```shell
coaps [enable]
```

Use `coaps enable` to enable collecting info on CoAP messages. CoAP message transmission and reception is detected through the special "coap" OTNS push events sent from the OT node binary to the simulator. Use `coaps` to show info of collected CoAP messages in YAML format.

```bash
> coaps enable
Done
```

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

### debug

A command that is optionally used to debug the CLI interaction with OTNS itself.

```shell
debug
debug echo "<string>"
debug fail
```

`debug` alone will do nothing except print "Done" to mark command completion. `debug echo` will echo (print) the given string on the command line. `debug fail` will deliberately raise an error, for testing purposes.

### del

Delete nodes by ID, IDs, ID ranges, or delete all nodes.

```shell
del <node-id> [<node-id> ...] | all | <node-id-1>-<node-id-2>
```

The node ID selection can use individual node(s), range(s) of nodes, "all" nodes, or a combination as shown in the examples below.

```bash
> del 1
Done
> del 3 4 5
Done
> del 6-19
Done
> del 20 38-42
Done
> del 43-53 87-97
Done
> del all
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

Use 'exe' without arguments to list the OpenThread (OT) executables, or shell scripts, that are preconfigured for each of the node types FTD (Full Thread Device), MTD (Minimal Thread Device) and BR (Thread Border Router). When a new node is created the executable currently in this list is used to start a node instance of that node type. The `br` (Border Router) node type is an FTD with some additional functions, and prefixes/routes, typical for a Thread 1.3 Border Router.

The line `Executables search path` lists the paths where the executable of that given name will be searched first. Finally, the lines `Detected ... path` lists the final detected path where the executable has been found. This is provided as a sanity check that the right executable has been detected for to-be-created OT nodes. If no explicit path is listed as detected path, it means that OTNS will try to launch the executable using the OS \$PATH.

```bash
> exe
ftd: ot-cli-ftd
mtd: ot-cli-mtd
br : ot-cli-ftd_br
Executables search path: [".", "./ot-rfsim/ot-versions", "./build/bin"]
Detected FTD path      : ./ot-rfsim/ot-versions/ot-cli-ftd
Detected MTD path      : ./ot-rfsim/ot-versions/ot-cli-mtd
Detected BR path       : ./ot-rfsim/ot-versions/ot-cli-ftd_br
Done
>
```

#### exe: Set OT executable for all node types

```shell
exe (default | v11 | v12 | v13 | v14)
```

Set all OpenThread (OT) executables, or shell scripts, for all node types to particular defaults. Value `default` will use the OTNS default executables: this is typically a recent OT development build. Values starting with `v` will use the pre-built binary of the specific indicated Thread version, e.g. `v12` denotes Thread v1.2.x.

NOTE: in the current commit of OTNS, `v14` equals the most recent OT development build. This may change in the future.

NOTE: the 'br' node type is currently not adapted to other versions.

```bash
> exe v11
ftd: ot-cli-ftd_v11
mtd: ot-cli-mtd_v11
br : ot-cli-ftd_br
Executables search path: [".", "./ot-rfsim/ot-versions", "./build/bin"]
Detected FTD path      : ./ot-rfsim/ot-versions/ot-cli-ftd_v11
Detected MTD path      : ./ot-rfsim/ot-versions/ot-cli-mtd_v11
Detected BR path       : ./ot-rfsim/ot-versions/ot-cli-ftd_br
Done
> exe default
ftd: ot-cli-ftd
mtd: ot-cli-mtd
br : ot-cli-ftd_br
Executables search path: [".", "./ot-rfsim/ot-versions", "./build/bin"]
Detected FTD path      : ./ot-rfsim/ot-versions/ot-cli-ftd
Detected MTD path      : ./ot-rfsim/ot-versions/ot-cli-mtd
Detected BR path       : ./ot-rfsim/ot-versions/ot-cli-ftd_br
Done
>
```

#### exe: Change OT executable for particular node type

```shell
exe ( ftd | mtd | br ) ["<path-or-filename-of-executable>"]
```

Change the OpenThread (OT) executable, or shell script, for a particular node types as provided in the first argument (ftd, mtd, or br). The path-or-filename is provided in the second argument and will replace the current default executable for that node type. If only the first argument is given, the current executable for this node type will be listed and no change is made. If only a filename is given, without full path, the executable will be located using the search paths listed under `Executables search path`.

Note that the default executable is used when normally adding a node using the GUI or a command such as `add router x 200 y 200` where the executable is not explictly specified. The "exe" argument of the "add" command will however override the default executable always, for example as in the command `add router x 200 y 200 exe "./my-override-ot-cli-ftd"` .

```bash
> exe ftd "./my-ot-cli-ftd"
ftd: ./my-ot-cli-ftd
Done
> exe br "./br-script.sh"
br : ./br-script.sh
Done
> exe
ftd: ./my-ot-cli-ftd
mtd: ot-cli-mtd
br : ./br-script.sh
Executables search path: [".", "./ot-rfsim/ot-versions", "./build/bin"]
Detected FTD path      : ./my-ot-cli-ftd
Detected MTD path      : ./ot-rfsim/ot-versions/ot-cli-mtd
Detected BR path       : ./br-script.sh
Done
> exe mtd
mtd: ot-cli-mtd
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

Simulate for a specified time in seconds or indefinitely (duration=`ever`). It is required in `-autogo=false` mode to advance the simulation. In `-autogo=true` mode, it can be optionally used to advance the simulation quickly by the given time. For example, in a paused simulation to quickly advance 64 us, 1 ms, 10 seconds, or an hour. The optional `speed` argument can be given to do the simulation at that speed e.g. to see the animations and log output better. The `duration` argument can optionally end with a time unit suffix: `us`, `ms`, `s`, `m`, or `h`.

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

### host

Add a simulated IP host for Thread nodes to communicate with.

```shell
host add "<hostname>" "<ipaddr>" <port> <maps-to-local-port>
host del "<hostname>" | "<ipaddr>"
host list
```

Any UDP/TCP packets sent to an off-mesh destination by a Thread node will first be routed over the mesh to a Thread Border Router (node type `br`). The BR will notify OTNS about such received packets. OTNS then looks up if any simulated IP host matches the packet's destination address/port and if so, it delivers the packet locally (localhost) to the mapped port number `<maps-to-local-port>`. This enables simulations with the behavior of Thread-external servers to be done fully locally, without causing any real network traffic.

Note: currently only IPv6 hosts are supported; IPv4 (via NAT64) may be added later.

### joins

Displays finished joiner sessions.

```bash
> joins
node=2    join=4.899s session=5.000s
Done
```

### kpi

Control the generation of Key Performance Indicators (KPIs) for a simulation.

```shell
kpi [ start | stop ]
kpi save [ "<filename>" ]
```

Use `kpi start` to start/restart KPI data recording at the current simulation time. KPI data recording will automatically stop at simulation exit. The KPI data is saved to a default JSON file `?_kpi.json` in the simulation's output folder. `kpi stop` will stop the KPI data recording at the current simulation time; this will also save results to the default file.

`kpi save` can be used at any moment, whether KPI data recording is active or not, to save the latest set of recorded KPI data to a file. If a filename is not provided, the default JSON file is used/overwritten. `kpi` without arguments inspects the state of the KPI collection.

```bash
> kpi start
Done
> go 3600
Done
> kpi
on
Done
> kpi save "kpi_scenario_1.json"
Done
> kpi stop
Done
> kpi
off
Done
>
```

NOTE: if any counters of nodes are reset using the OT CLI command `counters <type> reset` while KPI collection is ongoing, the results of KPI collection will become incorrect.

NOTE: KPI recording makes use of the [coaps](#coaps) command to enable CoAP message statistics. This may interfere with a user's ongoing CoAP message statistics collection, if any.

### load

Load a network topology from a YAML file.

```shell
load "<filename.yaml>" [add]
```

If the optional `add` parameter is used, the node IDs as defined in the YAML file will be incremented as needed to be higher than all current node IDs, and the new nodes will be added on top of nodes that are already there. All nodes in the YAML file can also be position-shifted prior to loading by changing the `pos-shift` parameter in the YAML file to a non-zero value. See [`save`](#save) for saving a network topology into a YAML file.

There are examples of the YAML format in the directory `./etc/mesh-topologies`.

### log

Inspect current OTNS log level, or set a new log level.

```shell
log [ debug | info | warn | error ]
```

The default log level is taken from the command line argument `-log`, or 'warn' is used if nothing is specified. Use 'debug' to see detailed log messages about OTNS internals. Log items display for OT nodes can be separately set using the [watch](#watch) CLI command.

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

Set default network and nodes info.

```shell
netinfo [version "<string>"] [commit "<string>"]
```

Sets information about OpenThread version and commit used for simulation nodes. This default information is then shown in the GUI, whenever a node is not selected. When a node is selected, the node-specific version/commit information will be used instead.

In the GUI, when the version/commit message is clicked, a web browser tab will be opened with the GitHub code for the particular OpenThread version/commit.

```bash
> netinfo version "Latest"
Done
> netinfo version "Latest" commit "a1816c1"
Done
> netinfo version "please select a node and then click this text to see the node's code." commit ""
Done
```

### node

Switch CLI context to specific node, or run command on node.

#### node: switch CLI context to specific node

```shell
node <node-id>
```

From within this new context, regular OT CLI commands (e.g. "help" or "state") can be used to directly interact with the node's CLI. The command 'exit' or 'node 0' can then be used again to exit the node context and return the CLI to global (OTNS) command context.

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

While in a node context, there is a shortcut to execute global-scope commands instead of node-specific OT CLI commands. This is adding the exclamation mark '!' character before the command. This is useful to avoid frequently changing between global and node contexts.

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

List current nodes in the simulation and key status information. The attribute 'failed' represents whether the node is currently in a simulated radio failure (true), or not (false).

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

where `<addr-type>` can be `linklocal`, `rloc`, `mleid`, `slaac`, or `any`.

NOTE: Sleepy End Devices (SEDs) typically don't respond to a ping request, while Synchronized Sleepy End Devices (SSEDs) do. A regular SED can be turned into a SSED by using the `csl period` command on the SED node.

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

Value `0` means no random packet loss, `0.5` means 50% of packets are randomly lost, while `1.0` means 100% of packets are lost.

NOTE: packets can be lost even if PLR is 0, for example if the RSSI of a frame is below the receiver's detection threshold, or if it has been interfered by another transmission. The PLR defines just an additional mechanism of purely random loss.

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
radio <node-id(s)> [on | off | ft <fail-duration> <fail-interval>]
```

Node IDs can be selected individually, as ranges (`<node-id-1>-<node-id-2`), as `all`, or combinations: see the [del](#del) command for details.

All `ft` parameters are in seconds (float). While a node's radio is off/failed, a red cross will be shown over the node in the Web GUI.

`ft 10 60` means the nodes' radio will be non-functional for a single window of 10 seconds, on average once every 60 seconds.

```bash
> radio 1 off
Done
> radio 1 on
Done
> radio 1 2 3 off
Done
> radio 1-5 on
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

Use without parameter to get the name of the currently used radiomodel (RF propagation model and radio chip characteristics applicable to all nodes). Or set the model to another model by providing the name or an alias of the model. Current models supported:

- `Ideal` (alias `I` or `1`) - has perfect radio reception within disc radius with constant good RSSI. CCA always finds the channel clear. There can be infinite parallel transmissions over the RF channel. If the OT node requests a transmission while one is already ongoing, it would be granted.
- `Ideal_Rssi` (alias `IR` or `2`) - has perfect radio reception within disc radius with decreasing RSSI over distance. CCA is like in the Ideal model.
- `MutualInterference` (alias `M` or `MI` or `3`) - has good to reasonable radio reception within disc radius with decreasing RSSI over distance. Outside the disc radius, there is still RF reception but of poor quality (Link Quality 0 or 1). CCA will consider nearby transmitting nodes, and will fail if energy is detected above CCA Threshold (which is configurable on the OT node on a per-node basis using the `ccathreshold` CLI command.) Concurrent transmissions will interfere and if the interferer signal is sufficiently strong, it will fail the radio frame transmission with FCS error. Only one transmission can occur at a time by a given node; if an additional transmission is requested by OT then the radio will report the ABORT failure. Also, CCA failure is reported if transmit is requested while the radio is receiving a frame.
- `MIDisc` (alias `MID` or `4`) - same as `MutualInterference` but limits transmissions/interference to a disc range equal to the node's radio-range parameter.
- `Outdoor` (alias `5`) - experimental outdoor propagation model. It assumes Line-of-Sight (LoS).

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

Use without the optional arguments to get a list of all current radiomodel parameters. These parameters apply to all nodes and links. Add the `param-name` to get only the value of that parameter. If both `param-name` and `new-value` are provided, the parameter value is set to `new-value`. It has to be a numeric value (float).

NOTE: How the parameter is used by the radiomodel may differ per radiomodel. Some parameters may not be used.

NOTE: To change radio hardware parameters of the simulated radio of a specific node, use the [rfsim](#rfsim) command.

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
TimeFadingSigmaMaxDb 4
MeanTimeFadingChange 120
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

### rfsim

Get or set parameters of a node's (OT-RFSIM node) simulated radio.

```shell
rfsim <node-id>
rfsim <node-id> <param-name>
rfsim <node-id> <param-name> <new-value>
```

Use with the `node-id` argument to get a list of all current OT-RFSIM radio parameters for that node. Add the `param-name` to get only the value of that parameter. If both `param-name` and `new-value` are provided, the parameter value is set to `new-value`. It has to be a numeric value (int).

In a physical radio platform, most of these parameters are typically fixed. In a simulation, these can be changed to explore different radios or different scenarios.

The following parameters are supported:

- `rxsens` - 802.15.4 receiver sensitivity (dBm), in the range -126 to 126. For correct radio operation, the receiver sensitivity MUST be kept lower than the current CCA ED threshold.
- `ccath` - 802.15.4 CCA Energy Detect (ED) threshold (dBm), in the range -126 to 126.
- `cslacc` - 802.15.4 Coordinated Sampled Listening (CSL) accuracy in ppm, range 0-255.
- `cslunc` - 802.15.4 CSL uncertainty in units of 10 microsec, range 0-255.
- `txintf` - for the `wifi` node type, sets the percentage of Wi-Fi traffic, range 0 to 100. Must not be >0 on other node types.

NOTE: To change global radio model parameters for all nodes, use the [radioparam](#radioparam) command.

```bash
> rfsim 1
rxsens               -100 (dBm)
ccath                -75 (dBm)
cslacc               20 (PPM)
cslunc               10 (10-us)
txintf               0 (%)
Done
> rfsim 1 cslacc 45
Done
> rfsim 1 cslacc
45
Done
> rfsim 1
rxsens               -100 (dBm)
ccath                -75 (dBm)
cslacc               45 (PPM)
cslunc               10 (10-us)
txintf               0 (%)
Done
>
```

### save

Save current network topology (nodes) into a YAML file.

```shell
save "<filename.yaml>"
```

Information about a node that will be saved in the file: type, position, and Thread version. Any internal state like 802.15.4 addresses, IP addresses, routing information, flash, counters etc. is not saved. The saved YAML file can be loaded again with [`load`](#load)

```bash
> save "./tmp/mynetwork.yaml"
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

### send

Send unicast and/or multicast data traffic between nodes, for testing purposes.

```shell
send udp|coap [non|con] <src-id> <dst-id(s)> [<addr-type>] [datasize <sz>]
send reset all
```

As node IDs for `<dst-id(s)>`, individual nodes, or ranges, or a combination, or "all" can be used, as shown in more detail in the [del](#del) command. If more than one node is selected in this way, a multicast message will be sent automatically. If it is one destination node, it will be unicast. In the present implementation, each subsequent multicast `send` message will be sent to a new IPv6 multicast group so that only the intended set of recipients will receive the message. This causes the number of multicast group memberships to grow over time, potentially. To reset all such memberships back to original state, `send reset all` can be used. This reset also stops any CoAP/UDP server active on all nodes and starts the numbering of multicast groups again at 1.

As protocol, `udp` or `coap` can be selected. For `coap`, also `non` (Non-Confirmable) or `con` (Confirmable) transmission can be chosen. If absent, `non` is assumed. For multicast, `non` is specified by RFC 7252 but for testing purposes also `con` can be used here. Note that CoAP responses are currently not generated for `non` (future addition may address this). Traffic protocols like tcp, tls, or coaps are currently not implemented. For ICMPv6 traffic see [ping](#ping).

The optional `<addr-type>` allows to specify the unicast address type to use (see [ping](#ping) for details). The optional `datasize` (or `ds`) argument sets the payload data size in bytes, between 0-~1220 for `udp` and a smaller range for `coap` of 0-~580 due to CLI line length limits.

Concurrent traffic can be generated by issuing multiple `traffic` commands from a Python script, or from the CLI while the simulation is paused or running with slow [speed](#speed).

```bash
> send udp 1 2
Done
Node<2>  32 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10002 0123456789ABCDEFGHIJKLMNOPQRSTUV
> send udp 2 1 rloc
Done
Node<1>  32 bytes from fdde:ad00:beef:0:0:ff:fe00:3400 10003 0123456789ABCDEFGHIJKLMNOPQRSTUV
> send udp 1 2 datasize 180
Done
Node<2>  180 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10021 0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrst
>
[...]
> send udp 1 2-5
Done
Node<2>  32 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10022 0123456789ABCDEFGHIJKLMNOPQRSTUV
Node<3>  32 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10022 0123456789ABCDEFGHIJKLMNOPQRSTUV
Node<4>  32 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10022 0123456789ABCDEFGHIJKLMNOPQRSTUV
Node<5>  32 bytes from fdde:ad00:beef:0:5f0e:224f:aa33:f5f2 10022 0123456789ABCDEFGHIJKLMNOPQRSTUV
> send coap 2 5
Done
Node<5>  coap request from fdde:ad00:beef:0:f5cc:a5:9fc:b417 POST with payload: 6c4648535159434755636e4d6664775653493669564b4166584744435762304d
>
```

For `coap`, message statistics/info/latency can be tracked by using [coaps](#coaps) as shown in the below example.

```bash
> coaps enable
Done
> send coap 2 3-5
Done
Node<3>  coap request from fdde:ad00:beef:0:f5cc:a5:9fc:b417 POST with payload: 64757479795a374a7569313654414947636a5a6f344e54364b667051756d3254
Node<4>  coap request from fdde:ad00:beef:0:f5cc:a5:9fc:b417 POST with payload: 64757479795a374a7569313654414947636a5a6f344e54364b667051756d3254
Node<5>  coap request from fdde:ad00:beef:0:f5cc:a5:9fc:b417 POST with payload: 64757479795a374a7569313654414947636a5a6f344e54364b667051756d3254
> coaps
- {time: 1151062448, src: 2, id: 65467, type: 1, code: 2, uri: t, dst_addr: 'ff13:0:0:0:0:0:deed:6', dst_port: 5683,
   receivers: [
   {time: 1151068024, dst: 3, src_addr: 'fdde:ad00:beef:0:f5cc:a5:9fc:b417', src_port: 5683},
   {time: 1151068024, dst: 4, src_addr: 'fdde:ad00:beef:0:f5cc:a5:9fc:b417', src_port: 5683},
   {time: 1151076032, dst: 5, src_addr: 'fdde:ad00:beef:0:f5cc:a5:9fc:b417', src_port: 5683}]}
Done
>
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

The below shows an example of a paused simulation, that is advanced by exactly 100 microseconds using the `go` command.

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

Disable detailed logging (watching) for selected node(s).

```shell
unwatch all | <node-id(s)>
```

With node number parameter(s), it disables the watch status for one or more nodes. A range of nodes `<node-id-1>-<node-id-2` can also be given as shown in the [del](#del) command. Using the `all` parameter will disable the watch status for all nodes. See [watch](#watch) for details.

### watch

Configure detailed logging (watching) for selected node(s).

```shell
watch [<node-id(s)>]
watch <node-id(s)> [<LogLevel>]
watch default [<LogLevel>]
```

The log entries of nodes are displayed in the CLI. This can be useful for interactive debugging or inspection of a node's behavior. The watch function is mostly independent of the OT node's log file: entries that are not displayed, are typically still written to the OT node log file.

Node IDs to watch can be selected using individual node(s), range(s) of nodes, "all" nodes, or a combination. See the [del](#del) command for details on node selection.

- To see all nodes currently being watched, use "watch" without parameters.
- Any log entries that are displayed due to watch, are also written to the OT node log file (if active).
- With the below examples, watching a node will only display OT stack log messages from level Info (I) or up. To see Debug (D) messages, or only Warn (W) or Error/Critical (C) messages, use the `<LogLevel>` parameter as shown further down.

```bash
> watch 1
Done
> watch 3 5 6
Done
> watch
1 3 5 6
> unwatch 1-5
Done
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
cause OT stack log messages from indicated log level, or higher (more important), to be shown. By default, only the Info (I) level or up is shown. Setting the level can be useful for interactive debugging or inspection of a node's behavior including the operation of its simulated radio.

- Valid long-form LogLevels are "trace", "debug", "info", "note", "warn", "error", or "crit" (same as "error").
- Valid short-form LogLevels that are named like in the OT stack log output are "D", "I", "N", "W", "C"; with additionally "T" for trace or "E" for error/critical available.
- This command can also be used to change the LogLevel of one or more nodes being already watched, to a new  
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

Using `all` enables the watch status for all present nodes.

```bash
> watch all
Done
> watch all debug
Done
>
```

#### watch default \[\<LogLevel\>\]

Using `default` sets the default watch status and `LogLevel` for all newly created nodes.

- Use `off` for disabling the default watch on new nodes. This also sets the watch LogLevel to `default` in case a manual watch is set later on without specifying a LogLevel parameter.
- Omit the `LogLevel` argument to see current default.

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

The optional `TabName` indicates which OTNS tab to open:

- if not provided, or "main", the default main simulation window will open.
- if "stats", the stats-viewer will be opened.
- if "energy", the energy-viewer will be opened.

NOTE: multiple web browser tabs/windows of the same type may be opened for the same simulation.

```bash
> web
Done
> web energy
Done
> web stats
Done
```
