# OTNS Guide

## Install Go

OTNS requires [Go 1.13+](https://golang.org/dl/) to build:

 - Install Go from https://golang.org/dl/
 - Add `$(go env GOPATH)/bin` (normally `$HOME/go/bin`) to `$PATH`.

## Get OTNS code

```bash
git clone https://github.com/openthread/ot-ns.git ./otns
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

To build OpenThread for OTNS using 'cmake' (preferred): below shows an example build for OT-NS that has support 
for OT-external RF/timing models. The build option OT_FULL_LOGS can also be set to 'ON' in this case, for extra
debug info.

```bash
$ git clone https://github.com/openthread/openthread openthread
$ cd openthread
$ ./script/bootstrap
$ ./bootstrap
$ ./script/cmake-build simulation -DOT_PLATFORM=simulation -DOT_OTNS=ON -DOT_SIMULATION_EXT_RF_MODELS=ON\
 -DOT_SIMULATION_VIRTUAL_TIME=ON -DOT_SIMULATION_VIRTUAL_TIME_UART=ON -DOT_SIMULATION_MAX_NETWORK_SIZE=999 \
 -DOT_COMMISSIONER=ON -DOT_JOINER=ON -DOT_BORDER_ROUTER=ON -DOT_SERVICE=ON -DOT_COAP=ON -DOT_FULL_LOGS=OFF
```

To build OpenThread for OTNS using 'make' (not preferred):

```bash
git clone https://github.com/openthread/openthread openthread
cd openthread
./script/bootstrap
./bootstrap
make -f examples/Makefile-simulation OTNS=1
```

## Run OTNS

After building OpenThread, run OTNS:

```bash
cd output/simulation/bin
otns
```

If started successfully, OTNS opens a web browser for network visualization and management. 

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
