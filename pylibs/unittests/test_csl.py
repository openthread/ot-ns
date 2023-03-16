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
    def testOneParentMultiCslChildren(self):
        ns = self.ns

        # setup a Parent Router with N SSED Children with different CSL Periods.
        ns.add("router", 100, 100)
        aCslPeriods = [3100, 500, 7225, 1024, 3125]
        for n in range(0,5):
            nodeid = ns.add("sed", 80 + n*20, 150)
            ns.node_cmd(nodeid,"csl period " + str(aCslPeriods[n]))
            ns.go(1)
        ns.go(10)

        # do some pings
        ns.ping(1,2)
        ns.go(1)
        ns.ping(2,1)
        ns.go(1)
        ns.ping(1,3)
        ns.go(2)
        ns.ping(3,1)
        ns.go(2)
        ns.ping(1,4)
        ns.go(1)
        ns.ping(4,1)
        ns.go(1)
        ns.ping(1,5)
        ns.go(1)
        ns.ping(5,1)
        ns.go(1)
        ns.ping(5,2)
        ns.go(2)
        ns.ping(2,5)
        ns.go(2)

        # long wait and some pings
        ns.go(300)
        ns.ping(3,4)
        ns.go(50)
        ns.ping(4,5)
        ns.go(10)

        # test ping results
        pings = ns.pings()
        self.assertTrue(pings)
        for srcid, dst, datasize, delay in pings:
            assert delay < 3000

if __name__ == '__main__':
    unittest.main()
