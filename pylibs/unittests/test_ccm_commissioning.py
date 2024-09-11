#!/usr/bin/env python3
# Copyright (c) 2024, The OTNS Authors.
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
import subprocess
import time
import unittest

from OTNSTestCase import OTNSTestCase
from otns.cli import OTNS


class CcmTests(OTNSTestCase):
    """
    Thread Commercial Commissioning Mode (CCM) tests. All tests make use of an external server,
    OT-Registrar, which runs from a Java JAR file. The Registrar forwards the Voucher, obtained
    from the vendor-controlled MASA server, to the Joiner. The Registrar also generates the
    final domain identity (LDevID certificate) for the Joiner to use.
    See RFC 8995 and IETF ANIMA WG cBRSKI draft for details.
    """

    registrar_process = None

    def setUp(self) -> None:
        logging.info("Setting up for test: %s", self.name())
        self.ns = OTNS(otns_args=['-log', 'trace', '-pcap', 'wpan-tap', '-seed', '4'])
        self.ns.speed = 1e6

    def tearDown(self) -> None:
        self.stopRegistrar()
        super().tearDown()

    def setFirstNodeDataset(self, n1) -> None:
        self.ns.node_cmd(n1, "dataset init new")
        self.ns.node_cmd(n1, "dataset channel 22")
        self.ns.node_cmd(n1, "dataset meshlocalprefix fd00:777e::")
        self.ns.node_cmd(n1, "dataset networkkey 00112233445566778899aabbccddeeff")  # allow easy Wireshark dissecting
        self.ns.node_cmd(n1, "dataset securitypolicy 672 orcCR 3")  # enable CCM-commissioning flag in secpolicy
        self.ns.node_cmd(n1, "dataset commit active")

    def setAlternativeDataset(self, n1) -> None:
        self.ns.node_cmd(n1, "dataset init active")
        self.ns.node_cmd(n1, "dataset channel 22")
        self.ns.node_cmd(n1, "dataset meshlocalprefix fd00:777e::")
        self.ns.node_cmd(n1, "dataset securitypolicy 672 orcCR 3")  # enable CCM-commissioning flag in secpolicy
        self.ns.node_cmd(n1, "dataset commit active")

    def startRegistrar(self):
        self.registrar_log_file = open("tmp/ot-registrar.log", 'w')
        self.registrar_process = subprocess.Popen([
            'java', '-jar', './etc/ot-registrar/ot-registrar-0.3-jar-with-dependencies.jar', '-registrar', '-vv', '-f',
            './etc/ot-registrar/credentials_registrar.p12'
        ],
                                                  stdout=self.registrar_log_file,
                                                  stderr=subprocess.STDOUT)
        self.assertIsNone(self.registrar_process.returncode)
        time.sleep(1)  # FIXME could detect when Registrar is ready to serve, with process.communicate()

    def stopRegistrar(self):
        if self.registrar_process is None:
            return
        logging.debug("stopping OT Registrar")
        self.registrar_process.terminate()
        if self.registrar_log_file is not None:
            self.registrar_log_file.close()
        self.registrar_process = None
        self.registrar_log_file = None

    def testAddCcmNodesMixedWithRegular(self):
        ns = self.ns

        n1 = ns.add("br", version="ccm")
        n2 = ns.add("router", version="ccm")
        n2 = ns.add("router", version="ccm")
        n2 = ns.add("router")

        ns.go(30)
        self.assertFormPartitions(1)

    def testCommissioningOneCcmNode(self):
        ns = self.ns
        self.startRegistrar()
        #ns.web()
        ns.coaps_enable()
        ns.radiomodel = 'MIDisc'  # enforce strict line topologies for testing

        n1 = ns.add("br", x=100, y=100, radio_range=120, version="ccm", script="")
        n2 = ns.add("router", x=100, y=200, radio_range=120, version="ccm", script="")

        # configure sim-host server that acts as BRSKI Registrar
        # TODO update IPv6 addr
        ns.cmd('host add "masa.example.com" "910b::1234" 5684 5684')

        # n1 is a BR out-of-band configured with initial dataset, and becomes leader+ccm-commissioner
        self.setFirstNodeDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        ns.go(15)
        state_n1 = ns.get_state(n1)
        self.assertTrue(state_n1 == "leader")
        ns.commissioner_start(n1)
        ns.go(5)
        ns.coaps()  # see emitted CoAP events

        # n2 joins as CCM joiner
        # because CoAP server is real, let simulation also move in near real time speed.
        ns.speed = 50
        ns.commissioner_ccm_joiner_add(n1, "*")
        ns.ifconfig_up(n2)
        ns.ccm_joiner_start(n2)
        ns.go(20)

        # check join result
        ns.coaps()  # see emitted CoAP events
        ns.cmd('host list')
        state_n2 = ns.get_state(n2)
        self.assertTrue(state_n2 == "router" or state_n2 == "child")
        ns.go(20)

        #ns.interactive_cli()

    def testCommissioningOneCcmNodeOneJoinerRouter(self):
        ns = self.ns
        self.startRegistrar()
        #ns.web()
        ns.watch_default('debug')
        ns.coaps_enable()
        ns.radiomodel = 'MIDisc'  # enforce strict line topologies for testing

        n1 = ns.add("br", x=100, y=100, radio_range=120, version="ccm")
        n2 = ns.add("router", x=100, y=200, radio_range=120, version="ccm")
        n3 = ns.add("router", x=100, y=300, radio_range=120, version="ccm", script="")

        # configure sim-host server that acts as BRSKI Registrar
        # TODO update IPv6 addr
        ns.cmd('host add "masa.example.com" "910b::1234" 5684 5684')

        # n1 is a BR out-of-band configured with initial dataset, and becomes leader+ccm-commissioner
        self.setAlternativeDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        ns.go(15)
        state_n1 = ns.get_state(n1)
        self.assertTrue(state_n1 == "leader")
        # n1 starts commissioner
        ns.commissioner_start(n1)
        ns.go(5)
        ns.coaps()  # see emitted CoAP events

        # n2 also added out-of-band, for Joiner Router role
        self.setAlternativeDataset(n2)
        ns.ifconfig_up(n2)
        ns.thread_start(n2)
        ns.go(20)
        state_n2 = ns.get_state(n2)
        self.assertTrue(state_n2 == "router" or state_n2 == "child")
        ns.coaps()  # see emitted CoAP events

        # n3 joins as CCM joiner - needs to search channel.
        # because CoAP server is real, let simulation also move in near real time speed.
        ns.speed = 1
        ns.commissioner_ccm_joiner_add(n1, "*")
        ns.ifconfig_up(n3)
        ns.node_cmd(n3, 'coaps x509')
        ns.joiner_startccm(n3)
        ns.go(10)
        ns.coaps()  # see emitted CoAP events
        ns.cmd('host list')

        # n3 automatically has enabled Thread and joined the network
        #ns.interactive_cli()
        state_n3 = ns.get_state(n3)
        self.assertTrue(state_n3 == "router" or state_n3 == "child")

    def testCommissioningOneHop(self):
        ns = self.ns
        ns.web()
        ns.coaps_enable()
        ns.radiomodel = 'MIDisc'  # enforce strict line topologies for testing

        n1 = ns.add("br", x=100, y=100, radio_range=120, script="", version="ccm")
        n2 = ns.add("router", x=100, y=200, radio_range=120, script="")
        n3 = ns.add("router", x=200, y=100, radio_range=120, script="", version="ccm")

        # configure sim-host server that acts as BRSKI Registrar
        # TODO update IPv6 addr
        ns.cmd('host add "masa.example.com" "910b::1234" 5683 5683')

        # n1 is a BR out-of-band configured with initial dataset, and becomes leader+ccm-commissioner
        self.setFirstNodeDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        self.go(15)
        self.assertTrue(ns.get_state(n1) == "leader")
        ns.commissioner_start(n1)

        # n2 joins as regular joiner
        ns.commissioner_joiner_add(n1, "*", "TEST123")
        ns.ifconfig_up(n2)
        ns.joiner_start(n2, "TEST123")
        self.go(20)
        ns.thread_start(n2)
        self.go(100)
        c = ns.counters()
        print('counters', c)
        joins = ns.joins()
        print('joins', joins)
        self.assertFormPartitionsIgnoreOrphans(1)  # ignore orphan n3
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success

        # n3 joins as CCM joiner
        # because CoAP server is real, let simulation also move in near real time speed.
        ns.speed = 5
        ns.commissioner_ccm_joiner_add(n1, "*")
        ns.ifconfig_up(n3)
        ns.joiner_startccm(n3)
        self.go(20)
        ns.thread_start(n3)
        self.go(100)

        ns.node_cmd(n3, 'dataset active')

        c = ns.counters()
        print('counters', c)
        joins = ns.joins()
        print('joins', joins)
        # ns.interactive_cli()
        self.assertFormPartitions(1)
        self.assertTrue(joins and joins[0][1] > 0)  # assert join success


if __name__ == '__main__':
    unittest.main()
