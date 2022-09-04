# OTNS Guide

## Install Go

OTNS requires [Go 1.17+](https://golang.org/dl/) to build:

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

To build OpenThread for OTNS:

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
