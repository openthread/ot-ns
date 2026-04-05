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
# When a running script is killed, the cleanup() will kill all child processes.

echo "[DEBG] ot-br.sh started"

cleanup() {
    echo "[DEBG] ot-br.sh: caught signal, cleaning up child processes"
    jobs -p | xargs -r kill
    wait
    exit 0
}

trap cleanup SIGINT SIGTERM

NODE_ID=$1
BACKBONE_IF_NAME=$2
RADIO_URL=$3
# TODO: 'wpan1' for node 1 etc doesn't work - needs to be wpan0. Check later if there's a way to use others.
THREAD_IF_NAME="wpan0"

echo "[DEBG]  NODE_ID=${NODE_ID}"
echo "[DEBG]  BACKBONE_IF_NAME=${BACKBONE_IF_NAME}"
echo "[DEBG]  THREAD_IF_NAME=${THREAD_IF_NAME}"
echo "[DEBG]  RADIO_URL=${RADIO_URL}"

echo "[DEBG] starting otbr-agent with sudo"
# All otbr-agent output needs to go to stderr, so that ot-ctl CLI interactions are not garbled.
sudo otbr-agent -s -d 7 -I ${THREAD_IF_NAME} -B ${BACKBONE_IF_NAME} "${RADIO_URL}" 1>&2 &
OTBR_PID=$!

echo "[DEBG] otbr-agent started in background (PID=${OTBR_PID})"
sleep 2

# check if the process is still alive without sending any signal
if ! kill -0 "${OTBR_PID}" 2>/dev/null; then
    echo "[ERRO] otbr-agent failed to start or exited immediately"
    exit 1
fi
echo "[DEBG] otbr-agent is running, starting ot-ctl CLI"
sudo ot-ctl

echo "[DEBG] ot-br.sh: ot-ctl CLI exited, cleaning up child processes"
jobs -p | xargs -r kill
wait
echo "[DEBG] ot-br.sh: script exit"
