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
from otns.cli import errors, OTNS


class ExeVersionTests(OTNSTestCase):
    """
    Unit tests of the 'exe' command to select OT node executables; running nodes with manually specified
    executable; and the 'vxx' parameters for adding nodes of different (prebuilt) OT versions.
    """
    def testAddNodeWrongExecutable(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.go(5)
        self.assertEqual(1, len(ns.nodes()))
        with self.assertRaises(errors.OTNSCliError):
            ns.add('router', executable='ot-cli-nonexistent')
        ns.go(5)
        self.assertEqual(1, len(ns.nodes()))
        self.assertEqual(1, len(ns.partitions()))

    def testExe(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.go(10)
        ns._do_command("exe ftd \"./path/to/non-existent-executable\"")
        with self.assertRaises(errors.OTNSCliError):
            ns.add('router')
        ns._do_command("exe")
        ns._do_command("exe default")
        ns.go(10)
        ns.add('router')
        ns.go(30)
        ns.add('fed')
        self.assertEqual(3, len(ns.nodes()))
        ns.go(140)
        self.assertEqual(3, len(ns.nodes()))
        self.assertEqual(1, len(ns.partitions()))

        ns._do_command("exe v11")
        ns.add('router')
        ns.go(10)
        ns._do_command("exe v12")
        ns.add('router')
        ns.go(10)
        ns._do_command("exe v13")
        ns.add('router')
        ns.go(10)
        self.assertEqual(6, len(ns.nodes()))
        ns.go(60)
        self.assertEqual(6, len(ns.nodes()))
        self.assertEqual(1, len(ns.partitions()))

def testAddVersionNodes(self):
        ns: OTNS = self.ns
        ns.add('router', x=250, y=250)
        ns.go(10)
        ns.add('router', version='v13')
        ns.go(10)
        ns.add('router', version='v12')
        ns.go(10)
        ns.add('router', version='v11')
        ns.go(10)
        self.assertEqual(4, len(ns.nodes()))
        self.assertEqual(1, len(ns.partitions()))


if __name__ == '__main__':
    unittest.main()
