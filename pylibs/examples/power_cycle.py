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

#
# Example of a large network going through a 'power cycle' jointly.

import logging

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

XGAP = 100
YGAP = 100
RADIO_RANGE = 130
N = 10  # network size is N^2

class PowerCycleExample:

    def setup_topology_size_n(self, n):
        for r in range(n):
            for c in range(n):
                self.ns.add("router", XGAP * (c+1), YGAP * (r+1), radio_range=RADIO_RANGE)

    def __init__(self):
        self.ns = OTNS(otns_args=["-log", "debug"]) # OTNS log level
        self.ns.logconfig(logging.DEBUG) #pyOTNS log level

    def run(self):
        ns: OTNS = self.ns
        ns.set_radioparam('MeterPerUnit', 0.2) # set scale of distances

        #ns.watch_default('trace')
        ns.set_title("Network Power Cycle Example - Topology setup")
        ns.web()

        self.setup_topology_size_n(N)
        ns.go(300)
        while len(ns.partitions()) > 1:
            ns.go(10)

        ns.set_title("Network Power Cycle Example - Power cycle upcoming")
        ns.speed = 1
        ns.go(2)

        ns.set_title("Network Power Cycle Example - Power down period")
        Nn=N*N
        #restart_nodes = [13, 14, 18,19]
        restart_nodes = range(1,Nn+1)
        for n in restart_nodes:
            ns.node_cmd(n,'reset')
        ns.go(3)

        ns.set_title("Network Power Cycle Example - Powered up & forming network")
        for n in restart_nodes:
            ns.node_cmd(n,'ifconfig up')
            ns.node_cmd(n,'thread start')

        form_time = 0
        ns.speed = 2
        while True:
            pars = ns.partitions()
            detached_nodes = []
            for n in range(1,Nn+1):
                if ns.get_state(n) == 'detached':
                    detached_nodes.append(n)
            if len(pars) == 1 and 0 not in pars and len(detached_nodes)==0:
                break
            ns.go(1)
            form_time += 1
            ns.set_title(f"Network Power Cycle Example - Network forming ongoing for ~{form_time} sec")

        ns.set_title(f"Network Power Cycle Example - Network re-formed after ~{form_time} sec")

        ns.speed = 1
        ns.autogo = True
        ns.interactive_cli()

if __name__ == '__main__':
    try:
        example = PowerCycleExample()
        example.run()

    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
