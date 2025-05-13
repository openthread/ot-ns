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

import unittest

from OTNSTestCase import OTNSTestCase


class BorderRouterTests(OTNSTestCase):

    def testAddDeleteBorderRouter(self):
        ns = self.ns
        nid = ns.add('br')
        self.assertNodeState(nid, 'detached')
        ns.go(10)
        self.assertNodeState(nid, 'leader')
        ns.delete(1)
        ns.go(10)
        self.assertTrue(len(ns.nodes()) == 0)

    def testBorderRouterDistributesOmrPrefix(self):
        ns = self.ns
        ns.radiomodel = 'MIDisc'  # force line topology

        ns.add('br', x=100, y=100)
        ns.go(10)

        ns.add('router', x=250, y=100)
        ns.go(10)

        ns.add('fed', x=400, y=100)
        ns.go(80)

        ns.ping(3, 1, addrtype='slaac', count=4)
        ns.go(50)
        ns.ping(1, 3, addrtype='slaac', count=4)
        ns.go(50)

        pings = ns.pings()
        self.assertTrue(len(pings) == 8)


if __name__ == '__main__':
    unittest.main()
