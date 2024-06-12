#!/usr/bin/env python3
# Copyright (c) 2020-2022, The OTNS Authors.
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
        self.ns = OTNS(otns_args=['-ot-script', 'none', '-log', 'debug', '-pcap', 'wpan-tap'])
        self.ns.speed = float('inf')

    def testRawNoSetup(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        self.go(10)
        # can not form any partition without setting network parameters
        self.assertTrue(0 in ns.partitions())

    def testRawSetup(self):
        ns = self.ns
        n1 = ns.add("router")
        n2 = ns.add("router")
        n3 = ns.add("router")

        # n1 with full dataset becomes Leader.
        ns.node_cmd(n1, "dataset init new")
        ns.node_cmd(n1, "dataset panid 0xface")
        ns.node_cmd(n1, "dataset extpanid dead00beef00cafe")
        ns.node_cmd(n1, "dataset networkkey 00112233445566778899aabbccddeeff")
        ns.node_cmd(n1, "dataset networkname test")
        ns.node_cmd(n1, "dataset channel 15")
        ns.node_cmd(n1, "dataset commit active")
        ns.ifconfig_up(n1)
        ns.thread_start(n1)

        # n2, n3 with partial dataset - if channel not given - will scan channels to find n1.
        # This can take some time and may fail even then (FIXME: find cause).
        # To prevent this failure, channel is provided here.
        for id in (n2, n3):
            ns.config_dataset(id, channel=15, panid=0xface, extpanid="dead00beef00cafe", network_name="test", networkkey="00112233445566778899aabbccddeeff")
            ns.ifconfig_up(id)
            ns.thread_start(id)

        self.go(300)
        self.assertFormPartitions(1)

    def testCommissioning(self):
        ns = self.ns

        n1 = ns.add("router")
        n2 = ns.add("router")

        ns.node_cmd(n1, "dataset init new")
        ns.node_cmd(n1, "dataset")
        ns.node_cmd(n1, "dataset commit active")
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


if __name__ == '__main__':
    unittest.main()
