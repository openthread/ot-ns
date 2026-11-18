#!/usr/bin/env python3
# Copyright (c) 2023-2024, The OTNS Authors.
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
import logging
import math
import random
import unittest
from typing import Dict

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS


class CslTests(OTNSTestCase):

    def testSsedConnectsToParent(self):
        ns = self.ns

        # add SSED
        nodeid = ns.add("ssed", 220, 100)
        ns.node_cmd(nodeid, "csl period 288000")
        ns.go(10)

        # Parent comes in, SSED connects
        ns.add("router", 100, 100)
        ns.go(10)
        self.assertFormPartitions(1)

        # SSED can ping parent
        ns.ping(1, 2)
        ns.go(1)
        ns.ping(1, 2)
        ns.go(1)
        self.assertPings(ns.pings(), 2, max_delay=2000, max_fails=1)

    def testOneParentMultiCslChildren(self):
        ns = self.ns

        # setup a Parent Router with N SSED Children with different CSL Periods.
        N = 8
        ns.add("router", 150, 100)
        # below CSL periods to test (given in units of 160 us)
        aCslPeriods = [3100, 500, 7225, 1024, 3125, 3124, 250, 5999, 777, 1024]
        for n in range(0, N):
            nodeid = ns.add("ssed", 80 + n * 20, 150)
            ns.node_cmd(nodeid, "csl period " + str(aCslPeriods[n] * 160))
            ns.go(1)
        ns.go(45)
        self.assertFormPartitions(1)

        for k in range(0, 5):
            # do some pings
            for n in range(0, N):
                ns.ping(1, 2 + n)
                ns.go(2)
                ns.ping(2 + n, 1)
                ns.go(2)

            # long wait and some pings
            ns.go(300)
            for n in range(0, N):
                ns.ping(1, 2 + n)
                ns.go(20)
                ns.ping(2 + n, 1)
                ns.go(20)

            # test ping results
            self.assertPings(ns.pings(), N * 4, max_delay=3000, max_fails=1)

    def testCslReenable(self):
        ns = self.ns

        # setup a Parent Router with SSED Child
        ns.add("router", 100, 100)
        ns.go(10)
        nodeid = ns.add("ssed", 200, 100)
        ns.node_cmd(nodeid, "csl period 288000")
        ns.go(10)
        self.assertFormPartitions(1)

        # SSED pings parent
        for n in range(0, 15):
            ns.ping(2, 1, datasize=n + 10)
            ns.go(5)
        self.assertPings(ns.pings(), 15, max_delay=3000, max_fails=1)

        # parent pings SSED
        for n in range(0, 15):
            ns.ping(1, 2, datasize=n + 10)
            ns.go(5)
        self.assertPings(ns.pings(), 15, max_delay=3000, max_fails=1)

        for k in range(0, 4):
            # disable CSL
            ns.node_cmd(nodeid, "csl period 0")
            ns.go(1)

            # SSED pings parent
            for n in range(0, 15):
                ns.ping(2, 1, datasize=n + 10)
                ns.go(5)
            self.assertPings(ns.pings(), 15, max_delay=3000, max_fails=1)

            # re-enable CSL
            ns.node_cmd(nodeid, "csl period 144000")
            ns.go(1)

            # SSED pings parent
            for n in range(0, 15):
                ns.ping(2, 1, datasize=n + 10)
                ns.go(5)
            self.assertPings(ns.pings(), 15, max_delay=3000, max_fails=1)

            # parent pings SSED
            for n in range(0, 15):
                ns.ping(1, 2, datasize=n + 10)
                ns.go(5)
            self.assertPings(ns.pings(), 15, max_delay=3000, max_fails=1)

    def testCslParameters(self):
        ns: OTNS = self.ns
        ns.add('router', x=100, y=100)
        ns.go(10)
        ns.add('router', x=150, y=100)
        self.assertEqual(['20'], ns.cmd('rfsim 1 cslacc'))
        self.assertEqual(['10'], ns.cmd('rfsim 2 cslunc'))

        ns.cmd('rfsim 1 cslacc 65')
        self.assertEqual(['65'], ns.cmd('rfsim 1 cslacc'))
        self.assertEqual(['10'], ns.cmd('rfsim 2 cslunc'))

        ns.cmd('rfsim 2 cslunc 223')
        self.assertEqual(['223'], ns.cmd('rfsim 2 cslunc'))

        ns.go(15)
        self.assertFormPartitions(1)

        ns.add('ssed', x=125, y=150)
        ns.go(10)
        self.assertFormPartitions(1)


class CslRxPairingTests(OTNSTestCase):

    def setUp(self) -> None:
        self.ns = OTNS(otns_args=['-log', 'debug', '-seed', '102'])

    def testConcurrentSsedTraffic(self):
        """Overlapping receptions must not stop delivery to a CSL child (#870)."""
        ns = self.ns
        # Ten SSEDs fit the default parent child table. Distinct periods and a
        # fixed traffic seed exercise overlapping receptions reproducibly.
        periods_ms = [400, 200, 100, 500, 100, 300, 100, 500, 300, 200]
        parent = ns.add('router', 300, 300)
        ns.go(15)
        children = []
        for index, period_ms in enumerate(periods_ms):
            angle = 2 * math.pi * index / len(periods_ms)
            child = ns.add('ssed', int(300 + 120 * math.cos(angle)), int(300 + 120 * math.sin(angle)))
            ns.node_cmd(child, f'csl period {period_ms * 1000}')
            children.append(child)
            ns.go(1.5)
        ns.go(10)
        self.assertFormPartitions(1)
        for child in children:
            self.assertNodeState(child, 'child')

        # Schedule concurrent downlink CoAP traffic for 60 s, averaging 0.3
        # arrivals per CSL period per child. Use simulated time throughout.
        rng = random.Random(990498)
        arrivals = []
        for child, period_ms in zip(children, periods_ms):
            when_us = 0.0
            while True:
                when_us += rng.expovariate(0.3 / (period_ms * 1000))
                if when_us >= 60000000:
                    break
                arrivals.append((round(when_us), child))
        arrivals.sort()

        ns.coaps_enable()
        start_time = ns.time
        now_us = 0
        accepted = {child: 0 for child in children}
        no_bufs = 0
        for when_us, child in arrivals:
            if when_us > now_us:
                ns.cmd(f'go {when_us - now_us}us')
                now_us = when_us
            try:
                ns.cmd(f'send coap con {parent} {child}')
                accepted[child] += 1
            except errors.OTNSCliError as exc:
                if 'NoBufs' not in str(exc):
                    raise
                no_bufs += 1
        ns.go(10)

        # Individual losses or transient NoBufs are not failures by themselves.
        # Every child must still receive traffic in the final 20 s or drain
        # interval; a stale receive timestamp can otherwise leave a child stuck.
        progressed = set()
        for message in ns.coaps() or []:
            if message.get('src') != parent or message.get('type') != 0:
                continue
            for receiver in message.get('receivers') or []:
                if receiver['time'] / 1e6 - start_time >= 40:
                    progressed.add(receiver['dst'])
        self.assertTrue(all(accepted.values()), f'No accepted sends to some children: {accepted}')
        missing = set(children) - progressed
        self.assertFalse(missing, f'No late delivery to children {sorted(missing)}; NoBufs={no_bufs}')


if __name__ == '__main__':
    unittest.main()
