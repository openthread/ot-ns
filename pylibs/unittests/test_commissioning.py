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
        self.ns = OTNS(otns_args=['-raw', '-log', 'debug'])
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
        n1 = ns.add("router", x=0, y=0)
        n2 = ns.add("router", x=50, y=0)
        n3 = ns.add("router", x=0, y=50)
        for id in (n1, n2, n3):
            ns.set_network_name(id, "test")
            ns.set_panid(id, 0xface)
            ns.set_networkkey(id, "00112233445566778899aabbccddeeff")
            ns.ifconfig_up(id)
            ns.thread_start(id)

        self.go(30)
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
        print('countes', c)
        joins = ns.joins()
        print('joins', joins)
        self.assertFormPartitions(1)
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success


if __name__ == '__main__':
    unittest.main()
