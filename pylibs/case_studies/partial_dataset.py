#!/usr/bin/env python3
# Copyright (c) 2024, The OTNS Authors.
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

# Case study on new Thread device that's commissioned with only a partial
# dataset, which excludes channel information. To still join the Thread
# Network, it will search over all channels to find its Parent.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_BR = 1
NUM_NODES = 1
DX = 50  # pixels spacing

# defines a setup script with partial dataset (no channel info)
SCRIPT_PARTIAL_DATASET = """
dataset clear
dataset networkkey 00112233445566778899aabbccddeeff
dataset commit active

ifconfig up
thread start
"""


def main():
    # use PCAP type wpan-tap to show the channel of each frame
    ns = OTNS(otns_args=['-seed', '74541', '-phy-tx-stats', '-pcap', 'wpan-tap'])
    ns.speed = 1000
    ns.radiomodel = 'MutualInterference'
    ns.web()
    ns.web('stats')
    ns.watch_default('trace')

    # setup of Border Routers
    for i in range(1, NUM_BR + 1):
        nid = ns.add("br", x=225 + i * 100, y=140)
        ns.node_cmd(nid, "dataset init active")
        ns.node_cmd(nid, "dataset channel 26")
        ns.node_cmd(nid, "dataset activetimestamp 1718642107")
        ns.node_cmd(nid, "dataset commit active")
    nid_br = 1
    ns.go(100)

    # setup of Router nodes or End devices
    cx = 100
    cy = DX
    for i in range(1, NUM_NODES + 1):
        nid = ns.add("router", x=cx, y=cy, script=SCRIPT_PARTIAL_DATASET)
        cx += DX
        if cx >= 600:
            cx = 100
            cy += DX

    # allow some time for channel scanning and connecting
    ns.kpi_start()
    ns.go(100)
    ns.kpi_stop()

    # at the end, node's status can be inspected via CLI. KPI files contain activity over channels.
    # The PCAP file contains channel info of all frames sent, to see the channel scanning happening.
    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
