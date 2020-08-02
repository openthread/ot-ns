#!/usr/bin/env python3
#
# Copyright (c) 2020, The OTNS Authors.
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
# ML-EID Connectivity Stress Test:
#   Nodes pings the BR by it's MLEID and measure the connectivity
# Topology:
#   Router x20
#   FED x10
#   MED x10
#   SED x10
# Fault Injections:
#   Nodes are constantly moving
#   Nodes fail for 60s in every 600s
#   Packet Loss Ratio set to 0.2
# Pass Criteria:
#   Max Delay (of all nodes) <= 3600s
#
import logging
import os
import random

from BaseStressTest import BaseStressTest

ROUTER_COUNT = 20

FED_COUNT = 10
MED_COUNT = 10
SED_COUNT = 10
TOTAL_NODE_COUNT = ROUTER_COUNT + FED_COUNT + MED_COUNT + SED_COUNT

RADIO_RANGE = 400

XMAX = 1000
YMAX = 1000

TOTAL_SIMULATION_TIME = 3600 * int(os.getenv("STRESS_LEVEL", 1))
MOVE_INTERVAL = 60
PING_INTERVAL = 60
PING_DATA_SIZE = 32

FAIL_DURATION = 60
FAIL_INTERVAL = 600
MOVE_COUNT = 5

BR = None  # the Border Router

SED_PULL_PERIOD = 1


class MleidConnectivityStressTest(BaseStressTest):
    SUITE = 'connectivity'

    def __init__(self):
        super(MleidConnectivityStressTest, self).__init__("ML-EID Connectivity Test",
                                                          ["Simulation Time", "Max Delay", "Min Delay", "Avg Delay"])
        self._last_ping_succ_time = {}
        self._cur_time = 0
        self._ping_fail_count = 0
        self._ping_succ_count = 0

    def run(self):
        ns = self.ns
        ns.packet_loss_ratio = 0.2
        ns.config_visualization(broadcast_message=False)

        assert ROUTER_COUNT >= 1
        BR = ns.add("router", x=random.randint(0, XMAX), y=random.randint(0, YMAX))
        ns.radio_set_fail_time(BR, fail_time=(FAIL_DURATION, FAIL_INTERVAL))

        BR_ADDR = self.expect_node_mleid(BR, 10)

        for i in range(ROUTER_COUNT - 1):
            nid = ns.add("router", x=random.randint(0, XMAX), y=random.randint(0, YMAX), radio_range=RADIO_RANGE)
            ns.radio_set_fail_time(nid, fail_time=(FAIL_DURATION, FAIL_INTERVAL))

        for i in range(FED_COUNT):
            nid = ns.add("fed", x=random.randint(0, XMAX), y=random.randint(0, YMAX), radio_range=RADIO_RANGE)
            ns.radio_set_fail_time(nid, fail_time=(FAIL_DURATION, FAIL_INTERVAL))

        for i in range(MED_COUNT):
            nid = ns.add("med", x=random.randint(0, XMAX), y=random.randint(0, YMAX), radio_range=RADIO_RANGE)
            ns.radio_set_fail_time(nid, fail_time=(FAIL_DURATION, FAIL_INTERVAL))

        for i in range(SED_COUNT):
            nid = ns.add("sed", x=random.randint(0, XMAX), y=random.randint(0, YMAX), radio_range=RADIO_RANGE)
            ns.radio_set_fail_time(nid, fail_time=(FAIL_DURATION, FAIL_INTERVAL))
            ns.set_poll_period(nid, SED_PULL_PERIOD)

        for nodeid in range(1, TOTAL_NODE_COUNT + 1):
            ns.ping(nodeid, BR_ADDR, datasize=PING_DATA_SIZE, count=TOTAL_SIMULATION_TIME // PING_INTERVAL,
                    interval=PING_INTERVAL)

        for _ in range(TOTAL_SIMULATION_TIME // MOVE_INTERVAL):
            nodeids = list(range(1, TOTAL_NODE_COUNT + 1))
            for nodeid in random.sample(nodeids, min(MOVE_COUNT, len(nodeids))):
                ns.move(nodeid, random.randint(0, XMAX), random.randint(0, YMAX))

            ns.go(MOVE_INTERVAL)
            self._collect_pings()

            self._cur_time += MOVE_INTERVAL

        ns.go(100)
        self._collect_pings()

        self._cur_time += 100

        delays = [TOTAL_SIMULATION_TIME - self._last_ping_succ_time.get(nodeid, 0) for nodeid in
                  range(1, TOTAL_NODE_COUNT + 1)]
        logging.debug("_last_ping_succ_time %s delays %s", self._last_ping_succ_time, delays)
        avg_delay = sum(delays) / TOTAL_NODE_COUNT
        self.result.append_row("%dh" % (TOTAL_SIMULATION_TIME // 3600),
                               '%ds' % max(delays), '%ds' % min(delays), '%ds' % avg_delay)
        self.result.fail_if(max(delays) > 3600, "Max Delay (%ds)> 3600s" % max(delays))

    def _collect_pings(self):
        for srcid, dstaddr, _, delay in self.ns.pings():
            if delay >= 10000:
                # ignore failed pings
                self._ping_fail_count += 1
                continue

            self._ping_succ_count += 1
            self._last_ping_succ_time[srcid] = self._cur_time


if __name__ == '__main__':
    MleidConnectivityStressTest().run()
