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
# ML-EID Connectivity Stress Test:
#   Nodes pings the BR by it's MLEID and measure the connectivity
# Topology:
#   Router xROUTER_COUNT
#   FED xFED_COUNT
#   MED xMED_COUNT
#   SED xSED_COUNT
# Fault Injections:
#   Nodes are constantly moving
#   Nodes fail for 30s in every 600s
#   Packet Loss Ratio set to 0.1
# Pass Criteria:
#   Max Delay (over all nodes) <= MAX_DELAY_TIME s
#   Avg Delay (over all nodes) <= MAX_AVG_DELAY_TIME s
#

import logging
import os
import random

from BaseStressTest import BaseStressTest

ROUTER_COUNT = 35
FED_COUNT = 5
MED_COUNT = 5
SED_COUNT = 5
TOTAL_NODE_COUNT = ROUTER_COUNT + FED_COUNT + MED_COUNT + SED_COUNT

RADIO_RANGE = 400
XMAX = 1000
YMAX = 1000

# Note: the "delay time" for a node is the time period since contact with the BR by means of ping was
# last successful.
TOTAL_SIMULATION_TIME = 3600 * int(os.getenv("STRESS_LEVEL", 1))
MAX_DELAY_TIME = 1800 # seconds - the max allowable delay time for any node.
MAX_AVG_DELAY_TIME = 1000 # seconds - the max allowable average of all nodes' delay times.
MOVE_INTERVAL = 60 # seconds
PING_INTERVAL = 10 # seconds
PING_DATA_SIZE = 32 # bytes
FAIL_DURATION = 30 # seconds (of failure during one FAIL_INTERVAL)
FAIL_INTERVAL = 600 # seconds
MOVE_COUNT = 3 # number of nodes moved per move-interval

BR = None  # the Border Router

SED_PULL_PERIOD = 1 # seconds


class MleidConnectivityStressTest(BaseStressTest):
    SUITE = 'connectivity'

    def __init__(self):
        super(MleidConnectivityStressTest, self).__init__("ML-EID Connectivity Test",
                                                          ["Simulation Time", "Max Delay", "Min Delay", "Avg Delay"])
        self._last_ping_succ_time = {}
        self._ping_fail_count = 0
        self._ping_succ_count = 0

    def run(self):
        ns = self.ns
        ns.packet_loss_ratio = 0.1
        ns.radiomodel = 'MIDisc'
        #ns.watch_default('warn') # enable OT node warnings or higher to be printed.
        ns.config_visualization(broadcast_message=False)

        assert ROUTER_COUNT >= 1
        BR = ns.add("router", x=random.randint(0, XMAX), y=random.randint(0, YMAX), radio_range=RADIO_RANGE)
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
            ns.go(PING_INTERVAL/TOTAL_NODE_COUNT) # spread out pings over time within interval.

        for _ in range(TOTAL_SIMULATION_TIME // MOVE_INTERVAL):
            nodeids = list(range(1, TOTAL_NODE_COUNT + 1))
            for nodeid in random.sample(nodeids, min(MOVE_COUNT, len(nodeids))):
                ns.move(nodeid, random.randint(0, XMAX), random.randint(0, YMAX))
            ns.go(MOVE_INTERVAL)
            self._collect_pings(MOVE_INTERVAL)

        ns.go(MOVE_INTERVAL)
        self._collect_pings(PING_INTERVAL)

        delays = [TOTAL_SIMULATION_TIME - self._last_ping_succ_time.get(nodeid, 0) for nodeid in
                  range(1, TOTAL_NODE_COUNT + 1)]
        logging.debug("_last_ping_succ_time %s delays %s", self._last_ping_succ_time, delays)
        avg_delay = sum(delays) / TOTAL_NODE_COUNT
        self.result.append_row("%dh" % (TOTAL_SIMULATION_TIME // 3600),
                               '%ds' % max(delays), '%ds' % min(delays), '%ds' % avg_delay)
        self.result.fail_if(avg_delay > MAX_AVG_DELAY_TIME, "Avg Delay (%ds)> %ds" % (avg_delay, MAX_AVG_DELAY_TIME))
        self.result.fail_if(max(delays) > MAX_DELAY_TIME, "Max Delay (%ds)> %ds" % (max(delays), MAX_DELAY_TIME))

    def _collect_pings(self, lastIntervalSec):
        current_sim_time = self.ns.time
        for srcid, dstaddr, _, delay in self.ns.pings():
            if delay >= 10000:
                # ignore failed pings
                self._ping_fail_count += 1
                continue

            self._ping_succ_count += 1
            self._last_ping_succ_time[srcid] = current_sim_time - lastIntervalSec


if __name__ == '__main__':
    MleidConnectivityStressTest().run()
