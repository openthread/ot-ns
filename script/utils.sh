#!/bin/bash
# Copyright (c) 2020-2025, The OTNS Authors.
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

# Collection of utility functions for scripts in this directory or elsewhere.

die()
{
    echo >&2 "ERROR: $*"
    exit 1
}

function realpathf()
{
    # the Python3 method is a backup. Used for max portability.
    realpath -s "$1" 2>/dev/null || python3 -c "import os; print(os.path.realpath('$1'))"
}

function installed()
{
    command -v "$1" >/dev/null 2>&1
}

# checks if a Golang package install would give us at least the minimum version
function check_go_version_installable()
{
    local min_ver="$1"
    local go_ver
    local golang_apt_pkg="$2"
    local golang_brew_pkg="$3"
    local pkg_info

    #shellcheck disable=SC2154
    if [[ $Darwin == 1 ]]; then
        pkg_info=$(brew info "${golang_brew_pkg}" 2>/dev/null)
    else
        pkg_info=$(apt show "${golang_apt_pkg}" 2>/dev/null | grep '^Version:')
    fi
    if [[ ${pkg_info} =~ 1\.([0-9]+)[^0-9] ]]; then
        go_ver=${BASH_REMATCH[1]}
    else
        die "unexpected output from package manager: ${pkg_info}"
    fi
    if [[ ${go_ver} -ge ${min_ver} ]]; then
        echo "Golang ('go') version 1.${go_ver} can be installed from package '${golang_apt_pkg}'"
        return 0
    fi
    echo "Golang ('go') version >= 1.${min_ver} cannot be installed from package '${golang_apt_pkg}', which has 1.${go_ver}"
    return 1
}

# install 'go' from default package while checking for minimum version
install_golang()
{
    local min_ver="$1"
    local golang_apt_pkg="$2"
    local golang_brew_pkg="$3"

    if ! check_go_version_installable "${min_ver}" "${golang_apt_pkg}" "${golang_brew_pkg}"; then
        die "Please install Go 1.${min_ver} or higher manually from: https://go.dev/dl/"
    fi
    install_package go --apt "${golang_apt_pkg}" --brew "${golang_brew_pkg}"
    if ! installed go; then
        die "Golang was installed but 'go' not found in PATH - please fix this manually, then retry."
    fi
}

# check if go version 1.N or higher is installed
function check_minimum_go_version()
{
    local go_min_ver="$1"
    local go_ver

    go_ver="$(go version)"
    if [[ $go_ver =~ go1\.([0-9]+)[^0-9] ]]; then
        go_ver="${BASH_REMATCH[1]}"
        if [[ ${go_ver} -lt ${go_min_ver} ]]; then
            echo "OTNS2 requires Golang ('go') version >= 1.${go_min_ver}; Your version: 1.${go_ver}"
            return 1
        fi
        return 0
    else
        die "unexpected output from 'go' command"
    fi
}

# check if python3 version 3.N or higher is installed
function check_minimum_python3_version()
{
    local py_min_ver="$1"
    local py_ver

    py_ver="$(python3 --version)"
    if [[ ${py_ver} =~ Python\ 3\.([0-9]+)[^0-9] ]]; then
        py_ver="${BASH_REMATCH[1]}"
        if [[ ${py_ver} -lt ${py_min_ver} ]]; then
            echo "OTNS2 requires Python ('python3') version >= 3.${py_min_ver}; Your version: 3.${py_ver}"
            return 1
        fi
        return 0
    else
        die "unexpected output from 'python3' command"
    fi
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

    die "Failed to install '$cmd'. Please install it manually."
}

install_pretty_tools()
{
    if ! installed golangci-lint; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$(go env GOPATH)"/bin v2.1.6
    fi

    install_package shfmt --apt shfmt --brew shfmt
    install_package shellcheck --apt shellcheck --brew shellcheck
}

get_openthread_commit()
{
    local commit="${1}"
    local ot_commit_dir="${2}"
    local ot_dir="${3}"
    local patch_id="${4}"

    ot_commit_dir=$(realpathf "${ot_commit_dir}")
    ot_dir=$(realpathf "${ot_dir}")
    mkdir -p "${ot_commit_dir}"
    (
        cd "${ot_dir}" || die "OpenThread directory not found: ${ot_dir}"
        git archive --format=tar "${commit}" | tar -x -C "${ot_commit_dir}"
        cd "${ot_commit_dir}" || die "OpenThread target directory for commit not found: ${ot_commit_dir}"
        patch -p1 <../etc/"${patch_id}".patch
    )
}
