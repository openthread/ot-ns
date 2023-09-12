#!/usr/bin/env python3
# Copyright (c) 2022-2023 The OTNS Authors.
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
from test_basic import BasicTests
from test_commissioning import CommissioningTests
from test_ping import PingTests
from test_csl import CslTests
from otns.cli import errors, OTNS


class RadioModelTests(OTNSTestCase):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'MutualInterference'

    def testRadioModelSwitching(self):
        ns = self.ns
        ns.radiomodel = 'Ideal'
        radio_range = 100

        ns.add("router",0, 0, radio_range=radio_range)
        ns.add("router",0, radio_range+1, radio_range=radio_range)
        ns.add("router",radio_range+1, radio_range+1, radio_range=radio_range)
        # Reason to raise TxPower is that near the range limit, in MI radio model, there may be a
        # valid link but there also may be not (Shadow Fading effects). For the Ideal model, the
        # Tx power does not influence the range.
        ns.node_cmd(1,'txpower 20')
        ns.node_cmd(2,'txpower 20')
        ns.node_cmd(3,'txpower 20')
        ns.go(20)
        self.assertFormPartitions(3)

        ns.radiomodel = 'MutualInterference'
        self.assertEqual('MutualInterference', ns.radiomodel)
        ns.go(200)
        self.assertFormPartitions(1)

        ns.radiomodel = 'Ideal_Rssi'
        ns.go(180)
        self.assertFormPartitions(3)
        self.assertEqual('Ideal_Rssi', ns.radiomodel)

        ns.radiomodel = 'MIDisc'
        self.assertEqual('MIDisc', ns.radiomodel)
        ns.go(200)
        self.assertFormPartitions(3)

        with self.assertRaises(errors.OTNSCliError):
            ns.radiomodel = 'NotExistingName'
        self.assertEqual('MIDisc', ns.radiomodel)

        ns.node_cmd(1,'txpower -60')
        ns.node_cmd(2,'txpower -60')
        ns.node_cmd(3,'txpower -60')
        ns.radiomodel = 'MutualInterference'
        self.assertEqual('MutualInterference', ns.radiomodel)
        ns.go(200)
        self.assertFormPartitions(3)

        ns.node_cmd(1,'txpower 20')
        ns.node_cmd(2,'txpower 20')
        ns.node_cmd(3,'txpower 20')
        ns.go(200)
        self.assertFormPartitions(1)


class BasicTests_Ideal(BasicTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal'


class BasicTests_IdealRssi(BasicTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal_Rssi'


class BasicTests_MIDisc(BasicTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'MIDisc'


class CommissioningTests_Ideal(CommissioningTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal'


class CommissioningTests_IdealRssi(CommissioningTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal_Rssi'


class CommissioningTests_MIDisc(CommissioningTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'MIDisc'


class PingTests_IdealRssi(PingTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal_Rssi'


class PingTests_MIDisc(PingTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'MIDisc'


class CslTests_IdealRssi(CslTests):

    # override
    def setUp(self):
        super().setUp()
        self.ns.radiomodel = 'Ideal_Rssi'


if __name__ == '__main__':
    loader = unittest.defaultTestLoader
    suite = loader.loadTestsFromTestCase(RadioModelTests)
    suite.addTest(loader.loadTestsFromTestCase(BasicTests_Ideal))
    suite.addTest(loader.loadTestsFromTestCase(BasicTests_IdealRssi))
    suite.addTest(loader.loadTestsFromTestCase(BasicTests_MIDisc))
    suite.addTest(loader.loadTestsFromTestCase(CommissioningTests_Ideal))
    suite.addTest(loader.loadTestsFromTestCase(CommissioningTests_IdealRssi))
    suite.addTest(loader.loadTestsFromTestCase(CommissioningTests_MIDisc))
    suite.addTest(loader.loadTestsFromTestCase(PingTests_IdealRssi))
    suite.addTest(loader.loadTestsFromTestCase(PingTests_MIDisc))
    suite.addTest(loader.loadTestsFromTestCase(CslTests_IdealRssi))
    unittest.TextTestRunner().run(suite)
