# OTNS CLI Reference

The OTNS CLI exposes configuration and management APIs via a
command line interface. Use the CLI to control OTNS or use it
directly in additional application code. For example, the OTNS
Python libraries use the CLI to manage simulations.

## OTNS command list

* [add](#add-type-x-x-y-y-rr-radio-range-id-node-id-restore)
* [coaps](#coaps-enable)
* [counters](#counters)
* [cv](#cv-option-onoff-)
* [del](#del-node-id-node-id-)
* [energy](#energy-save--filename-)
* [exe](#exe)
* [exit](#exit)
* [go](#go-duration-speed-particular-speed)
* [help](#help)
* [joins](#joins)
* [log](#log-level)
* [move](#move-node-id-x-y)
* [netinfo](#netinfo-version-string-commit-string-real-yn)
* [node](#node-node-id)
* [node](#node-node-id-command)
* [nodes](#nodes)
* [partitions (pts)](#partitions-pts)
* [ping](#ping-src-id-dst-id-addr-type--dst-addr--datasize-datasize-count-count-interval-interval-hoplimit-hoplimit)
* [pings](#pings)
* [plr](#plr)
* [radio](#radio-node-id-node-id--on--off--ft-fail-duration-fail-interval)
* [radiomodel](#radiomodel-modelname)
* [scan](#scan-node-id)
* [speed](#speed)
* [title](#title-string)
* [unwatch](#unwatch-node-id-node-id-)
* [watch](#watch-node-id-node-id-)
* [web](#web)

## OTNS command reference


### add \<type\> \[x \<x\>\] \[y \<y\>\] \[rr \<radio-range\>\] \[id \<node-id\>\] \[restore\] \[exe \<path\>\] \[v11 | v12\]

Add a node to the simulation and get the node ID. Node ID can be specified, otherwise OTNS assigns the next available 
one.

If `restore` option is specified, the node restores its network configuration from persistent storage.

The (advanced) `exe` option can be used to specify a node executable for the new node; however the `exe` command is 
better used for this.
The options `v11` and `v12` are a quick way to add a legacy Thread v1.1 or v1.2 node. This only works if the binaries 
for these nodes have been built using the build scripts in the `ot-rfsim` submodule.

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
```

### coaps enable

Enable collecting info of CoAP messages.

```
> coaps enable
Done
```

### coaps

Show info of collected CoAP messages in yaml format.

```
> coaps
- {time: 57019000, src: 2, id: 25421, type: 0, code: 2, uri: a/as, dst_addr: 'fdde:ad00:beef:0:0:ff:fe00:f000', dst_port: 61631, receivers: [{time: 57019961, dst: 1, src_addr: 'fdde:ad00:beef:0:0:ff:fe00:f001', src_port: 61631}]}
- {time: 57019961, src: 1, id: 25421, type: 2, code: 68, dst_addr: 'fdde:ad00:beef:0:0:ff:fe00:f001', dst_port: 61631, receivers: [{time: 57021242, dst: 2, src_addr: 'fdde:ad00:beef:0:0:ff:fe00:f000', src_port: 61631}]}
Done
```

### counters

Display runtime counters.

```bash
> counters
AlarmEvents                              95983
RadioEvents                              1674
StatusPushEvents                         47
DispatchByExtAddrSucc                    239
DispatchByExtAddrFail                    0
DispatchByShortAddrSucc                  188
DispatchByShortAddrFail                  0
DispatchAllInRange                       0
Done
```

### cv \[\<option\> on|off\] ...

Configure visualization options.

Visualization Options:
- bro: broadcast message
- uni: unicast message
- ack: ACK message
- rtb: router table
- ctb: child table

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

### del \<node-id\> \[<node-id> ...\]

Delete nodes by ID.

```bash
> del 1
Done
> del 1 2 3
Done
``` 

### energy \[save\] "\<filename\>"
To be documented - saves energy use information of nodes to file.

### exit

Exit OTNS, if in main context (no node selected). If in a node context, exits the node context.

```bash
node 3> exit
Done
> exit
Done
<EOF>
```

### exe

Use 'exe' without arguments to list the OpenThread (OT) executables, or shell scripts, that are preconfigured for each 
of the node types
FTD (Full Thread Device), MTD (Minimal Thread Device) and BR (Thread Border Router). When a new node is created the
executable currently in this list is used to start a node instance of the respective node type.

```bash
> exe
ftd: ./ot-cli-ftd
mtd: ./ot-cli-ftd
br : ./otbr-sim.sh
Done
>  
```

### exe (default | v11 | v12 )

Set all OpenThread (OT) executables, or shell scripts, for all node types to particular defaults. Value `default` will 
use the OT-NS default executables which is OpenThread as built by the user and placed in the `.` directory from 
where the simulator is run. Values starting with `v1` will use the pre-built binary of the specific indicated Thread 
version, e.g. `v12` denotes Thread v1.2.x. 

```bash
> exe default
ftd: ./ot-cli-ftd
mtd: ./ot-cli-ftd
br : ./otbr-sim.sh
Done
>
```

### exe \( ftd | mtd | br \) \["\<path-to-executable\>"\]

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
Done
> exe mtd
mtd: ./ot-cli-ftd
Done
```

### go \<duration\> \[speed \<particular-speed\>\]

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
Show help text for all supported CLI commands.

### help \<command\>
Show help text for a specific CLI command.

### joins

Connect finished joiner sessions.

```bash
> joins
node=2    join=4.899s session=5.000s
Done
```

### log \[ debug | info | warn | error \]

Inspect the current log level, or set a new log level. The default is taken from the command line argument,
or 'warn' if nothing specified. Use 'debug' to see detailed log messages.

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

### move \<node-id\> \<x\> \<y\>

Move a node to the target position.

```bash
> move 1 200 300
Done
```

### netinfo \[version "\<string\>"\] \[commit "\<string\>"\] \[real y|n\]

Set network info.

```bash
> netinfo version "Latest"
Done
> netinfo version "Latest" commit "b49ee08"
Done
> netinfo real y
Done
```
### node \<node-id\>

Switch CLI context to a specific OT node. From within this new context, regular OT commands (e.g. "help") can be 
used to directly interact with the node. The command 'exit' or 'node 0' can then be used again to exit the node 
context and return the CLI to global context.

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
commands. This is adding the exclamation mark '!' character before the command.

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

### node \<node-id\> "\<command\>"

Run an OpenThread CLI command on a specific node. 

```bash
> node 1 "state"
leader
Done
```

### nodes

List nodes.

```bash
> nodes
id=1	extaddr=62cfcf3c5556ac7c	rloc16=c000	x=200	y=300	failed=false
id=2	extaddr=6a7d9d31e3511147	rloc16=3000	x=278	y=708	failed=false
id=3	extaddr=266db93fad653782	rloc16=2800	x=207	y=666	failed=false
Done
```

### partitions (pts)

List partitions. 
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

### ping \<src-id\> \[\<dst-id\> \[\<addr-type\>\] | "\<dst-addr\>" \] \[datasize \<datasize\>\] \[count \<count\>\] \[interval \<interval\>\] \[hoplimit \<hoplimit\>\]

Ping from the source node to a destination (another node or an IPv6 address). 

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

Get the global packet loss ratio

```bash
> plr 
0
Done
```

### plr \<plr\>

Set the global packet loss ratio

```bash
> plr 0.5
0.5
Done
```

### radio \<node-id\> \[<node-id> ...\] \[on \| off \| ft \<fail-duration\> \<fail-interval\>\]

Set the radio on/off/fail time parameters in seconds. 

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
```

`ft 10 60` means the nodes' radio will on average be non-functional for 10 seconds every 60 seconds. 

### radiomodel \[\<modelName\>\]

Get the name of the currently used radiomodel (RF propagation model and radio chip characteristics for all nodes)
or set the current model to another model by providing the name or an alias of the model. Current models supported:

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

### scan \<node-id\>

Perform a network scan.

```bash
> scan 2
| J | Network Name     | Extended PAN     | PAN  | MAC Address      | Ch | dBm | LQI |
+---+------------------+------------------+------+------------------+----+-----+-----+
| 0 | OpenThread       | dead00beef00cafe | face | 66c6bfef495534af | 11 | -20 |   0 |
Done
```

### speed

Get the simulating speed.

```bash
> speed
8
Done
```

### speed \<speed\> 

Set the simulating speed. 

```bash
> speed 10
Done
> speed
10
Done
```

### speed (max | inf)

Set maximum simulating speed.

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

### title "\<string\>"

Set simulation title.

```bash
> title "Example"
Done
```

### title "\<string\>" \[x \<int\>\] \[y \<int\>\]

Set simulation title at specified position. 

```bash
> title "Example" x 100 y 200
Done
```

### title "\<string\>" \[fs \<int\>\]

Set simulation title with specified font size. 

```bash
> title "Example" fs 30
Done
```

### unwatch \<node-id\> \[<node-id> ...\]

Disable the watch status for one or more nodes. See [watch](#watch-node-id-node-id-) for details.

### unwatch all

Disable the watch status for all nodes. See [watch](#watch-node-id-node-id-) for details.

### watch \[\<node-id\>\] \[\<node-id\> ...\]

Enable additional, detailed log messages on selected node(s) only. This can be useful for interactive debugging or 
inspection of a node's behavior. 

* To see all nodes currently being watched, use "watch" without parameters.
* By default, watching a node will only display OT stack log messages from level Info (I) or up. To see Debug (D) 
  messages, or only Warn (W) or Error/Critical (C) messages, use 
  [watch \<LogLevel\>](#watch-node-id-node-id--logLevel)

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

#### watch \<node-id\> \[\<node-id\> ...\] \[\<LogLevel\>\]
This is an advanced use of the watch command with LogLevel option. Adding the `<LogLevel>` optional parameter will  
cause OT stack log messages from indicated log level, or higher (more important), to be shown. By default, only the 
Info (I) level or up is shown. Setting the level can be useful for interactive debugging or inspection of a node's behavior 
including the operation of its simulated radio.

* Valid long-form LogLevels are "debug", "info", "note", "warn", "error", or "crit" (same as "error").
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

Enable the watch status for all nodes. See [watch](#watch-node-id-node-id-) for details and 
[watch \<LogLevel\>](#watch-node-id-node-id--loglevel) for the LogLevel option.

```bash
> watch all
Done
> watch all debug
Done
> 
```

#### watch default \[\<LogLevel\>\]

Set the default watch status and `LogLevel` of all newly created nodes. See above for `LogLevel` values.

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

Open a web browser for visualization. 

```bash
> web
Done
```
