# OTNS Guide

## Install Go

OTNS requires [Go 1.17+](https://golang.org/dl/) to build:

 - Install Go from https://golang.org/dl/
 - Add `$(go env GOPATH)/bin` (normally `$HOME/go/bin`) to `$PATH`.

## Get OTNS code

```bash
git clone --recurse-submodules https://github.com/openthread/ot-ns.git ./otns
cd otns
```

## Install Dependencies

```bash
./script/install-deps
```

## Install OTNS

```bash
./script/install
```

## Build OpenThread for OTNS

OTNS uses POSIX simulation to simulate Thread nodes.

To build OpenThread for OTNS using 'cmake' (preferred, 'make' is not recommended): below shows an example build 
for OT-NS without generating full OT log messages (OT_FULL_LOGS=OFF). To see extra debug info of the node during
the simulation, use OT_FULL_LOGS=ON.

```bash
$ git clone https://github.com/openthread/openthread openthread
$ cd openthread
$ ./script/bootstrap
$ ./bootstrap
$ ./script/cmake-build simulation -DOT_PLATFORM=simulation -DOT_OTNS=ON
 -DOT_SIMULATION_VIRTUAL_TIME=ON -DOT_SIMULATION_VIRTUAL_TIME_UART=ON -DOT_SIMULATION_MAX_NETWORK_SIZE=999 \
 -DOT_COMMISSIONER=ON -DOT_JOINER=ON -DOT_BORDER_ROUTER=ON -DOT_SERVICE=ON -DOT_COAP=ON -DOT_FULL_LOGS=OFF
```

## Run OTNS

After building OpenThread, run OTNS:

```bash
cd build/simulation/examples/apps/cli
otns
```

If started successfully, OTNS opens a web browser for network visualization and management.
To see what command-line parameters are supported, use `-h`:

```bash
otns -h
```

## Use OTNS-Web

Use a web browser to manage the simulated Thread network:

* Add, delete, and move various types of OpenThread nodes
* Disable and recover node radios
* Adjust simulation speed

## Use OTNS CLI

See [OTNS CLI Reference](cli/README.md). 

## OTNS Python Scripting

[pyOTNS](pylibs/otns) library provides utilities to create and manage simulations through OTNS CLI. 

To review the `pyOTNS` documentation:
1. Start `pydoc3` document server:
    ```bash
    pydoc3 -p 8080
    ```
2. Open a web browser and navigate to http://localhost:8080/otns.html.

For example test scripts, see [pylibs/examples](pylibs/examples).
