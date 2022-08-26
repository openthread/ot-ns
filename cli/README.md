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
* [exit](#exit)
* [go](#go-duration-seconds--ever)
* [joins](#joins)
* [log](#log-level)
* [move](#move-node-id-x-y)
* [netinfo](#netinfo-version-string-commit-string-real-yn)
* [node](#node-node-id-command)
* [nodes](#nodes)
* [partitions (pts)](#partitions-pts)
* [ping](#ping-src-id-dst-id-addr-type--dst-addr--datasize-datasize-count-count-interval-interval-hoplimit-hoplimit)
* [pings](#pings)
* [plr](#plr)
* [radio](#radio-node-id-node-id--on--off--ft-fail-duration-fail-interval)
* [radiomodel](#radiomodel)
* [scan](#scan-node-id)
* [speed](#speed)
* [title](#title-string)
* [web](#web)

## OTNS command reference


### add \<type\> \[x \<x\>\] \[y \<y\>\] \[rr \<radio-range\>\] \[id \<node-id\>\] \[restore\]

Add a node to the simulation and get the node ID. Node ID can be specified, otherwise OTNS assigns the next available one.

If `restore` option is specified, the node restores its network configuration from persistent storage.

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

### exit

Exit OTNS.

```bash
> exit
Done
<EOF>
```

### go \[\<duration-seconds\> | ever\]

Simulate for a specified time in seconds or indefinitely (`ever`). **Only required in `-autogo=false` mode**

```bash
> go 1
Done
> go 10
Done
> go ever
<NEVER FINISHES>
```

### joins

Connect finished joiner sessions.

```bash
> joins
node=2    join=4.899s session=5.000s
Done
```
### log \[\<level\>\]

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

### radiomodel \[\"\<modelName\>\"\]

Get the name of the currently used radiomodel (RF propagation model and radio chip characteristics for all nodes)
or set the current model to another model.

```bash
> radiomodel
InterfereAll
Done
> radiomodel "Ideal"
Ideal
Done
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

### web

Open a web browser for visualization. 

```bash
> web
Done
```
