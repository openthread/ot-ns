#!/usr/bin/env python3
#
# Copyright (c) 2022, The OTNS Authors.
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
# OTNS Long Duration Stress test:
#   Simulate 4 nodes for a long duration (100 days).
#   OpenThread use MilliTimers with `uint32_t` as the underlying value representation.
#   These timers would wrap in about 50 days.
#   This test tries to make sure the OpenThread is functioning properly after a long duration.
# Topology:
#   Router x2, MED x1, SED x1
# Fault Injections:
#   10% packet loss ratio
# Pass Criteria:
#   All nodes are pinging successfully after running for a long duration.
#
import os
import random
import time

from BaseStressTest import BaseStressTest

RADIO_RANGE = 200
XMAX = 300
YMAX = 300

PACKET_LOSS_RATIO = 0.1
TOTAL_SIMULATION_TIME = 10 * 86400 * int(os.getenv("STRESS_LEVEL", "1"))
MOVE_INTERVAL = 3600
PING_INTERVAL = 300
PING_DATA_SIZE = 64

PING_TIMEOUT = PING_INTERVAL

assert TOTAL_SIMULATION_TIME // PING_INTERVAL <= 65535, "too many ping count"


class LongDurationStressTest(BaseStressTest):
    SUITE = 'long-duration'

    def __init__(self):
        super(LongDurationStressTest, self).__init__("Long-Duration stress test",
                                                     ['Simulation Time', 'Execution Time', 'Speed Up'],
                                                     ping_timeout=PING_TIMEOUT)
        self._cur_time = 0
        self._last_ping_succ_time = {}

    def rand_pos(self):
        return random.randint(0, XMAX), random.randint(0, YMAX)

    def run(self):
        ns = self.ns
        ns.packet_loss_ratio = PACKET_LOSS_RATIO

        router1 = ns.add("router", *self.rand_pos(), radio_range=RADIO_RANGE)
        router1_addr = self.expect_node_mleid(router1, 10)

        router2 = ns.add("router", *self.rand_pos(), radio_range=RADIO_RANGE)
        med = ns.add("med", *self.rand_pos(), radio_range=RADIO_RANGE)
        ns.set_child_timeout(med, PING_INTERVAL * 3)

        sed = ns.add("sed", *self.rand_pos(), radio_range=RADIO_RANGE)
        ns.set_poll_period(sed, 60)
        ns.set_child_timeout(sed, PING_INTERVAL * 3)

        for nodeid in (med, sed):
            self._last_ping_succ_time[nodeid] = 0
            ns.ping(nodeid, router1_addr, datasize=PING_DATA_SIZE, count=TOTAL_SIMULATION_TIME // PING_INTERVAL,
                    interval=PING_INTERVAL)

        t0 = time.time()

        for _ in range(TOTAL_SIMULATION_TIME // MOVE_INTERVAL):
            self.ns.go(MOVE_INTERVAL)
            self._cur_time += MOVE_INTERVAL

            self._collect_pings()

            for nodeid in (router1, router2, med, sed):
                self.ns.move(nodeid, *self.rand_pos())

        duration = time.time() - t0

        self.result.append_row('%ds' % TOTAL_SIMULATION_TIME, '%ds' % duration,
                               '%d' % (TOTAL_SIMULATION_TIME / duration))
        self.result.fail_if(TOTAL_SIMULATION_TIME / duration < 3000, "Speed Up < 3000")
        self.result.fail_if(self._last_ping_succ_time[med] < self._cur_time - 86400,
                            "MED not connected for a long time")
        self.result.fail_if(self._last_ping_succ_time[sed] < self._cur_time - 86400,
                            "SED not connected for a long time")

    def _collect_pings(self):
        for srcid, dstaddr, _, delay in self.ns.pings():
            if delay >= PING_TIMEOUT:
                # ignore failed pings
                continue

            self._last_ping_succ_time[srcid] = self._cur_time


if __name__ == '__main__':
    LongDurationStressTest().run()
