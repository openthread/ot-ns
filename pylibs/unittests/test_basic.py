#!/usr/bin/env python3
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

import sys
import unittest
from typing import Dict

from OTNSTestCase import OTNSTestCase
from otns.cli import OTNS


class BasicTests(OTNSTestCase):

    def testOneRouter(self):
        for i in range(100):
            print(f"testOneRouter round {i + 1}", file=sys.stderr)
            r = self.ns.add("router")
            self.assertEqual(1, r)
            self.ns.go(3)
            self.assertNodeState(r, "leader")
            self.assertFormPartitons(1)
            self.ns.delete(r)

    def testGetSetSpeed(self):
        ns = self.ns
        self.assertEqual(ns.speed, OTNS.MAX_SIMULATE_SPEED)
        ns.speed = 2
        self.assertEqual(ns.speed, 2)
        ns.speed = float('inf')
        self.assertEqual(ns.speed, OTNS.MAX_SIMULATE_SPEED)

    def testGetSetMDR(self):
        ns = self.ns
        assert ns.packet_loss_ratio == 0
        ns.packet_loss_ratio = 0.5
        assert ns.packet_loss_ratio == 0.5
        ns.packet_loss_ratio = 1
        assert ns.packet_loss_ratio == 1
        ns.packet_loss_ratio = 2
        assert ns.packet_loss_ratio == 1

    def testOneNode(self):
        ns = self.ns
        ns.add("router")
        ns.go(10)
        self.assertFormPartitons(1)

    def testAddNode(self):
        ns = self.ns
        ns.add("router")
        ns.go(10)
        self.assertFormPartitons(1)

        ns.add("router")
        ns.add("fed")
        ns.add("med")
        ns.add("sed")
        ns.go(100)
        self.assertFormPartitons(1)

    def testDelNode(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        ns.go(10)
        self.assertFormPartitons(1)
        ns.delete(1)
        ns.go(10)
        self.assertTrue(len(ns.nodes()) == 1 and 1 not in ns.nodes())

    def testMDREffective(self):
        ns = self.ns
        ns.packet_loss_ratio = 1
        self.assertTrue(ns.packet_loss_ratio, 1)
        ns.add("router")
        ns.add("router")
        ns.add("router")
        ns.go(1000)
        self.assertFormPartitons(3)

    def testRadioInRange(self):
        ns = self.ns
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range - 1, radio_range=radio_range)
        ns.go(100)
        self.assertFormPartitons(1)

    def testRadioNotInRange(self):
        ns = self.ns
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range + 1, radio_range=radio_range)
        ns.go(100)
        self.assertFormPartitons(2)

    def testNodeFailRecover(self):
        ns = self.ns
        ns.add("router")
        fid = ns.add("router")
        ns.go(100)
        self.assertFormPartitons(1)

        ns.radio_off(fid)
        ns.go(240)
        print(ns.partitions())
        self.assertFormPartitons(2)

        ns.radio_on(fid)
        ns.go(100)
        self.assertFormPartitons(1)

    def testFailTime(self):
        ns = self.ns
        id = ns.add("router")
        ns.radio_set_fail_time(id, fail_time=(2, 10))
        total_count = 0
        failed_count = 0
        for i in range(1000):
            ns.go(1)
            nodes = ns.nodes()
            failed = nodes[id]['failed']
            total_count += 1
            failed_count += failed

        self.assertAlmostEqual(failed_count / total_count, 0.2, delta=0.1)

    def testCliCmd(self):
        ns = self.ns
        id = ns.add("router")
        ns.go(10)
        self.assertTrue(ns.get_state(id), 'leader')

    def testCounters(self):
        ns = self.ns

        def assert_increasing(c0: Dict[str, int], c1: Dict[str, int]):
            for k0, v0 in c0.items():
                self.assertGreaterEqual(c1.get(k0, 0), v0)
            for k1, v1 in c1.items():
                self.assertGreaterEqual(v1, c0.get(k1, 0))

        c0 = counters = ns.counters()
        self.assertTrue(counters)
        self.assertTrue(all(x == 0 for x in counters.values()))
        ns.add("router")
        ns.add("router")
        ns.add("router")

        ns.go(10)
        c10 = ns.counters()
        assert_increasing(c0, c10)

        ns.go(10)
        c20 = ns.counters()
        assert_increasing(c10, c20)


if __name__ == '__main__':
    unittest.main()
