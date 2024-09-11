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

function die()
{
    echo "fatal: $1"
    false
}

function realpathf()
{
    # the Python3 method is a backup. Used for max portability.
    realpath -s "$1" || python3 -c "import os; print(os.path.realpath('$1'))"
}

function installed()
{
    command -v "$1" >/dev/null 2>&1
}

function repeat()
{
    local n=$1
    local cmd="${*:2}"

    for _ in $(seq "$n"); do
        $cmd && return 0
    done
    return 1
}

apt_update_once=0

install_package()
{
    local cmd=$1
    shift 1

    # only install if command not available
    if installed "$cmd"; then
        return 0
    fi

    while (("$#")); do
        case "$1" in
            --apt)
                if installed apt-get; then
                    if [ "$apt_update_once" -eq "0" ]; then
                        sudo apt-get update || die "apt-get update failed"
                        apt_update_once=1
                    fi

                    sudo apt-get install -y --no-install-recommends "$2"
                    return 0
                fi

                shift 2
                ;;
            --brew)
                if installed brew; then
                    brew install "$2"
                    return 0
                fi

                shift 2
                ;;
            --snap)
                if installed snap; then
                    sudo snap install "$2"
                    return 0
                fi

                shift 2
                ;;
            *)
                PARAMS="$PARAMS $1"
                echo "Error: Unsupported flag $1" >&2
                return 1
                ;;
        esac
    done

    die "Failed to install $cmd. Please install it manually."
}

function install_pretty_tools()
{
    # TODO Known bug: version v1.59.0 won't work with Go 1.23 or higher. Requires version <= 1.22
    if ! installed golangci-lint; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)"/bin v1.59.0
    fi

    install_package shfmt --apt shfmt --brew shfmt
    install_package shellcheck --apt shellcheck --brew shellcheck
}
