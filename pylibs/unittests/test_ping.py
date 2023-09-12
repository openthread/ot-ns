#!/usr/bin/env python3
# Copyright (c) 2020-2023, The OTNS Authors.
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

import logging
import tracemalloc
import unittest

from OTNSTestCase import OTNSTestCase

tracemalloc.start()


class PingTests(OTNSTestCase):

    def testPing(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")

        # The pinging starts already while the network is still being formed. Hence, the first few pings will
        # fail with 'Error 5: Busy' from the OT node.
        for i in range(100):
            ns.ping(1, 2, datasize=10)
            ns.ping(2, 1, datasize=10)
            ns.go(1)

        pings = ns.pings()
        logging.debug('len(pings)=%d, expected ~180' % (len(pings)))
        self.assertTrue(len(pings) < 200)
        self.assertTrue(len(pings) > 150)

        dst_addrs = (str(ns.get_mleid(1)), str(ns.get_mleid(2)))

        for srcid, dst, datasize, delay in pings:
            assert srcid in (1, 2)
            assert dst in dst_addrs
            assert datasize == 10
            assert delay > 0
            assert delay <= 10000

        self.assertFalse(ns.pings())
        ns.go(11)
        self.assertFalse(ns.pings())

    def testPingLineTopology(self):
        ns = self.ns

        for i in range(10):
            ns.add("router", i*80, 200)
        for i in range(100):
            ns.go(20)
            pts = ns.partitions()
            if len(pts) == 1 and 0 not in pts:
                break

        # two-fragment ping packet adds extra closely spaced traffic. This impacts performance greatly in
        # hidden-node situations.
        for pingDataSize in [64,128]:
            #ns.go(50)
            pingDelays = []

            for i in range(80):
                ns.ping(1, 10, datasize=pingDataSize)
                ns.go(0.100)
                ns.ping(10, 1, datasize=pingDataSize) # reverse-direction ping may collide with earlier ping
                ns.go(10.900)

            pings = ns.pings()
            self.assertTrue(pings)
            for srcid, dst, datasize, delay in pings:
                self.assertTrue(srcid in (1, 10))
                self.assertTrue(datasize == pingDataSize)
                pingDelays.append(delay)

            self.assertFalse(ns.pings())

            pingSuccess = 1.0 - (pingDelays.count(10000) / len(pingDelays))
            pingDelays = list(filter(lambda a: a < 10000, pingDelays))
            pingAvg = sum(pingDelays) / len(pingDelays)

            logging.info(f"Ping success rate   : {pingSuccess}")
            logging.info(f"Average ping latency: {pingAvg}")

            if pingDataSize == 64:
                if ns.radiomodel == 'MIDisc':
                    # For the MIDisc radio model, the hidden-node problem pops up strongly. Lower expectations.
                    self.assertTrue(pingAvg < 600 and pingSuccess > 0.2)
                elif ns.radiomodel == 'MutualInterference':
                    # Some degree of hidden-node problem.
                    self.assertTrue(pingAvg < 600 and pingSuccess > 0.5)
                else:
                    self.assertTrue(pingAvg < 500 and pingSuccess > 0.75)
            else:
                self.assertTrue(pingAvg < 600 and pingSuccess > 0.1)

if __name__ == '__main__':
    unittest.main()
