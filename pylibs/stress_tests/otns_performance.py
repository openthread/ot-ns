#!/usr/bin/env python3
#
# Copyright (c) 2020-2023, The OTNS Authors.
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
# OTNS Performance Stress test:
#   Simulate 4x8 nodes at max speed without injected traffic or failure for PERF_SIMULATE_TIME, measure the
#   execution (real) time.
# Topology:
#   Router 4x8
# Fault Injections:
#   None
# Pass Criteria:
#   Execution time <= 60s
#
import time

from BaseStressTest import BaseStressTest

XGAP = 200
YGAP = 200

ROWS, COLS = 4, 8
assert ROWS * COLS <= 32

WAIT_NETWORK_FORM_PARTITION_TIME = 1000
WAIT_NETWORK_STABILIZE_TIME = 100
PERF_SIMULATE_TIME = 3600


class OtnsPerformanceStressTest(BaseStressTest):
    SUITE = 'otns-performance'

    def __init__(self):
        super(OtnsPerformanceStressTest,
              self).__init__("OTNS Performance Test",
                             ['Simulation Time', 'Execution Time', 'Speed Up', 'Alarm Events', 'Radio Events'],
                             rand_seed=48392)

    def run(self):
        ns = self.ns

        for r in range(ROWS):
            for c in range(COLS):
                nid = ns.add("router", 100 + XGAP * c, 100 + YGAP * r)
                # make sure every node become Router
                ns.node_cmd(nid, "routerupgradethreshold 32")
                ns.node_cmd(nid, 'routerdowngradethreshold 33')
                ns.node_cmd(nid, 'routerselectionjitter 10')
                ns.go(10)  # give time to connect

        for r in range(ROWS):
            for c in range(COLS):
                nid = 1 + c + r * COLS
                expected_state = 'leader' if (r, c) == (0, 0) else 'router'
                self.expect_node_state(nid, expected_state, 10)  # give time to become Router

        secs = 0
        formed_one_partition_ok = False
        while secs < WAIT_NETWORK_FORM_PARTITION_TIME:
            ns.go(1)
            pars = ns.partitions()
            if len(pars) == 1 and 0 not in pars:
                formed_one_partition_ok = True
                break

        # should always form 1 partition after WAIT_NETWORK_FORM_PARTITION_TIME
        assert formed_one_partition_ok, ns.partitions()
        # run WAIT_NETWORK_STABILIZE_TIME to allow the network to stabilize
        ns.go(WAIT_NETWORK_STABILIZE_TIME)
        counter0 = ns.counters()

        t0 = time.time()
        ns.go(PERF_SIMULATE_TIME)
        t1 = time.time()
        duration = t1 - t0
        print(duration, PERF_SIMULATE_TIME / duration)

        counter = ns.counters()
        for k in counter:
            counter[k] -= counter0[k]

        print('counters', ns.counters())

        self.result.append_row('%ds' % PERF_SIMULATE_TIME, '%ds' % duration, '%d' % (PERF_SIMULATE_TIME / duration),
                               counter['AlarmEvents'], counter['RadioEvents'])

        self.result.fail_if(duration > 60, f'Execution Time ({duration}) > 60s')
        self.result.fail_if(counter['AlarmEvents'] > 900000,
                            f"Too many AlarmEvents: {counter['AlarmEvents']} > 900000")
        self.result.fail_if(counter['RadioEvents'] > 680000,
                            f"Too many RadioEvents: {counter['RadioEvents']} > 680000")


if __name__ == '__main__':
    OtnsPerformanceStressTest().run()
