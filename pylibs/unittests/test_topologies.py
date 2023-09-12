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

import math
import random
import unittest

from OTNSTestCase import OTNSTestCase

class TopologiesTests(OTNSTestCase):

    def testLargeNetwork(self):
        n = 1024 # network size (nodes)
        ns = self.ns
        ns.loglevel = 'info'
        ns.add("router")
        ns.go(10)
        self.assertEqual(ns.get_state(1), "leader")

        for i in range(n-1):
            ns.add("router")
            ns.go(0.001)

        ns.go(1)
        self.assertEqual(n,len(ns.nodes()))
        self.assertEqual(ns.get_state(1), "leader")

    def testDenseNetwork(self):
        nn = 144 # number of nodes
        x0 = 100 # start coordinate (x,y) (pixels)
        dx = 100 # coordinate delta x/y between nodes (pixels)
        rr = 700 # radio range (pixels)
        probability_med = 0.20 # probability that new node is a MED

        ns = self.ns
        ns.loglevel = 'info'
        x = x0
        y = x0
        for i in range(nn):
            if random.random() < probability_med:
                ns.add("med", x, y, radio_range=rr)
            else:
                ns.add("router", x, y, radio_range=rr)
            ns.go(1)
            x += dx
            if x >= (x0+math.sqrt(nn)*dx):
                x = x0
                y += dx

        ns.go(20)
        self.assertEqual(nn,len(ns.nodes()))

if __name__ == '__main__':
    unittest.main()
