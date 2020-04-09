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

import unittest

from OTNSTestCase import OTNSTestCase
from otns import consts


class BasicRawTests(OTNSTestCase):
    def __init__(self, *args, **kwargs):
        super(BasicRawTests, self).__init__(*args, **kwargs, otns_args=['-raw'])

    def testInitialPanid(self):
        r1 = self.ns.add("router")
        self.assertEqual(self.ns.get_panid(r1), 0xffff)

    def testSamePanid(self):
        ns = self.ns
        r1 = ns.add("router")
        r2 = ns.add("router")

        ns.set_panid(r1, 0xface)
        self.assertEqual(0xface, self.ns.get_panid(r1))
        ns.set_panid(r2, 0xface)
        self.assertEqual(0xface, self.ns.get_panid(r2))

        for nid in [r1, r2]:
            self.ns.ifconfig_up(nid)
            self.ns.thread_start(nid)

        ns.go(10)
        self.assertFormPartitons(1)

    def testDifferentPanid(self):
        ns = self.ns
        r1 = ns.add("router")
        r2 = ns.add("router")

        ns.set_panid(r1, 0xface)
        self.assertEqual(0xface, self.ns.get_panid(r1))
        ns.set_panid(r2, 0xabcd)
        self.assertEqual(0xabcd, self.ns.get_panid(r2))

        for nid in [r1, r2]:
            self.ns.ifconfig_up(nid)
            self.ns.thread_start(nid)

        ns.go(10)
        self.assertFormPartitons(2)

    def testDefaultNetworkName(self):
        r1 = self.ns.add("router")
        self.assertEqual(consts.DEFAULT_NETWORK_NAME, self.ns.get_network_name(r1))

    def testDefaultChannel(self):
        r1 = self.ns.add("router")
        self.assertEqual(consts.DEFAULT_CHANNEL, self.ns.get_channel(r1))


if __name__ == '__main__':
    unittest.main()
