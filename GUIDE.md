# OTNS Guide

## Install Go

OTNS requires [Go 1.17 or higher](https://golang.org/dl/) to build:

 - Install Go from https://golang.org/dl/
 - Add `$(go env GOPATH)/bin` (normally `$HOME/go/bin`) to `$PATH`.

## Get OTNS code

```bash
git clone --recurse-submodules https://github.com/EskoDijk/ot-ns.git ./otns
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

This fork of OTNS uses POSIX simulation to simulate Thread nodes, with a specific platform `ot-rfsim`.
The simulator starts the node executable `ot-cli-ftd` which is used to simulate all node types.

To build the executable with platform `ot-rfsim` for OTNS using 'cmake', see the example build below. 
It shows a build with default settings.  

```bash
$ cd ot-rfsim
$ ./script/bootstrap
$ ./script/build
```

## Run OTNS

After building OpenThread, run OTNS from the directory where the `ot-cli-ftd` binary was generated:

```bash
cd build/bin
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
