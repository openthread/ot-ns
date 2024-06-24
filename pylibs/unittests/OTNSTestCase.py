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
import logging
import tracemalloc
import unittest

from otns.cli import OTNS


class OTNSTestCase(unittest.TestCase):

    @classmethod
    def setUpClass(cls) -> None:
        tracemalloc.start()
        logging.basicConfig(level=logging.DEBUG, format='%(asctime)-15s - %(levelname)s - %(message)s')

    def name(self) -> str:
        return self.id().replace('__main__.','')

    def setUp(self) -> None:
        logging.info("Setting up for test: %s", self.name())
        self.ns = OTNS(otns_args=['-log', 'debug']) # may add '-watch', 'trace' to see detailed OT node traces.

    def tearDown(self) -> None:
        self.ns.close()
        self.ns.save_pcap("tmp/unittest_pcap", self.name() + ".pcap" )

    def go(self, duration: float) -> None:
        """
        Run the simulation for a given duration.

        :param duration: the duration to simulate
        """
        self.ns.go(duration)

    def assertFormPartitions(self, count: int):
        pars = self.ns.partitions()
        self.assertEqual(count, len(pars), f"Partitions count mismatch: expected {count}, but is {len(pars)}")
        self.assertTrue(0 not in pars, pars)

    def assertFormPartitionsIgnoreOrphans(self, count: int):
        pars = self.ns.partitions()
        parsNoOrphans = []
        parsNoOrphans[:] = (value for value in pars if value != 0)
        self.assertEqual(count, len(parsNoOrphans), f"Partitions count mismatch: expected {count}, but is {len(parsNoOrphans)}")

    def assertNodeState(self, nodeid: int, state: str):
        cur_state = self.ns.get_state(nodeid)
        self.assertEqual(state, cur_state, f"Node {nodeid} state mismatch: expected {state}, but is {cur_state}")

    def assertPings(self, pings, n, max_delay=1000, max_fails=0):
        self.assertEqual(n, len(pings))
        n_fails = 0
        for srcid, dst, datasize, delay in pings:
            if delay == OTNS.MAX_PING_DELAY:
                n_fails += 1
            else:
                self.assertTrue(delay <= max_delay)
        self.assertTrue(n_fails <= max_fails)
