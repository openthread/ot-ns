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
#
import logging
import os
import unittest
from typing import Dict

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS


class BasicTests(OTNSTestCase):
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

    def testOneNodex100(self):
        if os.getenv("VIRTUAL_TIME_UART") != "1":
            self.skipTest("VIRTUAL_TIME_UART not enabled")
            return

        for i in range(100):
            logging.info("testOneNode round %d", i + 1)
            ns = self.ns
            ns.add("router")
            ns.go(3)
            self.assertFormPartitions(1)
            self.tearDown()
            self.setUp()

    def testAddNode(self):
        ns = self.ns
        ns.add("router")
        self.goConservative(3)
        self.assertFormPartitions(1)

        ns.add("router")
        ns.add("fed")
        ns.add("med")
        ns.add("sed")
        self.goConservative(33)
        self.assertFormPartitions(1)

    def testAddNodeWithID(self):
        ns = self.ns
        for new_id in [50, 55, 60]:
            nid = ns.add("router", id=new_id)
            self.assertEqual(nid, new_id)
            self.goConservative(1)

    def testAddNodeWithExistingID(self):
        ns = self.ns
        new_id = 50
        nid = ns.add("router", id=new_id)
        self.assertEqual(nid, new_id)
        self.goConservative(1)
        self.assertRaises(errors.OTNSCliError, lambda: ns.add("router", id=new_id))

    def testDelNode(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        self.goConservative(10)
        self.assertFormPartitions(1)
        ns.delete(1)
        self.goConservative(10)
        self.assertTrue(len(ns.nodes()) == 1 and 1 not in ns.nodes())

    def testDelManyNodes(self):
        ns = self.ns
        many = 32
        for i in range(many):
            ns.add("router", x=(i % 6) * 100, y=(i // 6) * 150)

        ns.go(10)
        for i in range(1, many + 1):
            ns.delete(i)
            ns.go(5)

        self.assertTrue(ns.nodes() == {})

    def testMDREffective(self):
        ns = self.ns
        ns.packet_loss_ratio = 1
        self.assertTrue(ns.packet_loss_ratio, 1)
        ns.add("router")
        ns.add("router")
        ns.add("router")
        self.goConservative(100)
        self.assertFormPartitions(3)

    def testRadioInRange(self):
        ns = self.ns
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range - 1, radio_range=radio_range)
        self.goConservative(10)
        self.assertFormPartitions(1)

    def testRadioNotInRange(self):
        ns = self.ns
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range + 1, radio_range=radio_range)
        self.goConservative(10)
        self.assertFormPartitions(2)

    def testNodeFailRecover(self):
        ns = self.ns
        ns.add("router")
        fid = ns.add("router")
        self.goConservative(10)
        self.assertFormPartitions(1)

        ns.radio_off(fid)
        self.goConservative(240)
        print(ns.partitions())
        self.assertFormPartitions(2)

        ns.radio_on(fid)
        self.goConservative(100)
        self.assertFormPartitions(1)

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
        self.goConservative(3)
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

        self.goConservative(10)
        c10 = ns.counters()
        assert_increasing(c0, c10)

        self.goConservative(10)
        c20 = ns.counters()
        assert_increasing(c10, c20)

    def testConfigVisualization(self):
        ns = self.ns
        vopts = ns.config_visualization()
        print('vopts', vopts)
        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertTrue(opt in vopts)

            set_vals = (False, True) if vopts[opt] else (True, False)
            for v in set_vals:
                vopts[opt] = v
                self.assertTrue(ns.config_visualization(**{opt: v}) == vopts)

        vopts = ns.config_visualization(broadcast_message=True, unicast_message=True, ack_message=True,
                                        router_table=True,
                                        child_table=True)

        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertTrue(vopts[opt])

        vopts = ns.config_visualization(broadcast_message=False, unicast_message=False, ack_message=False,
                                        router_table=False,
                                        child_table=False)

        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertFalse(vopts[opt])

    def testWithOTNS(self):
        """
        make sure OTNS works in with-statement
        """
        self.tearDown()

        with OTNS(otns_args=['-log', 'debug']) as ns:
            ns.add("router")

        # run a second time to make sure the previous simulation is properly terminated
        with OTNS(otns_args=['-log', 'debug']) as ns:
            ns.add("router")

    def testSetRouterUpgradeThreshold(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        self.assertEqual(16, ns.get_router_upgrade_threshold(nid))
        for val in range(0, 33):
            ns.set_router_upgrade_threshold(nid, val)
            self.assertEqual(val, ns.get_router_upgrade_threshold(nid))

    def testSetRouterUpgradeThresholdEffective(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        ns.go(10)
        self.assertNodeState(nid, 'leader')

        reed = ns.add("router")
        ns.set_router_upgrade_threshold(reed, 1)
        ns.go(100)
        self.assertNodeState(reed, 'child')

        ns.set_router_upgrade_threshold(reed, 2)
        ns.go(100)
        self.assertNodeState(reed, 'router')

    def testSetRouterDowngradeThreshold(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        self.assertEqual(23, ns.get_router_downgrade_threshold(nid))
        for val in range(0, 33):
            ns.set_router_downgrade_threshold(nid, val)
            self.assertEqual(val, ns.get_router_downgrade_threshold(nid))


if __name__ == '__main__':
    unittest.main()
