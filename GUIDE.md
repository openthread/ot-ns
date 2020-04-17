# OTNS Guide

## Install Go

OTNS requires [Go 1.11+](https://golang.org/dl/) to build:

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
cd output/x86_64-unknown-linux-gnu/bin
otns
```
or 
```bash
otns -bin output/x86_64-unknown-linux-gnu/bin
```

> `x86_64-unknown-linux-gnu` is the output directory for building OpenThread 
> on Linux. On macOS, the directory will be `x86_64-apple-darwin`. 
> Check the `output` folder for the correct directory to use.

If started successfully, OTNS opens a web browser for network visualization and management. 

## Use OTNS-Web

Use a web browser to manage the simulated Thread network:

* Add, delete, and move various types of OpenThread nodes
* Disable and recover node radios
* Adjust simulation speed

## Use OTNS CLI

See [OTNS CLI Reference](cli/README.md). 

## OTNS Python Scripting

[otns library](pylibs/otns) provides utilities to create and manage simulations through OTNS CLI. 

Check the scripts in [pylibs/examples](pylibs/examples) for examples.