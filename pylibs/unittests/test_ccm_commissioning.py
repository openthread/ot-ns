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

    # for OT-Registrar processes, kept for entire class.
    registrar_process = None
    masa_process = None
    registrar_log_file = None

    thread_domain_name = "TestCcmDomain"

    @classmethod
    def setUpClass(cls):
        OTNSTestCase.setUpClass()
        cls.startRegistrar()

    @classmethod
    def tearDownClass(cls):
        cls.stopRegistrar()
        OTNSTestCase.tearDownClass()

    def setUp(self):
        logging.info("Setting up for test: %s", self.name())
        self.ns = OTNS(otns_args=['-log', 'trace', '-pcap', 'wpan-tap', '-seed', '4'])
        self.ns.watch_default('debug')
        self.ns.speed = 1e6
        # configure sim-host server that acts as BRSKI Registrar. TODO update IPv6 addr
        self.ns.cmd('host add "masa.example.com" "910b::1234" 5684 5684')
        self.ns.coaps_enable()
        self.ns.radiomodel = 'MIDisc'  # enforce strict line topologies for testing

    def tearDown(self):
        logging.info("End of test: %s", self.name())
        self.ns.close()
        super().tearDown()

    def setActiveDataset(self, n1) -> None:
        self.ns.node_cmd(n1, "dataset init new")
        self.ns.node_cmd(n1, "dataset networkname CcmTestNet")
        self.ns.node_cmd(n1, "dataset channel 22")
        self.ns.node_cmd(n1, "dataset activetimestamp 456789")
        self.ns.node_cmd(n1, "dataset panid 0x1234")
        self.ns.node_cmd(n1, "dataset extpanid 39758ec8144b07fb")
        self.ns.node_cmd(n1, "dataset pskc 3ca67c969efb0d0c74a4d8ee923b576c")
        self.ns.node_cmd(n1, "dataset meshlocalprefix fd00:777e:10ca::")
        self.ns.node_cmd(n1, "dataset networkkey 00112233445566778899aabbccddeeff")  # allow easy Wireshark dissecting
        self.ns.node_cmd(n1, "dataset securitypolicy 672 orcCR 3")  # enable CCM-commissioning flag in secpolicy
        self.ns.node_cmd(n1, "dataset commit active")

    @classmethod
    def startRegistrar(cls):
        if cls.registrar_process is not None:
            return
        logging.debug("starting OT Registrar")
        cls.registrar_log_file = open("tmp/ot-registrar.log", 'w')
        cls.registrar_process = subprocess.Popen([
            'java', '-jar', './etc/ot-registrar/ot-registrar.jar', '-registrar', '-vv', '-f',
            './etc/ot-registrar/credentials_registrar.p12', '-m', 'localhost:9443', '-d', CcmTests.thread_domain_name
        ],
                                                 stdout=cls.registrar_log_file,
                                                 stderr=subprocess.STDOUT)
        cls.masa_log_file = open("tmp/ot-masa.log", 'w')
        cls.masa_process = subprocess.Popen([
            'java', '-jar', './etc/ot-registrar/ot-registrar.jar', '-masa', '-vv', '-f',
            './etc/ot-registrar/credentials_masa.p12'
        ],
                                            stdout=cls.masa_log_file,
                                            stderr=subprocess.STDOUT)
        cls.verifyRegistrarStarted()

    @classmethod
    def verifyRegistrarStarted(cls) -> None:
        for n in range(1, 20):
            time.sleep(0.5)
            with open("tmp/ot-registrar.log", 'r') as file:
                if "Registrar listening (CoAPS)" in file.read():
                    with open("tmp/ot-masa.log", 'r') as file2:
                        if "MASA server listening (HTTPS)" in file2.read():
                            return
        cls.stopRegistrar()
        raise Exception("OT-Registrar or OT-Masa not started correctly")

    @classmethod
    def stopRegistrar(cls):
        if cls.registrar_process is not None:
            logging.debug("stopping OT Registrar")
            cls.registrar_process.terminate()
            cls.registrar_process.wait()
            cls.registrar_process = None
        if cls.registrar_log_file is not None:
            cls.registrar_log_file.close()
            cls.registrar_log_file = None
        if cls.masa_process is not None:
            logging.debug("stopping MASA server")
            cls.masa_process.terminate()
            cls.masa_process.wait()
            cls.masa_process = None
        if cls.masa_log_file is not None:
            cls.masa_log_file.close()
            cls.masa_log_file = None

    def enrollBr(self, nid):
        self.ns.coaps()  # clear coaps

        # BR enrolls via AIL.
        self.ns.speed = 2
        self.ns.node_cmd(nid, "ipaddr add fd12::5")  # dummy address to allow sending to AIL. TODO resolve in stack.
        self.ns.joiner_startccmbr(nid)
        self.ns.go(3)

        coap_events = self.ns.coaps()  # see emitted CoAP events
        self.assertEqual(4, len(coap_events))  # messages are /rv, /vs, /sen, /es

    def testCcmNodesMixedWithRegular(self):
        ns = self.ns

        # Start a regular non-CCM network simulation, with CCM/non-CCM nodes mixed.
        ns.add("br", version="ccm")
        ns.add("router", version="ccm")
        ns.add("router", version="ccm")
        ns.add("med", version="ccm")
        # FIXME SSED CCM node support needs to be enabled still.
        #ns.add("ssed", version="ccm")
        ns.add("br")
        ns.add("router")
        ns.add("med")
        ns.add("ssed")

        ns.go(30)
        self.assertFormPartitions(1)

    def testOneCcmBorderRouter(self):
        ns = self.ns

        # n1 uses cBRSKI via its infrastructure network interface to get LDevID.
        # because CoAP server is real, let simulation also move in near real time speed.
        n1 = ns.add("br", version="ccm", script="")

        self.enrollBr(n1)

        ns.speed = 1e6
        self.setActiveDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        ns.go(10)

        ns.get_ipaddrs(1)
        ns.cmd('host list')  # list external hosts, just to check Registrar was in
        self.assertEqual(CcmTests.thread_domain_name, ns.get_domain_name(n1))
        state_n1 = ns.get_state(n1)
        self.assertTrue(state_n1 == "leader")
        self.assertFormPartitions(1)

    def testOneCcmNodeViaBr(self):
        ns = self.ns

        n1 = ns.add("br", x=100, y=100, radio_range=120, version="ccm", script="")
        n2 = ns.add("router", x=100, y=200, radio_range=120, version="ccm", script="")

        self.enrollBr(n1)

        # n1 is a BR out-of-band configured with initial dataset, and becomes leader+ccm-commissioner
        ns.speed = 1e6
        self.setActiveDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        ns.go(18)
        state_n1 = ns.get_state(n1)
        self.assertTrue(state_n1 == "leader")
        ns.commissioner_start(n1)
        ns.go(5)
        ns.coaps()  # see emitted CoAP events

        # n2 joins as CCM joiner
        # because CoAP server is real, let simulation also move in near real time speed.
        ns.speed = 2
        ns.commissioner_ccm_joiner_add(n1, "*")
        ns.ifconfig_up(n2)
        ns.joiner_startae(n2)
        ns.go(10)

        coap_events = ns.coaps()  # see emitted CoAP events
        self.assertEqual(4, sum(ev["src"] == n2 for ev in coap_events))

        # n2 performs NKP
        ns.joiner_startnkp(n2)
        ns.go(10)
        coap_events = ns.coaps()  # see emitted CoAP events

        # check that Joiner Finalize messages was sent.
        self.assertEqual(1, sum(ev["src"] == n2 and "uri" in ev and ev["uri"] == "c/jf" for ev in coap_events))

        # using the Active Dataset now obtained, join the network.
        ns.thread_start(n2)
        ns.speed = 1e6
        ns.go(10)

        # check join result
        ns.cmd('host list')
        self.assertEqual(CcmTests.thread_domain_name, ns.get_domain_name(n2))
        state_n2 = ns.get_state(n2)
        self.assertTrue(state_n2 == "router" or state_n2 == "child")
        self.assertFormPartitions(1)
        ns.go(20)

    def testOneCcmNodeOneJoinerRouter(self):
        ns = self.ns

        n1 = ns.add("br", x=100, y=100, radio_range=120, version="ccm", script="")
        n2 = ns.add("router", x=100, y=200, radio_range=120, version="ccm", script="")
        n3 = ns.add("router", x=100, y=300, radio_range=120, version="ccm", script="")

        self.enrollBr(n1)

        # n1 is a BR out-of-band configured with initial dataset, and becomes leader+ccm-commissioner
        ns.speed = 1e6
        self.setActiveDataset(n1)
        ns.ifconfig_up(n1)
        ns.thread_start(n1)
        ns.go(18)
        state_n1 = ns.get_state(n1)
        self.assertTrue(state_n1 == "leader")
        # n1 starts commissioner
        ns.commissioner_start(n1)
        ns.go(5)
        ns.coaps()  # see emitted CoAP events

        # n2 also added out-of-band, for Joiner Router role
        self.setActiveDataset(n2)
        ns.ifconfig_up(n2)
        ns.thread_start(n2)
        ns.go(20)
        state_n2 = ns.get_state(n2)
        self.assertTrue(state_n2 == "router" or state_n2 == "child")
        self.assertFormPartitionsIgnoreOrphans(1)
        ns.coaps()  # see emitted CoAP events

        # n3 performs AE/cBRSKI - needs to search channel.
        # because CoAP server is real, let simulation also move in near real time speed.
        ns.speed = 2
        ns.commissioner_ccm_joiner_add(n1, "*")
        ns.ifconfig_up(n3)
        ns.joiner_startae(n3)
        ns.go(10)

        # n3 performs NKP
        ns.joiner_startnkp(n3)
        ns.go(10)

        # n3 starts Thread
        ns.speed = 1e6
        ns.thread_start(n3)
        ns.go(30)
        ns.coaps()  # see emitted CoAP events
        ns.cmd('host list')  # list external hosts, just to check Registrar was in

        # test n3 joined the network
        self.assertEqual(CcmTests.thread_domain_name, ns.get_domain_name(n3))
        state_n3 = ns.get_state(n3)
        self.assertTrue(state_n3 == "router" or state_n3 == "child")
        self.assertFormPartitions(1)


if __name__ == '__main__':
    unittest.main()
