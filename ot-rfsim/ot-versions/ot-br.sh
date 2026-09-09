#!/usr/bin/env bash
#
#  Copyright (c) 2026, The OpenThread Authors.
#  All rights reserved.
#
#  Redistribution and use in source and binary forms, with or without
#  modification, are permitted provided that the following conditions are met:
#  1. Redistributions of source code must retain the above copyright
#     notice, this list of conditions and the following disclaimer.
#  2. Redistributions in binary form must reproduce the above copyright
#     notice, this list of conditions and the following disclaimer in the
#     documentation and/or other materials provided with the distribution.
#  3. Neither the name of the copyright holder nor the
#     names of its contributors may be used to endorse or promote products
#     derived from this software without specific prior written permission.
#
#  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
#  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
#  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
#  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
#  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
#  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
#  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
#  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
#  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
#  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
#  POSSIBILITY OF SUCH DAMAGE.
#

# ot-br.sh - script to start an OTBR node from an OTNS simulation.

# Log helpers - all script logging goes to stderr, keeping stdout clean for the ot-ctl CLI interaction.
debug() { echo "[DEBG]-OTBR.SH-: $*" 1>&2; }
info()  { echo "[INFO]-OTBR.SH-: $*" 1>&2; }
crit()  { echo "[CRIT]-OTBR.SH-: $*" 1>&2; }

# When a running script receives SIGTERM
cleanup()
{
    debug "caught signal, cleaning up child processes"
    jobs -p | xargs -r kill
    wait
    debug "script exit"
    exit 0
}

trap cleanup SIGINT SIGTERM

sudo_command_failed()
{
    local cmd="$1"
    local cmd_path
    crit "passwordless sudo failed for '${cmd}'"
    crit "add below lines to /etc/sudoers (run: sudo visudo):"

    cmd="otbr-agent"
    cmd_path=$(command -v "$cmd" 2>/dev/null)
    crit "    ${USER} ALL=(ALL) NOPASSWD: SETENV: ${cmd_path:-/usr/local/sbin/${cmd}}"

    cmd="ot-ctl"
    cmd_path=$(command -v "$cmd" 2>/dev/null)
    crit "    ${USER} ALL=(ALL) NOPASSWD: ${cmd_path:-/usr/local/bin/${cmd}}"
    exit 1
}

socket_in_use()
{
    if command -v ss >/dev/null 2>&1; then
        ss -x src "$1" | grep -q "$1"
    elif command -v lsof >/dev/null 2>&1; then
        lsof "$1" >/dev/null 2>&1
    else
        return 1 # can't check, assume not in use (1 = false)
    fi
}

debug "script started"

NODE_ID=$1
BACKBONE_IF_NAME=$2
AGENT_PARAM=$3
RADIO_URL=$4
THREAD_IF_NAME="wpan${PORT_OFFSET}_${NODE_ID}"
REST_PORT=$((8080 + NODE_ID))
EXTRA_DELAY=0
MAX_WAIT_SEC=5
SOCKET_PATH="/run/openthread-${THREAD_IF_NAME}.sock"
DATA_PATH="/var/lib/thread"

debug "  PORT_OFFSET     =${PORT_OFFSET}"
debug "  NODE_ID         =${NODE_ID}"
debug "  BACKBONE_IF_NAME=${BACKBONE_IF_NAME}"
debug "  THREAD_IF_NAME  =${THREAD_IF_NAME}"
debug "  REST_PORT       =${REST_PORT}"
debug "  RADIO_URL       =${RADIO_URL}"
debug "  SOCKET_PATH     =${SOCKET_PATH}"
debug "  DATA_PATH       =${DATA_PATH}"
debug "  AGENT_PARAM     =${AGENT_PARAM}"

# check for passwordless sudo access to commands
sudo -n otbr-agent -V >/dev/null 2>&1 || sudo_command_failed otbr-agent
sudo -n ot-ctl -h >/dev/null 2>&1 || sudo_command_failed ot-ctl

# check for existing socket usage
if [ -S "${SOCKET_PATH}" ]; then
    if socket_in_use "${SOCKET_PATH}"; then
        crit "socket ${SOCKET_PATH} is already in use - is otbr-agent already running on ${THREAD_IF_NAME}?"
        exit 1
    fi
    debug "existing (unused?) file ${SOCKET_PATH} detected - start with extra delay"
    EXTRA_DELAY=3
fi

info "starting otbr-agent"
# All otbr-agent output redirected to stderr, so that ot-ctl CLI interactions are not garbled.
sudo -n PORT_OFFSET="${PORT_OFFSET}" otbr-agent --data-path "${DATA_PATH}" -s -d 7 -I "${THREAD_IF_NAME}" \
    -B "${BACKBONE_IF_NAME}" --rest-listen-port "${REST_PORT}" "${AGENT_PARAM}" "${RADIO_URL}" 1>&2 &
SUDO_OTBR_PID=$!

debug "otbr-agent started in background (parent PID=${SUDO_OTBR_PID}) - waiting until ready"

# wait for otbr-agent to create its Unix socket
elapsed=0
while [ ! -S "${SOCKET_PATH}" ]; do
    if ! kill -0 "${SUDO_OTBR_PID}" 2>/dev/null; then
        crit "otbr-agent exited before socket was ready"
        exit 1
    fi
    if [ "${elapsed}" -ge "$((MAX_WAIT_SEC * 10))" ]; then
        crit "timed out waiting for otbr-agent socket (${MAX_WAIT_SEC}s)"
        exit 1
    fi
    sleep 0.1
    elapsed=$((elapsed + 1))
done
debug "otbr-agent socket ready after $((elapsed / 10)).$((elapsed % 10))s, adding ${EXTRA_DELAY} s delay"
sleep ${EXTRA_DELAY}
info "starting ot-ctl CLI"
sudo -n ot-ctl -I "${THREAD_IF_NAME}"

debug "ot-br.sh: ot-ctl CLI exited, cleaning up child processes"
jobs -p | xargs -r kill
wait
debug "ot-br.sh: script exit"
