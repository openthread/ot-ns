#!/usr/bin/env python3
#
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
#
import signal
import threading
import time
import unittest
from subprocess import TimeoutExpired

from otns.cli.errors import OTNSCliEOFError
from OTNSTestCase import OTNSTestCase


class SignalsTest(OTNSTestCase):
    """
    This test verifies OTNS handles OS signals gracefully.
    """

    def testSIGINT(self):
        self._test_signal_exit(signal.SIGINT)

    def testSIGTERM(self):
        self._test_signal_exit(signal.SIGTERM)

    def testSIGQUIT(self):
        self._test_signal_exit(signal.SIGQUIT)

    def testSIGHUP(self):
        self._test_signal_exit(signal.SIGHUP)

    def testSIGKILL(self):
        self._test_signal_exit(signal.SIGKILL)

    def testSIGALRM(self):
        self._test_signal_ignore(signal.SIGALRM)

    def testSIGCHLD(self):
        self._test_signal_ignore(signal.SIGCHLD)

    def _test_signal_ignore(self, sig: int):
        t = threading.Thread(target=self._send_signal, args=(1, sig))
        t.start()
        t2 = threading.Thread(target=self._run_simulation)
        t2.start()

        t.join()
        self.assertRaises(TimeoutExpired, self.ns._otns.wait, timeout=0.1)

        self.ns._otns.send_signal(signal.SIGKILL)
        t2.join()
        exit_code = self.ns._otns.wait()
        self.assertNotEqual(0, exit_code, "exit code should not be 0")

    def _test_signal_exit(self, sig: int):
        t = threading.Thread(target=self._send_signal, args=(1, sig))
        t.start()
        self._run_simulation()
        t.join()
        exit_code = self.ns._otns.wait()
        if sig == signal.SIGKILL:
            self.assertNotEqual(0, exit_code, "exit code should not be 0")
        else:
            self.assertEqual(0, exit_code, "exit code should be 0")

    def _run_simulation(self):
        self.ns.speed = float('inf')
        self.ns.add("router")
        self.ns.add("router")
        self.ns.add("router")
        try:
            self.ns.go()
        except OTNSCliEOFError:
            return

    def _send_signal(self, delay: float, sig: int):
        print(f"sleep {delay} ...")
        time.sleep(delay)
        print(f'sending signal {sig}')
        self.ns._otns.send_signal(sig)


if __name__ == '__main__':
    unittest.main()
