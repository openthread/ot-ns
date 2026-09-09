#!/usr/bin/env python3
# Copyright (c) 2020-2026, The OTNS Authors.
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

    # Default substrings of OTNS warn/error log messages that fail the test in tearDown.
    # Subclasses may rebind this at the class level. Inside an individual test,
    # mutate self.logFailureSignatures (a per-instance copy made in setUp) to add or
    # remove signatures without affecting other OTNS unit tests.
    logFailureSignatures = {
        'received unexpected event', 'RfSimEvent with unexpected param', 'received unexpected (longer) node output',
        'received from unknown Node', 'gRPC server could not listen', 'gRPC server quit with error',
        'visualization error: ', 'RecvEvents timeout: alive nodes are', 'processEvent() detects unknown node',
        'New node did not send a correct init event EventTypeNodeInfo',
        'sent incomplete event data, closing node connection', 'received unknown status push: ',
        ' received from apparent duplicate Node ', 'Node did not exit in time, sending SIGKILL',
        'Kill existing grpcwebproxy process failed'
    }

    @classmethod
    def setUpClass(cls) -> None:
        tracemalloc.start()
        logging.basicConfig(level=logging.DEBUG, format='%(asctime)-15s - %(levelname)s - %(message)s')

    def name(self) -> str:
        return self.id().replace('__main__.', '')

    def setUp(self) -> None:
        logging.info("Setting up for test: %s", self.name())
        self.ns = OTNS(otns_args=['-log', 'debug'])  # may add '-watch', 'trace' to see detailed OT node traces.
        # Per-instance mutable copy: tests may add/remove from self.logFailureSignatures
        # without bleeding into other tests via the shared class-level set.
        self.logFailureSignatures = set(type(self).logFailureSignatures)

    def tearDown(self) -> None:
        self.ns.close()
        self.ns.save_pcap("tmp/unittest_pcap", self.name() + ".pcap")

        # Fail test if any panic/fatal log entries were recorded
        log_entries = self.ns.get_log_entries(levels={'panic', 'fatal'})
        if log_entries:
            msg = "OTNS emitted %d panic/fatal log entries during the test:" % len(log_entries)
            for level, message in log_entries:
                msg += f"\n  [{level}] {message}"
            self.fail(msg)

        # Fail test for specific warnings/errors we never want to see
        log_entries = self.ns.get_log_entries(levels={'warn', 'error'})
        matched = [(lv, msg) for lv, msg in log_entries if any(sig in msg for sig in self.logFailureSignatures)]
        if matched:
            msg = "OTNS emitted %d warning/error log entries matching a failure signature:" % len(matched)
            for level, message in matched:
                msg += f"\n  [{level}] {message}"
            self.fail(msg)

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
        self.assertEqual(count, len(parsNoOrphans),
                         f"Partitions count mismatch: expected {count}, but is {len(parsNoOrphans)}")

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
