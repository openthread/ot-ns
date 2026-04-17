#!/usr/bin/env python3
#
# Copyright (c) 2026, The OTNS Authors.
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

import time
import unittest

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS


class RealtimeTests(OTNSTestCase):
    """
    Unit tests that need to run in the OTNS -realtime mode: for example RCP nodes.
    """

    def setUp(self) -> None:
        super().setUp()
        self.ns.close()
        self.ns = OTNS(otns_args=['-log', 'debug', '-realtime'])

    def testAddRcpNode(self):
        ns: OTNS = self.ns
        ns.loglevel = 'trace'  # trace logging to reveal realtime (RT) simulation control details.

        ns.add('router')
        ns.add('rcp')
        self.assertEqual(2, len(ns.nodes()))

        # let the simulation advance
        time.sleep(15)
        self.assertEqual(2, len(ns.nodes()))
        self.assertFormPartitions(1)

    def testAddDelRcpNodes(self):
        ns: OTNS = self.ns

        ns.add('rcp')
        ns.add('rcp')
        ns.add('rcp')
        ns.add('fed')
        ns.add('med')
        self.assertEqual(5, len(ns.nodes()))

        # let the simulation advance
        time.sleep(45)
        self.assertEqual(5, len(ns.nodes()))
        self.assertFormPartitions(1)

        ns.delete(1)
        self.assertEqual(4, len(ns.nodes()))
        time.sleep(10)
        self.assertEqual(4, len(ns.nodes()))
        self.assertTrue(1 not in ns.nodes())
        self.assertFormPartitions(1)

        ns.delete(3)
        self.assertEqual(3, len(ns.nodes()))
        time.sleep(10)
        self.assertEqual(3, len(ns.nodes()))
        self.assertTrue(1 not in ns.nodes())
        self.assertTrue(3 not in ns.nodes())
        self.assertFormPartitions(1)


if __name__ == '__main__':
    unittest.main()
