#!/bin/bash
# Copyright (c) 2020-2024, The OTNS Authors.
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the
#    names of its contributors may be used to endorse or promote products
#    derived from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

set -euox pipefail

if [[ "$(uname)" == "Darwin" ]]; then
    export readonly Darwin=1
    export readonly Linux=0
elif [[ "$(uname)" == "Linux" ]]; then
    export readonly Darwin=0
    export readonly Linux=1
else
    die "Unknown OS: $(uname)"
fi

# shellcheck source=script/utils.sh
. "$(dirname "$0")"/utils.sh

export readonly SCRIPTDIR
SCRIPTDIR=$(realpath "$(dirname "$0")")
export readonly OTNSDIR
OTNSDIR=$(realpath "$SCRIPTDIR"/..)
export readonly GOPATH
GOPATH=$(go env GOPATH)
export PATH=$PATH:"$GOPATH"/bin
mkdir -p "$GOPATH"/bin

export readonly GOLINT_ARGS=(-E goimports -E whitespace -E goconst -E exportloopref -E unconvert)
export readonly OTNS_BUILD_JOBS
OTNS_BUILD_JOBS=$(getconf _NPROCESSORS_ONLN)
export readonly OTNS_EXCLUDE_DIRS=(web/site/node_modules/ openthread/ build/)

go_install()
{
    local pkg=$1
    go install "${pkg}" || go get "${pkg}"
}

get_openthread()
{
    if [[ ! -f ./openthread/script/bootstrap ]]; then
        # --depth 1 is not used here, due to need to build historic commits for OT nodes.
        git submodule update --init --recursive
    fi
}

function get_build_options()
{
    local cov=${COVERAGE:-0}
    if [[ $cov == 1 ]]; then
        echo "-DOT_COVERAGE=ON"
    else
        # TODO: MacOS CI build fails for empty options. So we give one option here that is anyway set.
        echo "-DOT_OTNS=ON"
    fi
}

build_openthread()
{
    get_openthread
    install_openthread_buildtools
    (
        cd ot-rfsim
        ./script/build_latest "$(get_build_options)"
    )
}

build_openthread_br()
{
    if [[ ! -f ./ot-rfsim/ot-versions/ot-cli-ftd_br ]]; then
        get_openthread
        install_openthread_buildtools
        (
            cd ot-rfsim
            ./script/build_br "$(get_build_options)"
        )
    fi
}

build_openthread_versions()
{
    get_openthread
    install_openthread_buildtools
    (
        cd ot-rfsim
        ./script/build_all "$(get_build_options)"
    )
}

activate_python_venv()
{
    if [ ! -d .venv-otns ]; then
        python3 -m venv .venv-otns
    fi
    # shellcheck source=/dev/null
    source .venv-otns/bin/activate
}
