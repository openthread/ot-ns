# OTNS Guide
This guide covers the installation of Go, installation of OTNS, use of the OTNS Web UI and the OTNS CLI.

## Install Go

OTNS requires [Go 1.18 or higher](https://golang.org/dl/) to build:

 - Install Go from https://golang.org/dl/
 - Add `$(go env GOPATH)/bin` (normally `$HOME/go/bin`) to `$PATH`.

## Get OTNS code

```bash
git clone https://github.com/EskoDijk/ot-ns.git ./otns
cd otns
```

## Automated installation

An automated way to install dependencies, OTNS and all OT nodes is the following command:

```bash
./script/test build_openthread_versions
```

Alternatively the extensive unit tests can be run also with the following command:

```bash
./script/test py-unittests
```

However, this can take a long time (5-10 minutes).

## Manual step-by-step installation

As alternative to above automated installation, scripts can also be run for the individual phases of the installation. 
This is shown in the below subsections. Here, also some more explanation is given of the result of each phase.

### Install Dependencies

```bash
./script/install-deps
```

Running this script will also set up a Python 3 virtual environment (venv) in `.venv-otns` in the project directory.

### Install OTNS

```bash
./script/install
```

This installs `otns` in the Go binary directory of the user (typically `~/go/bin`) and makes the command available in 
the path. Also, it installs the pyOTNS library in the local Python virtual environment `.venv-otns`.
The OT nodes required for running a simulation are not yet installed at this point.

### Install OT nodes

```bash
./script/install-nodes
```

This checks for availability of prebuilt OT nodes, and builds any OT nodes not yet present. This includes a standard 
set of nodes like FTD, MTD, Border Router (BR) and different Thread versions (1.1, 1.2, 1.3.0, 1.4). This build 
can take a long time. During the build specific commits of the `openthread` Git repo submodule will be checked out in 
order to access older OpenThread codebases. In case the build stops unexpectedly and the script is aborted, it may be 
the case that an older OT commit is checked out in the `./openthread` subdirectory. This can be manually restored 
again by running `git submodule update`.

These nodes of specific versions can be added to a simulation using specific flags in the `add` command that adds
a node. Type `help add` in OTNS to see this.

## Manually Build OpenThread for OTNS (Optional)

This fork of OTNS uses POSIX simulation to simulate Thread nodes, with a specific platform `ot-rfsim` located in the 
`ot-rfsim` directory.
The simulator uses node executables such as `ot-cli-ftd`. By default, the `install-nodes` script will build 
a common set of OpenThread nodes of different version (v1.1, v1.2, v1.3.0, and "latest" v1.4) that 
are used in the various examples and unit-tests of OTNS.

To build or rebuild yourself an executable with platform `ot-rfsim` for OTNS, see the example build below. 
It shows a build with default settings that builds an OpenThread node of version "latest".  This version is the 
latest one that is bundled with the current checked out commit of OTNS. It gets bumped occassionally to the 
latest OpenThread main branch after verifying that it works.

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
$ ./script/build_br
$ ./script/build_latest
$ cd ..
```

NOTE: all of the above version-specific build scripts will check if the submodule 'openthread' is at the right specific 
commit that is expected. The `./script/build_all` script provides the automated Git commit checkout and building for 
all versions.

The generic build script can be invoked as shown below. This will build whatever code is currently 
residing in the 'openthread' submodule without switching to a specific OT version and without clearing any previous 
build files. So, it can be used for (faster) incremental builds while developing code.

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

## Run OTNS Interactively

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
* Pause/start the simulation
* Inspect properties and state of nodes
* Open a graph showing node type/status statistics over time
* See some logged events
* See nodes' energy usage (Beta feature - pending validation)

## Use OTNS CLI

See [OTNS CLI Reference](cli/README.md). 

## OTNS Python Scripting

[pyOTNS](pylibs/otns) library provides utilities to create and manage simulations through OTNS CLI. It is installed in a 
Python 3 virtual environment `.venv-otns`. 

### Python Scripting Documentation

To review the `pyOTNS` documentation:
1. Start `pydoc3` document server:
    ```bash
    pydoc3 -p 8080
    ```
2. Open a web browser and navigate to http://localhost:8080/otns.html.

### Example Python Scripts
For example test scripts, see [pylibs/examples](pylibs/examples). 

The scripts in the 
[pylibs/unittests](pylibs/unittests), [pylibs/stress_tests](pylibs/stress_tests) and [pylibs/case_studies](pylibs/case_studies)
directories provide even more versatile examples. The executable/runnable .py scripts in these directories 
can all be run directly from the command line, by typing `./pylibs/<directory>/<script>.py`.

Note that the example [`interactive_cli.py`](pylibs/examples/interactive_cli.py) explains how to set up a Python 
scripted simulation in combination with a user typing CLI commands interactively. The interaction part then typically 
happens at the end of a simulation, after a mesh network topology has been set up.

The example [`interactive_cli_threaded.py`](pylibs/examples/interactive_cli_threaded.py) explains how to set up CLI 
interaction that runs in parallel with the Python script being executed.

### Running a simulation from a Python script

To ensure that the `pyOTNS` library can be found, and the virtual environment is not yet active, enable it first:

```bash
$ source .venv-otns/bin/activate
```

Then run a Python script:

```bash
(.venv-otns) $ ./pylibs/examples/farm.py
...
```
