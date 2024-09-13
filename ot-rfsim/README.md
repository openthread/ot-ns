# OpenThread on RF-SIMulator (OT-RFSIM) platform

This directory contains 'ot-rfsim', an OpenThread platform driver for simulated OT nodes. A simulated OT node can be started in the OTNS simulator. It connects to the simulator using the Unix Domain Socket provided in the commandline parameters.

The easiest way to use this code is just to install OTNS using the `bootstrap` script, following the [OTNS Guide](../GUIDE.md).

## Prerequisites

This CMake project requires an [openthread](https://github.com/openthread/openthread) repository to build the OT nodes. This can be a custom one, indicated by the environment variable `OT_DIR`, or else the default openthread repository will be selected which is a Git submodule located in `../openthread`. If not yet initialized, this submodule will be automatically initialized upon first build.

## Building

For more detailed building instructions see [GUIDE.md](../GUIDE.md).

### Build for use in OT-NS with custom build configuration

Below shows first an example custom build to generate the OT node binaries for an OT-NS simulation. This includes debug logging, for extra debug info that can then be optionally displayed in OT-NS using the `watch` command. The OT debug logging is also stored in a log file, per node.

```bash
$ ./script/build
```

The new executables will end up in the directory `build/custom/bin` by default. Within 'build', there are subdirectories for different node type builds, depending on which build scripts are run (see further below).

Below shows an example custom build for OT-NS with build option OT_FULL_LOGS set to 'OFF', to disable the debug logging. This helps to speed up the simulation because far less socket communication events are then generated.

```bash
$ ./script/build -DOT_FULL_LOGS=OFF
```

Below shows an example build for OT-NS with build option OT_CSL_RECEIVER set to 'OFF', to disable the CSL receiver. This is normally enabled for the FTD build, so that it can emulate an MTD SED with CSL. But there may be a specific reason to disable it for an FTD build. (E.g. because a separate MTD build is done with CSL enabled, already.)

```bash
$ ./script/build -DOT_CSL_RECEIVER=OFF
```

After a successful build, the executable files are found in the directory `./build/custom/bin`.

### Build default v1.1, v1.2, v1.3, v1.\<Latest\> , or Border Router nodes for OT-NS

There are some scripts (`./script/build_*`) for building specific versions of OpenThread nodes for use in OT-NS. There are specific commands in OT-NS to add e.g. v1.1, or v1.2 nodes, all mixed in one simulation.

These build scripts produce executables that are copied into the `ot-versions` directory. The scripts will use specific branches of the `openthread` repository, which are included as submodules in the Git OTNS project. To keep the data size small, each submodule is cloned shallow (depth=1) by default.

## Running

The built custom executables in `bin` can be briefly tested on the command line as follows:

```bash
$ cd build/custom/bin
$ ./ot-cli-ftd
Usage: ot-cli-ftd <NodeId> <OTNS-Unix-socket-file> [<random-seed>]
$
```

This will print a usage message and exit the node.

The `ot-cli-ftd` is by default used in the OT-NS simulator for the "router" and "fed" (FTD) node types. The `ot-cli-mtd` is by default used for MED, SED and SSED. The BR uses `ot-cli-ftd_br`.

One way (but not recommended) is to use the `ot-cli-ftd` is to `cd` to the path where the file is and start OT-NS:

```bash
$ cd build/custom/bin
$ otns
> add router x 50 y 50
1
Done
>
```

The standard way is to run OT-NS from the same directory from where it was installed. In this case, it will use the binaries that are built into `./ot-rfsim/ot-versions`. These binaries can be built using the various `./script/build_*` scripts. Users can copy their own builds into this folder as well and refer to the named binaries when starting a new node in OTNS.
