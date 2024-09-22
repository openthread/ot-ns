#!/usr/bin/env python3
# Copyright (c) 2020-2024, The OTNS Authors.
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

import tracemalloc
import unittest

from OTNSTestCase import OTNSTestCase
from otns.cli import OTNS

tracemalloc.start()


class CommissioningTests(OTNSTestCase):

    def setUp(self) -> None:
        self.ns = OTNS(otns_args=['-ot-script', 'none', '-log', 'debug', '-pcap', 'wpan-tap', '-seed', '23'])
        self.ns.speed = float('inf')

    def setFirstNodeDataset(self, n1) -> None:
        self.ns.node_cmd(n1, "dataset init new")
        self.ns.node_cmd(n1, "dataset networkkey 00112233445566778899aabbccddeeff")  # allow easy Wireshark dissecting
        self.ns.node_cmd(n1, "dataset commit active")

    def testRawNoSetup(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        self.go(250)
        # can not form any partition without setting network parameters
        self.assertTrue(0 in ns.partitions())

    def testRawSetup(self):
        ns = self.ns
        ns.watch_default('trace')  # for most detailed radio logs
        n1 = ns.add("router")
        n2 = ns.add("router")
        n3 = ns.add("router")

        # n1 with full dataset becomes Leader.
        ns.config_dataset(n1,
                          channel=21,
                          panid=0xface,
                          extpanid="dead00beef00cafe",
                          networkkey="00112233445566778899aabbccddeeff",
                          active_timestamp=1719172243,
                          network_name="test",
                          set_remaining=True)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)

        # n2, n3 with partial dataset - will wait for Leader to join to.
        for id in (n2, n3):
            ns.config_dataset(id,
                              network_name="test",
                              networkkey="00112233445566778899aabbccddeeff",
                              set_remaining=False)
            ns.ifconfig_up(id)
            ns.thread_start(id)

        self.go(50)
        self.assertFormPartitions(1)

    def testCommissioningOneHop(self):
        ns = self.ns

        n1 = ns.add("router")
        n2 = ns.add("router")

        self.setFirstNodeDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        self.go(35)
        self.assertTrue(ns.get_state(n1) == "leader")

        ns.commissioner_start(n1)
        ns.commissioner_joiner_add(n1, "*", "TEST123")

        ns.ifconfig_up(n2)
        ns.joiner_start(n2, "TEST123")
        self.go(100)
        ns.thread_start(n2)
        self.go(100)
        c = ns.counters()
        print('counters', c)
        joins = ns.joins()
        print('joins', joins)
        self.assertFormPartitions(1)
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success

    def testCommissioningOneHopWithSteeringDataAndDomain(self):
        ns = self.ns

        n1 = ns.add("router")
        n2 = ns.add("med")
        joiner_eui = ns.node_cmd(2, "eui64")[0]

        self.setFirstNodeDataset(n1)
        # Set a non-default Thread Domain Name. This will be sent in the MLE Discovery Response.
        ns.node_cmd(1, 'domainname TestingDomainThr')
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        self.go(35)
        self.assertTrue(ns.get_state(n1) == "leader")

        ns.commissioner_start(n1)
        ns.commissioner_joiner_add(n1, joiner_eui, "J01NM3")

        ns.ifconfig_up(n2)
        ns.joiner_start(n2, "J01NM3")
        self.go(100)
        ns.thread_start(n2)
        self.go(100)
        c = ns.counters()
        print('counters', c)
        joins = ns.joins()
        print('joins', joins)
        self.assertFormPartitions(1)
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success

    def testCommissioningThreeHop(self):
        ns = self.ns
        ns.radiomodel = 'MIDisc'

        n1 = ns.add("router", radio_range=110)
        n2 = ns.add("router", radio_range=110)
        n3 = ns.add("router", radio_range=110)
        n4 = ns.add("router", radio_range=110)

        self.setFirstNodeDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        self.go(35)
        self.assertTrue(ns.get_state(n1) == "leader")
        ns.commissioner_start(n1)
        ns.commissioner_joiner_add(n1, "*", "TEST123")

        ns.ifconfig_up(n2)
        ns.joiner_start(n2, "TEST123")
        self.go(40)
        ns.thread_start(n2)

        ns.ifconfig_up(n3)
        ns.joiner_start(n3, "TEST123")
        self.go(40)
        ns.thread_start(n3)

        ns.ifconfig_up(n4)
        ns.joiner_start(n4, "TEST123")
        self.go(100)
        ns.thread_start(n4)

        self.go(100)
        c = ns.counters()
        print('counters', c)
        joins = ns.joins()
        print('joins', joins)
        self.assertFormPartitions(1)
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success


if __name__ == '__main__':
    unittest.main()
