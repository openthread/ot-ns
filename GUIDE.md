# OTNS Guide
This guide covers the installation of Go, installation of OTNS, use of the OTNS Web UI and the OTNS CLI.

## Install Go

OTNS requires [Go 1.18 or higher](https://golang.org/dl/) to build:

 - Install Go from https://golang.org/dl/
 - Add `$(go env GOPATH)/bin` (normally `$HOME/go/bin`) to `$PATH`.

## Get OTNS code
Recursive cloning of submodules is needed to obtain the right platform code to build simulated OpenThread nodes.

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

This command also builds and installs the required OpenThread nodes.

## Build OpenThread for OTNS (Optional)

This fork of OTNS uses POSIX simulation to simulate Thread nodes, with a specific platform `ot-rfsim`.
The simulator uses node executables such as `ot-cli-ftd`. By default, the `install` script will build 
a common set of OpenThread nodes of different version (v1.1, v1.2, v1.3.0, v1.3.1, and "latest") that 
are used in the various examples and unit-tests of OTNS.

To build or rebuild yourself an executable with platform `ot-rfsim` for OTNS, see the example build below. 
It shows a build with default settings that builds an OpenThread node of version "latest".  This version is the 
latest one that is bundled with the current checkout out commit of OTNS. It gets bumped occassionally to the 
latest OpenThread main branch.

NOTE: the `bootstrap` step only has to be executed only once, to install the required dev tools.

```bash
$ cd ot-rfsim
$ ./script/bootstrap
$ ./script/build_latest
$ cd ..
```

To build earlier OpenThread code versions for inclusion in the simulation, such as 1.1, 1.2 or 1.3, the following 
commands can be used:

```bash
$ cd ot-rfsim
$ ./script/build_v11
$ ./script/build_v12
$ ./script/build_v13
$ ./script/build_v131
$ cd ..
```

These nodes of specific versions can be added to a simulation using specific flags in the `add` command that adds 
a node. Type `help add` in OTNS to see this.

NOTE: all of the above version-specific build scripts may manipulate the submodule 'openthread' to get a specific 
desired historical commit.

The generic build script can be invoked as shown below. This will build whatever code is currently 
residing in the 'openthread' submodule without switching to a specific OT version.

```bash
$ cd ot-rfsim
$ ./script/build
$ cd ..
```

Finally, the generic build script allows setting or overriding any OT build arguments. See an example below:

```bash
$ cd ot-rfsim
$ ./script/build -DOT_FULL_LOGS=OFF -DOT_COAP_OBSERVE=ON -DOT_TCP=ON
$ cd ..
```

In this example a node is built with debug logs off (for speed in simulation), CoAP-observe enabled, and TCP enabled.

## Run OTNS

Preferably run OTNS from the working directory (i.e. the root of this repo):

```bash
cd ~/otns
otns
```

Running from this directory ensures that OTNS can find the standard binaries (version latest, v11, v12, v13, etc - 
these are stored in `./ot-rfsim/ot-versions`).
OTNS can be run also from any directory in which the node executable(s) such as `ot-cli-ftd` (and optionally 
`ot-cli-mtd`) are placed. In this case, it will use the executable from the current directory if it can find the 
right one.

If started successfully, OTNS by default opens a web browser for network visualization and management.
To see what command-line parameters are supported, use `-h`:

```bash
otns -h
```

## Use OTNS-Web

Use a web browser to manage the simulated Thread network:

* Add, delete, and move various types of OpenThread nodes
* Disable and recover node radios
* Adjust simulation speed
* See some logged events
* See nodes' energy usage (Alpha feature - pending validation)

## Use OTNS CLI

See [OTNS CLI Reference](cli/README.md). 

## OTNS Python Scripting

[pyOTNS](pylibs/otns) library provides utilities to create and manage simulations through OTNS CLI. 

### Python Scripting Documentation

To review the `pyOTNS` documentation:
1. Start `pydoc3` document server:
    ```bash
    pydoc3 -p 8080
    ```
2. Open a web browser and navigate to http://localhost:8080/otns.html.

### Example Python Scripts
For example test scripts, see [pylibs/examples](pylibs/examples). The scripts in the 
[pylibs/unittests](pylibs/unittests) and [pylibs/stress_tests](pylibs/stress_tests) 
directories provide even more versatile examples.

Note that the example [`interactive_cli.py`](pylibs/examples/interactive_cli.py) explains how to set up a Python 
scripted simulation in combination with a user typing CLI commands interactively. The interaction part then typically 
happens at the end of a simulation, after a mesh network topology has been set up.

The example [`interactive_cli_threaded.py`](pylibs/examples/interactive_cli_threaded.py) explains how to set up CLI 
interaction that runs in parallel with the Python script being executed.