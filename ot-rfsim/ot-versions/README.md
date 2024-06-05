# ot-versions directory

This contains the `ot-cli-ftd` binary for a number of specific OpenThread builds of current or previously released
codebases. Older versions can be used for testing legacy node behavior and backwards-compatibility.
Newer versions typically have more features enabled.

Versions:

- v11 - A Thread v1.2 codebase compiled with v1.1 version flag. (v1.1 codebase is too old to compile with OT-RFSIM.
  For this reason, a 1.2 codebase is used.)

- v12 - A Thread v1.2 codebase compiled with v1.2 version flag.

- v13 - A Thread v1.3 codebase compiled with v1.3 version flag; tag
  [thread-reference-20230119](https://github.com/openthread/openthread/tree/thread-reference-20230119).

- latest - A recent OpenThread `main` branch commit that's the default `openthread` submodule. The version is selected
  with v1.4 version flag, currently. If in the future the Thread version increases, this build will track that.

Build scripts: the build scripts to build all of the versions are `../script/build_*`. Each of these specific build
scripts invokes the general `build` script.
