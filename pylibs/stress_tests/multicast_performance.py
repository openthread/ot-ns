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
# CoAP Multicast Stress Test:
#   The Border Router multicasts COAP messages to all nodes and measure the coverage and delay.
# Topology:
#   Router xROUTER_COUNT (one being Border Router)
#   FED xFED_COUNT
#   MED xMED_COUNT
#   SED xSED_COUNT
# Fault Injections:
#   No global packet loss set - uses radiomodel MutualInterference.
# Pass Criteria:
#   Average Coverage >= 70%
#   Average Delay < 200ms (2000ms for SEDs)
#
import os

from BaseStressTest import BaseStressTest

ROUTER_COUNT = 8
FED_COUNT = 8
MED_COUNT = 8
SED_COUNT = 8

TOTAL_NODE_COUNT = ROUTER_COUNT + FED_COUNT + MED_COUNT + SED_COUNT

RADIO_RANGE = 210

XMAX = 1000
YMAX = 1000

TOTAL_SIMULATION_TIME = 100 * int(os.getenv("STRESS_LEVEL", "1"))

BR = None  # the Border Router
SVR1, SVR1_DATA = "svr1", "svr1"
BR_ADDR = 'fdde:ad00:beef:0:0:ff:fe00:fc10'
LINK_LOCAL_ALL_THREAD_NODES_MULTICAST_ADDRESS = 'ff33:0040:fdde:ad00:beef:0000:0000:0001'

SED_POLL_PERIOD = 1


class StressTest(BaseStressTest):
    SUITE = 'multicast-performance'

    def __init__(self):
        super(StressTest, self).__init__("Multicast Performance Test", ["Router", "FED", "MED", "SED"])
        self._last_ping_succ_time = {}
        self._cur_time = 0
        self._ping_fail_count = 0
        self._ping_succ_count = 0

    def to_hex_str(self, s: str):
        return ''.join(['%02x' % ord(c) for c in s])

    def run(self):
        ns = self.ns
        ns.coaps_enable()
        ns.packet_loss_ratio = 0.0

        assert ROUTER_COUNT >= 1
        BR = ns.add("router", x=200, y=200, radio_range=RADIO_RANGE)
        assert BR == 1
        ns.node_cmd(BR, "prefix add 2001:dead:beef:cafe::/64 paros med")
        ns.node_cmd(BR, f"service add 44970 {self.to_hex_str(SVR1)} {self.to_hex_str(SVR1_DATA)}")
        ns.node_cmd(BR, "netdata register")

        self.expect_node_addr(BR, BR_ADDR, 10)

        for i in range(1, ROUTER_COUNT):
            ns.add("router", x=200 + 100 * i, y=200, radio_range=RADIO_RANGE)

        for i in range(FED_COUNT):
            ns.add("fed", x=200 + 100 * i, y=100, radio_range=RADIO_RANGE)

        for i in range(MED_COUNT):
            ns.add("med", x=150 + 100 * i, y=300, radio_range=RADIO_RANGE)

        for i in range(SED_COUNT):
            nid = ns.add("sed", x=200 + 100 * i, y=400, radio_range=RADIO_RANGE)
            ns.set_poll_period(nid, SED_POLL_PERIOD)

        for nid in range(1, TOTAL_NODE_COUNT + 1):
            ns.node_cmd(nid, f'coap start')
            ns.node_cmd(nid, f'coap resource test')

        ns.go(120)

        SEND_INTERVAL = 10

        router_delays = []
        fed_delays = []
        med_delays = []
        sed_delays = []
        router_coverages = []
        fed_coverages = []
        med_coverages = []
        sed_coverages = []

        for _ in range(TOTAL_SIMULATION_TIME // SEND_INTERVAL):
            ns.node_cmd(BR, f'coap post {LINK_LOCAL_ALL_THREAD_NODES_MULTICAST_ADDRESS} test non turnonthelightplease')

            ns.go(SEND_INTERVAL)

            multicast_msg = None
            coaps = ns.coaps()
            send_time = None
            req_received = {}

            for msg in coaps:
                if msg['code'] == 2 and msg['uri'] == 'test':
                    assert multicast_msg is None, (msg, multicast_msg)
                    multicast_msg = msg
                    req_received = {m['dst']: m['time'] for m in multicast_msg['receivers']}
                    send_time = msg['time']

            assert multicast_msg is not None

            router_coverages.append(
                sum(1 for nid in range(1, ROUTER_COUNT + 1) if nid in req_received) / (ROUTER_COUNT - 1))
            fed_coverages.append(
                sum(1 for nid in range(ROUTER_COUNT + 1, ROUTER_COUNT + FED_COUNT + 1) if nid in req_received) /
                FED_COUNT)
            med_coverages.append(
                sum(1 for nid in range(ROUTER_COUNT + FED_COUNT + 1, ROUTER_COUNT + FED_COUNT + MED_COUNT + 1)
                    if nid in req_received) / MED_COUNT)
            sed_coverages.append(
                sum(1 for nid in range(ROUTER_COUNT + FED_COUNT + MED_COUNT + 1, ROUTER_COUNT + FED_COUNT + MED_COUNT +
                                       SED_COUNT + 1) if nid in req_received) / SED_COUNT)

            router_delays += [time - send_time for nid, time in req_received.items() if 1 <= nid <= ROUTER_COUNT]
            fed_delays += [
                time - send_time
                for nid, time in req_received.items()
                if ROUTER_COUNT + 1 <= nid <= ROUTER_COUNT + FED_COUNT
            ]
            med_delays += [
                time - send_time
                for nid, time in req_received.items()
                if ROUTER_COUNT + FED_COUNT + 1 <= nid <= ROUTER_COUNT + FED_COUNT + MED_COUNT
            ]
            sed_delays += [
                time - send_time
                for nid, time in req_received.items()
                if ROUTER_COUNT + FED_COUNT + MED_COUNT + 1 <= nid <= ROUTER_COUNT + FED_COUNT + MED_COUNT + SED_COUNT
            ]

        def format_delay(coverages, delays):
            if not delays:
                return '-'

            avg = int(self.avg(delays) // 1000)
            _max = int(max(delays) // 1000)
            return f'cov:{int(self.avg(coverages) * 100)}%%, avg:{avg}ms, max:{_max}ms'

        self.result.append_row(format_delay(router_coverages, router_delays), format_delay(fed_coverages, fed_delays),
                               format_delay(med_coverages, med_delays), format_delay(sed_coverages, sed_delays))

        self.result.fail_if(self.avg(router_coverages) < 0.7, 'Router coverage < 70%')
        self.result.fail_if(self.avg(fed_coverages) < 0.7, 'FED coverage < 70%')
        self.result.fail_if(self.avg(med_coverages) < 0.7, 'MED coverage < 70%')
        self.result.fail_if(self.avg(sed_coverages) < 0.7, 'SED coverage < 70%')

        self.result.fail_if(self.avg(router_delays) / 1000 > 200, 'Router avg. delay > 200ms')
        self.result.fail_if(self.avg(fed_delays) / 1000 > 200, 'FED avg. delay > 200ms')
        self.result.fail_if(self.avg(med_delays) / 1000 > 200, 'MED avg. delay > 200ms')
        self.result.fail_if(self.avg(sed_delays) / 1000 > 2000, 'SED avg. delay > 2000ms')


if __name__ == '__main__':
    StressTest().run()
