#!/usr/bin/env python3
#
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
#
import os
import logging
import os
import random
import signal
import threading
import time
import unittest
from subprocess import TimeoutExpired

from OTNSTestCase import OTNSTestCase
from otns.cli.errors import OTNSExitedError

WAIT_OTNS_TIMEOUT = 3


class SignalsTest(OTNSTestCase):
    """
    This test verifies OTNS handles OS signals gracefully.
    """

    def testSIGINT(self):
        self._test_signal_exit(signal.SIGINT)

    def testSIGTERM(self):
        self._test_signal_exit(signal.SIGTERM)

    def testSIGTERMx200(self):
        N = 200
        for i in range(N):
            logging.info("round %d", i + 1)
            self._test_signal_exit(signal.SIGTERM, 0.1 * random.random())

            self.tearDown()
            self.setUp()

    def testSIGQUIT(self):
        self._test_signal_exit(signal.SIGQUIT)

    def testSIGHUP(self):
        self._test_signal_exit(signal.SIGHUP)

    def testSIGKILL(self):
        self._test_signal_exit(signal.SIGKILL)

    def testSIGALRM(self):
        self._test_signal_ignore(signal.SIGALRM)

    def testCommandHandleSignalOkx100(self):
        for i in range(100):
            self._testCommandHandleSignalOk()

            self.tearDown()
            self.setUp()

    def _testCommandHandleSignalOk(self):
        t = threading.Thread(target=self._send_signal, args=(0.02, signal.SIGTERM))
        t.start()
        try:
            self.ns.speed = float('inf')
            while True:
                self.ns.add("router")
        except OTNSExitedError as ex:
            self.assertEqual(0, ex.exit_code)

        t.join()

    def _test_signal_ignore(self, sig: int):
        t = threading.Thread(target=self._send_signal, args=(1, sig))
        t.start()
        t2 = threading.Thread(target=self._run_simulation)
        t2.start()

        t.join()
        self.assertRaises(TimeoutExpired, self.ns._otns.wait, timeout=0.1)

        self.ns._otns.send_signal(signal.SIGTERM)
        t2.join()
        exit_code = self.ns._otns.wait()
        self.assertEqual(0, exit_code, "exit code should be 0")

    def _test_signal_exit(self, sig: int, duration: float = 1):
        self._setup_simulation()

        t = threading.Thread(target=self._send_signal, args=(duration, sig))
        t.start()
        self._run_simulation()
        t.join()
        try:
            exit_code = self.ns._otns.wait(timeout=WAIT_OTNS_TIMEOUT)
        except TimeoutExpired:
            logging.error('OTNS exit-signal handling took too long. Debug info follows below.')
            logging.error('OTNS error code: %s', self.ns._otns.returncode)
            os.system(f"curl http://localhost:8997/debug/pprof/goroutine?debug=2")
            raise

        if sig == signal.SIGKILL:
            self.assertNotEqual(0, exit_code, "exit code should not be 0")
            # Kill all ot-cli-ftd child processes.
            os.system('killall -KILL ot-cli-ftd')
        else:
            self.assertEqual(0, exit_code, "exit code should be 0")

    def _setup_simulation(self):
        self.ns.speed = float('inf')
        self.ns.add("router")
        self.ns.add("router")
        self.ns.add("router")

    def _run_simulation(self):
        try:
            self.ns.go()
        except OTNSExitedError:
            return
        except BrokenPipeError:
            return

    def _send_signal(self, delay: float, sig: int):
        logging.debug(f"sleep {delay} ...")
        time.sleep(delay)
        logging.debug(f'sending signal {sig}')
        self.ns._otns.send_signal(sig)


if __name__ == '__main__':
    unittest.main()
