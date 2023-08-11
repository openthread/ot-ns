#!/usr/bin/env python3
# Copyright (c) 2023, The OTNS Authors.
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
import unittest
from typing import Dict

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS


class CslTests(OTNSTestCase):
    #override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'MutualInterference'

    def verifyPings(self, pings, n, maxDelay=1000, maxFails=0):
        self.assertEqual(n, len(pings))
        nFails = 0
        for srcid, dst, datasize, delay in pings:
            if delay == OTNS.MAX_PING_DELAY:
                nFails += 1
            else:
                self.assertTrue(delay <= maxDelay)
        self.assertTrue(nFails <= maxFails)

    def testSsedConnectsToParent(self):
        ns = self.ns

        # add SSED
        nodeid = ns.add("sed", 220, 100)
        ns.node_cmd(nodeid,"csl period 288000")
        ns.go(10)

        # Parent comes in, SSED connects
        ns.add("router", 100, 100)
        ns.go(10)
        self.assertFormPartitions(1)

        # SSED can ping parent
        ns.ping(1,2)
        ns.go(1)
        ns.ping(1,2)
        ns.go(1)
        self.verifyPings(ns.pings(), 2, maxDelay=2000, maxFails=1)

    def testOneParentMultiCslChildren(self):
        ns = self.ns

        # setup a Parent Router with N SSED Children with different CSL Periods.
        N = 8
        ns.add("router", 100, 100)
        # below CSL periods to test (given in units of 160 us)
        aCslPeriods = [3100, 500, 7225, 1024, 3125, 3124, 250, 5999, 777, 1024]
        for n in range(0,N):
            nodeid = ns.add("sed", 80 + n*20, 150)
            ns.node_cmd(nodeid,"csl period " + str(aCslPeriods[n] * 160))
            ns.go(1)
        ns.go(45)
        self.assertFormPartitions(1)

        for k in range(0,5):
            # do some pings
            for n in range(0,N):
                ns.ping(1,2+n)
                ns.go(2)
                ns.ping(2+n,1)
                ns.go(2)

            # long wait and some pings
            ns.go(300)
            for n in range(0,N):
                ns.ping(1,2+n)
                ns.go(20)
                ns.ping(2+n,1)
                ns.go(20)

            # test ping results
            self.verifyPings(ns.pings(), N*4, maxDelay=3000, maxFails=1)

    def testCslReenable(self):
        ns = self.ns

        # setup a Parent Router with SSED Child
        ns.add("router", 100, 100)
        ns.go(10)
        nodeid = ns.add("sed", 200, 100)
        ns.node_cmd(nodeid,"csl period 288000")
        ns.go(10)
        self.assertFormPartitions(1)

        # SSED pings parent
        for n in range(0,15):
            ns.ping(2,1,datasize=n+10)
            ns.go(5)
        self.verifyPings(ns.pings(), 15, maxDelay=3000, maxFails=1)

        # parent pings SSED
        for n in range(0,15):
            ns.ping(1,2,datasize=n+10)
            ns.go(5)
        self.verifyPings(ns.pings(), 15, maxDelay=3000, maxFails=1)

        for k in range(0,4):
            # disable CSL
            ns.node_cmd(nodeid,"csl period 0")
            ns.go(1)

            # SSED pings parent
            for n in range(0,15):
                ns.ping(2,1,datasize=n+10)
                ns.go(5)
            self.verifyPings(ns.pings(), 15, maxDelay=3000, maxFails=1)

            # re-enable CSL
            ns.node_cmd(nodeid,"csl period 144000")
            ns.go(1)

            # SSED pings parent
            for n in range(0,15):
                ns.ping(2,1,datasize=n+10)
                ns.go(5)
            self.verifyPings(ns.pings(), 15, maxDelay=3000, maxFails=1)

            # parent pings SSED
            for n in range(0,15):
                ns.ping(1,2,datasize=n+10)
                ns.go(5)
            self.verifyPings(ns.pings(), 15, maxDelay=3000, maxFails=1)

if __name__ == '__main__':
    unittest.main()
