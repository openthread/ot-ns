#!/usr/bin/env python3
# Copyright (c) 2023, The OTNS Authors.
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
 
from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

# set to 0, 1 or 2
TEST_SCENARIO = 0

SPEED = 1e6
RADIO_RANGE = 250
PING_DATA_SIZE = 64
NUM_PINGS = 100
PING_INTERVAL = 1


# get a string that prints statistics/info of node 'nodeid'
def print_stats(ns, nodeid: int):
    return "CCA threshold (dBm): " + "\n".join(ns.node_cmd(nodeid, 'ccathreshold')) + "\n\n" + \
           "\n".join(ns.node_cmd(nodeid, 'neighbor table')) + "\n" + \
           "\n".join(ns.node_cmd(nodeid, 'counters mac'))

def print_pings(pings):
    s = ""
    for p in pings:
        s += "   " + str(p) + "\n"
    return s

def main():
    # Use -watch trace to enable detailed logging of individual nodes, including radio/CCA details.
    ns = OTNS(otns_args=["-log", "info", "-watch", "trace"])
    ns.speed = SPEED
    ns.web()
    # The test setups are only valid for this radio model.
    ns.radiomodel = 'MutualInterference'

    if TEST_SCENARIO == 0:
        print('\n ==============Scenario A (Static): Interferer out of Source range ======================')
        src = ns.add("router", x=350, y=300, radio_range=RADIO_RANGE)
        ns.go(10)
        dst = ns.add("router", x=700, y=300, radio_range=RADIO_RANGE)
        intf = ns.add("router", x=870, y=300, radio_range=RADIO_RANGE)

    elif TEST_SCENARIO == 1:
        print('\n ==============Scenario B (Static): all 3 nodes in mutual range ======================')
        src = ns.add("router", x=350, y=300, radio_range=RADIO_RANGE)
        ns.go(10)
        dst = ns.add("router", x=700, y=300, radio_range=RADIO_RANGE)
        intf = ns.add("router", x=525, y=450, radio_range=RADIO_RANGE)
        ns.node_cmd(dst, "state router") # due to very low RSSI, a node may stay End Device. This forces it to Router.
        ns.node_cmd(intf, "state router")

    elif TEST_SCENARIO == 2:
        print('\n ==============Scenario C (Static): Interferer out of Destination range ======================')
        src = ns.add("router", x=350, y=300, radio_range=RADIO_RANGE)
        ns.go(10)
        dst = ns.add("router", x=700, y=300, radio_range=RADIO_RANGE)
        intf = ns.add("router", x=175, y=300, radio_range=RADIO_RANGE)
        ns.node_cmd(dst, "state router") # due to very low RSSI, a node may stay End Device. This forces it to Router.
        ns.node_cmd(intf, "state router")

    else:
        return

    # simulate some time for network to form stably
    ns.go(600)

    # do the pings - interferer pings to an unused multicast address at the same times as src,
    # causing potential interference. This is done to avoid ping reply's from src or dst
    # that would disrupt the ongoing ping process between src and dst.
    ns.node_cmd(intf, f'ping async ff02::dead 64 {NUM_PINGS} {PING_INTERVAL}')
    ns.ping(src,dst, datasize=PING_DATA_SIZE, count=NUM_PINGS, interval=PING_INTERVAL)
    ns.go(NUM_PINGS * PING_INTERVAL + 60)

    ns.loglevel = 'warn'  # suppress debug output interfering with below prints.

    print('\n*** Ping results:\n', print_pings(ns.pings()))
    print('\n*** Source:\n', print_stats(ns,src))
    print('\n*** Destination:\n', print_stats(ns,dst))
    print('\n*** Interferer:\n', print_stats(ns,intf))


if __name__ == '__main__':
    main()
