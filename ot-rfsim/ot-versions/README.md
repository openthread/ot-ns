# ot-versions directory

This contains the `ot-cli-ftd` and `ot-cli-mtd` binaries for a number of specific OpenThread builds of current or previously released codebases. Older versions can be used for testing legacy node behavior and backwards-compatibility. Newer versions typically have more features enabled.

Versions are listed below. With each version tag, the directory in this OTNS repository is listed in which the code for that version is located as a Git submodule.

- v11 - `openthread-v11` - A Thread v1.2 codebase compiled with v1.1 version flag. (v1.1 codebase is too old to compile with OT-RFSIM. For this reason, a 1.2 codebase is used.)

- v12 - `openthread-v12` - A Thread v1.2 codebase compiled with v1.2 version flag.

- v13 - `openthread-v13` - A Thread v1.3 codebase compiled with v1.3 version flag; tag [thread-reference-20230119](https://github.com/openthread/openthread/tree/thread-reference-20230119).

- latest - `openthread` - A recent OpenThread `main` branch commit that's the default `openthread` submodule. The version is selected with v1.4 version flag, currently. If in the future the Thread version increases, this build will track that.

- br - `openthread` - Same code as 'latest', but builds a Thread Border Router (BR).

- ccm - `openthread-ccm` - (In development) A codebase supporting Thread Commercial Commissioning Mode (CCM).

- br-ccm - `openthread-ccm` - (In development) A Thread CCM Border Router.

Build scripts: the build scripts to build all of the versions are `../script/build_<version-tag>`. Each of these specific build scripts invokes the general `build` script. The `../script/build_all` builds all versions.
