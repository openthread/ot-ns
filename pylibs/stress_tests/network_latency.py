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
# Ping Latency Stress Test:
#   12 routers ping each other to measure the ping delay of different hops and different data size
# Topology:
#   Routers x6 forming Circle I
#   Routers x6 forming Circle II
# Fault Injections:
#   None
# Pass Criteria:
#   Max ping latency < (3 * datasize + 500 * (datasize>=128)) ms
#   (for fragmented ping messages there's an added 500 ms to cater for fragment losses, MAC retries, hidden node
#    situations, etc.)
import logging
import math

from BaseStressTest import BaseStressTest

RADIUS = 150
RADIO_RANGE = int(RADIUS * 2.5)
REPEAT = 3


class StressTest(BaseStressTest):
    SUITE = 'network-latency'

    LEFT_NODES = (1, 2, 3, 4, 5, 6)
    RIGHT_NODES = (7, 8, 9, 10, 11, 12)
    LEFT_ROUTER = 6
    RIGHT_ROUTER = 7

    def __init__(self):
        super(StressTest, self).__init__("Ping Latency Test",
                                         ["Data Size", "Hop x1 Latency", "Hop x2 Latency", "Hop x3 Latency"])

        self._ping_latencys_by_datasize = {}
        self.ns.loglevel = 'warn'
        self.ns.radiomodel = 'MIDisc'
        self.ns.set_title("Network (Ping) Latency test - setup phase")

    def add_6_nodes(self, x, y, start_angle):
        for i in range(6):
            angle = start_angle - math.pi / 3 * i
            nid = self.ns.add("router", int(x + math.cos(angle) * RADIUS), int(y + math.sin(angle) * RADIUS),
                              radio_range=RADIO_RANGE)
            self.ns.set_router_upgrade_threshold(nid, 32)
            self.ns.set_router_downgrade_threshold(nid, 33)

    def ping_go(self, src: int, dst: int, datasize: int):
        assert datasize >= 4
        self.ns.ping(src, dst, addrtype='rloc', datasize=datasize)
        self.ns.go(1)

    def pings_1_hop(self, datasize: int):
        # from left to other left node
        for n1 in self.LEFT_NODES:
            for n2 in self.LEFT_NODES:
                if n1 != n2:
                    self.ping_go(n1, n2, datasize)

        # from right to other right now
        for n1 in self.RIGHT_NODES:
            for n2 in self.RIGHT_NODES:
                if n1 != n2:
                    self.ping_go(n1, n2, datasize)

    def pings_2_hop(self, datasize: int):
        # ping 2 hops
        for n1 in self.LEFT_NODES:
            for n2 in self.RIGHT_NODES:
                if (n1 != self.LEFT_ROUTER and n2 == self.RIGHT_ROUTER) or \
                        (n1 == self.LEFT_ROUTER and n2 != self.RIGHT_ROUTER):
                    self.ping_go(n1, n2, datasize)
                    self.ping_go(n2, n1, datasize)

    def pings_3_hop(self, datasize: int):
        for n1 in self.LEFT_NODES:
            for n2 in self.RIGHT_NODES:
                if n1 == self.LEFT_ROUTER or n2 == self.RIGHT_ROUTER:
                    continue
                self.ping_go(n1, n2, datasize)
                self.ping_go(n2, n1, datasize)

    def collect_pings(self, hop: int) -> None:
        pings = self.ns.pings()
        for srcid, dst, datasize, latency in pings:
            if latency == 10000: # skip the failed pings (packet lost)
                continue
            latencys = self._ping_latencys_by_datasize.setdefault(datasize, [[0, 0], [0, 0], [0, 0]])
            latency_info = latencys[hop - 1]
            latency_info[0] += 1
            latency_info[1] += latency
            logging.debug(f'ping from {srcid} to {dst} datasize {datasize} latency {latency}')

    def run(self):
        ns = self.ns
        Y = 300
        X1 = 300
        X2 = X1 + RADIUS * 4
        self.add_6_nodes(X1, Y, -math.pi / 3)
        self.add_6_nodes(X2, Y, math.pi)

        self.expect_all_nodes_become_routers()

        # wait for a period of time so that all routes are discovered
        for i in range(5):
            ns.go(90)
            self.pings_3_hop(datasize=4)
            ns.go(10)
            ns.pings()  # throw away ping results

        # now start the real tests
        logging.debug("real test starts...")
        ns.set_title("Network (Ping) Latency test - test phase")
        for _ in range(REPEAT):
            self.ns.radiomodel = 'MIDisc' # reset the radiomodel with new static random deviations.
            for datasize in (32, 64, 128, 256, 512, 1024):
                self.pings_1_hop(datasize)
                ns.go(10)  # wait for all ping replies
                self.collect_pings(hop=1)

                self.pings_2_hop(datasize)
                ns.go(10)  # wait for all ping replies
                self.collect_pings(hop=2)

                self.pings_3_hop(datasize)
                ns.go(10)  # wait for all ping replies
                self.collect_pings(hop=3)

        for datasize, latencys in sorted(self._ping_latencys_by_datasize.items()):
            row = ['%dB' % datasize]
            for n, s in latencys:
                if n > 0:
                    latency = s / n
                else:
                    latency = 0
                row.append('%dms' % latency)
                maxlatency = 3 * datasize + 500 * (datasize>=128)
                self.result.fail_if(latency > maxlatency,
                     f"average ping latency (for datasize={datasize}) is {latency} ms > {maxlatency} ms")

            self.result.append_row(*row)


if __name__ == '__main__':
    StressTest().run()
