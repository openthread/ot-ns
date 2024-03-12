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
#

import filecmp
import logging
import os.path
import time
import unittest
from typing import Dict

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS


class BasicTests(OTNSTestCase):
    def testGetSetSpeed(self):
        ns = self.ns
        self.assertEqual(ns.speed, OTNS.DEFAULT_SIMULATE_SPEED)
        ns.speed = 2
        self.assertEqual(ns.speed, 2)
        ns.speed = float('inf')
        self.assertEqual(ns.speed, OTNS.MAX_SIMULATE_SPEED)

    def testGetSetPlr(self):
        ns = self.ns
        assert ns.packet_loss_ratio == 0
        ns.packet_loss_ratio = 0.5
        assert ns.packet_loss_ratio == 0.5
        ns.packet_loss_ratio = 1
        assert ns.packet_loss_ratio == 1
        ns.packet_loss_ratio = 2
        assert ns.packet_loss_ratio == 1

    def testOneNodex100(self):
        for i in range(100):
            logging.info("testOneNodex100 round %d", i + 1)
            ns = self.ns
            ns.add("router")
            ns.go(10)
            self.assertFormPartitions(1)
            self.tearDown()
            self.setUp()

    def testAddNode(self):
        ns = self.ns
        ns.add("router")
        self.go(10)
        self.assertFormPartitions(1)

        ns.add("router")
        ns.add("fed")
        ns.add("med")
        ns.add("sed")
        ns.add("ssed")

        self.go(33)
        self.assertFormPartitions(1)

    def testAddNodeWithID(self):
        ns = self.ns
        for new_id in [50, 55, 60]:
            nid = ns.add("router", id=new_id)
            self.assertEqual(nid, new_id)
            self.go(1)
        self.go(130)
        self.assertFormPartitions(1)

    def testAddNodeWithExistingID(self):
        ns = self.ns
        new_id = 50
        nid = ns.add("router", id=new_id)
        self.assertEqual(nid, new_id)
        self.go(1)
        self.assertRaises(errors.OTNSCliError, lambda: ns.add("router", id=new_id))

    def testAddNodeWithRawFlag(self):
        ns = self.ns
        nid = ns.add("router", script="# No CLI commands send in this node-script.")
        self.assertEqual(1,nid)
        ns.add("router")
        ns.add("router")
        self.go(20)
        pars = self.ns.partitions()
        self.assertEqual(2,len(pars))
        self.assertEqual([1],pars[0]) # Node 1 is expected to be unconnected

    def testRestoreNode(self):
        ns = self.ns
        ns.add("router", x=0, y=0)

        self.go(10)
        self.assertEqual(ns.get_state(1), "leader")

        n=0
        for type in ("router", "fed", "med", "sed"):
            nodeid = ns.add(type, x=n*10, y=10)
            self.go(10)
            self.assertFormPartitions(1)
            rloc16 = ns.get_rloc16(nodeid)
            print('rloc16', rloc16)

            ns.delete(nodeid)
            ns.go(10)

            self.assertEqual(nodeid, ns.add(type, x=n*10, y=10, restore=True))
            self.assertEqual(rloc16, ns.get_rloc16(nodeid))
            self.go(0.1)
            while len(ns.partitions()) > 1 and ns.time < 100:
                self.go(0.1)
            self.assertFormPartitions(1)
            n += 1

    def testMoveNode(self):
        ns = self.ns
        ns.add("router", x=100, y=200)
        ns.add("router", x=200, y=200)
        ns.go(25)
        self.assertFormPartitions(1)
        ns.move(2, 132000, 132000)
        ns.go(180)
        self.assertFormPartitions(2)
        ns.move(2, 200, 200, 50) # move in 3D
        ns.go(250)
        self.assertFormPartitions(1)
        ns.move(2, 198, 199) # move in 2D plane only. Z stays as it was.
        ns.go(180)
        node_info = ns.nodes()[2]
        print(node_info)
        self.assertEqual(198, node_info['x'])
        self.assertEqual(199, node_info['y'])
        self.assertEqual(50, node_info['z'])
        self.assertFormPartitions(1)

    def testDelNode(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        self.go(25)
        self.assertFormPartitions(1)
        ns.delete(1)
        self.go(10)
        self.assertTrue(len(ns.nodes()) == 1 and 1 not in ns.nodes())

    def testDelManyNodes(self):
        for j in range(4):
            ns = self.ns
            many = 32

            for i in range(many):
                ns.add("router", x=(i % 6) * 100, y=(i // 6) * 150)

            ns.go(10)
            for i in range(1, many + 1):
                ns.delete(i)
                ns.go(5)

            self.assertTrue(ns.nodes() == {})
            self.tearDown()
            self.setUp()

    def testDelNodeAndImmediatelyRecreate(self):
        # repeat multiple times to catch some goroutine race conditions that only happen sometimes.
        for i in range(100):
            logging.info("testDelNodeAndImmediatelyRecreate round %d", i + 1)

            ns = self.ns
            ns.loglevel = 'debug'
            ns.watch_default('debug') # add extra detail in all node's logs
            id = ns.add("router")
            self.assertTrue(len(ns.nodes()) == 1 and 1 in ns.nodes() and id == 1)
            self.go(i/100)
            self.assertTrue(len(ns.nodes()) == 1 and 1 in ns.nodes())

            ns.delete(1)
            self.assertTrue(len(ns.nodes()) == 0)
            id = ns.add("router")
            self.assertTrue(len(ns.nodes()) == 1 and 1 in ns.nodes() and id == 1)

            ns.add("router")
            ns.add("router")
            id = ns.add("router")
            self.assertTrue(len(ns.nodes()) == 4 and id == 4)

            ns.delete(1, 2, 3, 4)
            self.assertTrue(len(ns.nodes()) == 0)
            if i>90:
                ns.go(0)

            ns.add("router")
            id = ns.add("router")
            self.assertTrue(len(ns.nodes()) == 2 and id == 2)

            self.tearDown()
            self.setUp()

    def testDelNonExistingNodes(self):
        ns = self.ns
        ns.add("router")
        ns.add("router")
        self.go(25)
        self.assertFormPartitions(1)
        ns.delete(1,3,4,5)
        self.go(10)
        self.assertTrue(len(ns.nodes()) == 1 and 1 not in ns.nodes())

    def testPlrEffective(self):
        ns = self.ns
        ns.packet_loss_ratio = 1
        self.assertTrue(ns.packet_loss_ratio, 1)
        ns.add("router")
        ns.add("router")
        ns.add("router")
        self.go(100)
        self.assertFormPartitions(3)

    def testRadioInRange(self):
        ns = self.ns
        ns.radiomodel = 'Ideal'
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range - 1, radio_range=radio_range)
        self.go(15)
        self.assertFormPartitions(1)

    def testRadioNotInRange(self):
        ns = self.ns
        ns.radiomodel = 'Ideal'
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, radio_range + 1, radio_range=radio_range)
        self.go(10)
        self.assertFormPartitions(2)

    def testNodeFailRecover(self):
        ns = self.ns
        ns.add("router")
        fid = ns.add("router")
        self.go(20)
        self.assertFormPartitions(1)

        ns.radio_off(fid)
        self.go(240)
        print(ns.partitions())
        self.assertFormPartitions(2)

        ns.radio_on(fid)
        self.go(100)
        self.assertFormPartitions(1)

    def testFailTime(self):
        ns = self.ns
        id = ns.add("router")
        ns.radio_set_fail_time(id, fail_time=(2, 10))
        total_count = 0
        failed_count = 0
        for i in range(1000):
            ns.go(1)
            nodes = ns.nodes()
            failed = nodes[id]['failed']
            total_count += 1
            failed_count += failed

        self.assertAlmostEqual(failed_count / total_count, 0.2, delta=0.1)

        # fail-interval param must be greater than fail-duration.
        with self.assertRaises(errors.OTNSCliError):
            ns.radio_set_fail_time(id, fail_time=(18,16))

    def testCliCmd(self):
        ns = self.ns
        id = ns.add("router")
        self.go(10)
        self.assertEqual('leader', ns.get_state(id))

    def testCounters(self):
        ns = self.ns

        def assert_increasing(c0: Dict[str, int], c1: Dict[str, int]):
            for k0, v0 in c0.items():
                self.assertGreaterEqual(c1.get(k0, 0), v0)
            for k1, v1 in c1.items():
                self.assertGreaterEqual(v1, c0.get(k1, 0))

        c0 = counters = ns.counters()
        self.assertTrue(counters)
        self.assertTrue(all(x == 0 for x in counters.values()))
        ns.add("router")
        ns.add("router")
        ns.add("router")

        self.go(10)
        c10 = ns.counters()
        assert_increasing(c0, c10)

        self.go(10)
        c20 = ns.counters()
        assert_increasing(c10, c20)

    def testConfigVisualization(self):
        ns = self.ns
        vopts = ns.config_visualization()
        print('vopts', vopts)
        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertTrue(opt in vopts)

            set_vals = (False, True) if vopts[opt] else (True, False)
            for v in set_vals:
                vopts[opt] = v
                self.assertTrue(ns.config_visualization(**{opt: v}) == vopts)

        vopts = ns.config_visualization(broadcast_message=True, unicast_message=True, ack_message=True,
                                        router_table=True,
                                        child_table=True)

        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertTrue(vopts[opt])

        vopts = ns.config_visualization(broadcast_message=False, unicast_message=False, ack_message=False,
                                        router_table=False,
                                        child_table=False)

        for opt in ('broadcast_message', 'unicast_message', 'ack_message', 'router_table', 'child_table'):
            self.assertFalse(vopts[opt])

    def testWithOTNS(self):
        """
        make sure OTNS works in with-statement
        """
        self.tearDown()

        with OTNS(otns_args=['-log', 'debug']) as ns:
            self.assertEqual(OTNS.DEFAULT_SIMULATE_SPEED, ns.speed)
            ns.speed = 19999
            nid = ns.add("router")
            self.assertEqual(1, nid)
            ns.go(10)
            self.assertEqual(10, ns.time)

        # run a second time to make sure the previous simulation is properly terminated
        with OTNS(otns_args=['-log', 'warn', '-speed', '18123']) as ns:
            self.assertEqual(18123, ns.speed)
            nid = ns.add("router")
            self.assertEqual(1, nid)
            ns.go(10)
            self.assertEqual(10, ns.time)

        with OTNS() as ns:
            ns.add('router')
            ns.add('router')
            self.assertEqual(OTNS.DEFAULT_SIMULATE_SPEED, ns.speed)
            self.assertEqual(0, ns.time)

    def testSetRouterUpgradeThreshold(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        self.assertEqual(16, ns.get_router_upgrade_threshold(nid))
        for val in range(0, 33):
            ns.set_router_upgrade_threshold(nid, val)
            self.assertEqual(val, ns.get_router_upgrade_threshold(nid))

    def testSetRouterUpgradeThresholdEffective(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        ns.go(10)
        self.assertNodeState(nid, 'leader')

        reed = ns.add("router")
        ns.set_router_upgrade_threshold(reed, 1)
        ns.go(130)
        self.assertNodeState(reed, 'child')

        ns.set_router_upgrade_threshold(reed, 2)
        ns.go(150)
        self.assertNodeState(reed, 'router')

    def testSetRouterDowngradeThreshold(self):
        ns: OTNS = self.ns
        nid = ns.add("router")
        self.assertEqual(23, ns.get_router_downgrade_threshold(nid))
        for val in range(0, 33):
            ns.set_router_downgrade_threshold(nid, val)
            self.assertEqual(val, ns.get_router_downgrade_threshold(nid))

    def testCoaps(self):
        ns: OTNS = self.ns
        ns.coaps_enable()
        for i in range(10):
            id = ns.add('router', x=i*10, y=0)
            ns.node_cmd(id, 'routerselectionjitter 1')
            ns.go(5)

        ns.go(10)
        msgs = ns.coaps()
        routers = {}
        for msg in msgs:
            if msg.get('uri') == 'a/as':
                routers[msg['src']] = msg['id']

        # Node 2 ~ 10 should become Routers by sending `a/as`
        self.assertEqual(set(routers), set(range(2, 11)))

    def testMultiRadioChannel(self):
        ns = self.ns
        radio_range = 100
        ns.add("router", 0, 0, radio_range=radio_range)
        ns.add("router", 0, 50, radio_range=radio_range)
        ns.add("router", 50, 0, radio_range=radio_range)
        ns.add("router", 50, 50, radio_range=radio_range)
        self.go(130)
        self.assertFormPartitions(1)

        for n in [1,2]:
            ns.node_cmd(n, "ifconfig down")
            ns.node_cmd(n, "channel 20")
            ns.node_cmd(n, "ifconfig up")
            ns.node_cmd(n, "thread start")
        self.go(300)
        self.assertFormPartitions(2)

    def testLoglevel(self):
        ns: OTNS = self.ns
        ns.loglevel = "warn"
        id = ns.add("router")
        self.go(10)
        self.assertEqual(ns.loglevel, "warn")
        ns.loglevel = "debug"
        id = ns.add("router")
        self.go(10)
        self.assertEqual(ns.loglevel, "debug")
        with self.assertRaises(errors.OTNSCliError):
            ns.loglevel = "invalid_log_level"
        self.assertEqual(ns.loglevel, "debug")
        ns.loglevel = "info"
        ns.loglevel = "error"

    def testWatch(self):
        ns: OTNS = self.ns
        for i in range(10):
            ns.add('router')
            ns.go(2)
        ns.watch(3, 4, 5, 6, 8)
        ns.go(5)
        ns.unwatch(5, 6)
        self.assertEqual([3, 4, 8], ns.watched())
        ns.unwatch_all()
        self.assertEqual([], ns.watched())
        ns.watch_all('debug')
        self.assertEqual([1,2,3,4,5,6,7,8,9,10], ns.watched())
        ns.watch_all('warn')
        self.assertEqual([1,2,3,4,5,6,7,8,9,10], ns.watched())

    def testWatchDefault(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.watch_default('trace')
        for i in range(9):
            ns.add('router')
            ns.go(2)
        self.assertEqual([2,3,4,5,6,7,8,9,10], ns.watched())
        ns.watch_default('off')
        ns.add('router')
        self.assertEqual([2,3,4,5,6,7,8,9,10], ns.watched())
        ns.watch_default('info')
        ns.add('router')
        self.assertEqual([2,3,4,5,6,7,8,9,10, 12], ns.watched())

    def testWatchNonExistingNodes(self):
        ns: OTNS = self.ns
        for i in range(10):
            ns.add('router')
            ns.go(2)
        with self.assertRaises(errors.OTNSCliError):
            ns.watch(3, 4, 11) # node 11 does not exist
        ns.go(5)
        self.assertEqual([3, 4], ns.watched())
        ns.unwatch(5, 6) # nodes not being watched, but no error.
        self.assertEqual([3, 4], ns.watched())
        ns.unwatch(3, 5, 66) # 3 is watched but 5 not, 66 non-existing, but no error.
        self.assertEqual([4], ns.watched())

    def testHelp(self):
        ns: OTNS = self.ns
        ns.cmd("help")
        ns.cmd("help plr")
        ns.cmd("help radiomodel")
        ns.cmd("help add")
        ns.cmd("help exe")

    def testGoUnits(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.add('router')

        ns.go(10)
        self.assertEqual(10.0, ns.time) # ns.time returns microseconds
        ns.go(0.001)
        self.assertEqual(10.001, ns.time)
        ns.go(1e-3)
        self.assertEqual(10.002, ns.time)
        ns.go(1e-6)
        self.assertEqual(10.002001, ns.time)
        ns.go(3e-5)
        self.assertEqual(10.002031, ns.time)
        ns.go(1e-7)
        self.assertEqual(10.002031, ns.time) # no time advance: rounded to nearest microsecond.
        ns.go(0.000999) # almost 1 ms
        self.assertEqual(10.003030, ns.time)
        ns.go(4.0000004)
        self.assertEqual(14.003030, ns.time) # rounded to nearest microsecond.

    def testScan(self):
        ns: OTNS = self.ns
        ns.radiomodel = 'MutualInterference'
        ns.add('router')

        with self.assertRaises(errors.OTNSCliError):
            ns.cmd("scan 2")

        ns.add('router')
        ns.add('router')
        ns.add('router', x=100, y=200)
        ns.add('router')
        ns.add('router')

        ns.go(50)
        ns.cmd("scan 1")
        ns.go(5)
        ns.cmd("scan 2")
        ns.go(1)
        ns.cmd("scan 3")
        ns.go(15)
        ns.speed = 1
        ns.cmd("scan 6")
        # no go() period at the end to test starting scan and immediately stopping simulation.

    def testInvalidNodeCmd(self):
        ns: OTNS = self.ns
        with self.assertRaises(errors.OTNSCliError):
            ns.node_cmd(1,'state')
        ns.add('router')
        ns.node_cmd(1,'state')
        ns.go(20)

        ns.node_cmd(1,'dns config 2001::1234 1234 5000 2 0 srv_txt_sep udp')
        with self.assertRaises(errors.OTNSCliError):
            ns.node_cmd(1,'sdfklsjflksj')
        with self.assertRaises(errors.OTNSCliError):
            ns.node_cmd(1,'dns config nonexistoption')
        ns.node_cmd(1,'dns config')

        ns.node_cmd(1,'dns resolve nonexistent.example.com')
        ns.go(30) # error response comes during the go period.

        with self.assertRaises(errors.OTNSCliError):
            ns.node_cmd(1,'dns resolvea b c d e f')

        ns.go(1)

    def testExitCmd(self):
        ns: OTNS = self.ns
        # Tested here that the exit command itself does not raise errors.OTNSExitedError.
        # That error is only raised for unexpected exits of OTNS, or when performing an
        # action while OTNS has already exited.
        ns.cmd('exit')

        with self.assertRaises(errors.OTNSExitedError):
            ns.go(1)
        with self.assertRaises(errors.OTNSExitedError):
            ns.add('router')
        with self.assertRaises(errors.OTNSExitedError):
            ns.cmd('exit')

    def testAutoGo(self):
        ns: OTNS = self.ns
        self.assertFalse(ns.autogo)

        ns.add('router')
        ns.add('router')
        ns.add('router')
        ns.go(60)
        self.assertEqual(60, ns.time)

        # With 1 realtime second of autogo, simulation moves approx speed * 1 =~ 5 seconds forward.
        # It can be somewhat lower if the simulation doesn't manage to run at requested speed.
        t1 = ns.time
        ns.speed = 5
        ns.autogo = True
        time.sleep(3)
        self.assertTrue(ns.time > 11 + t1)
        self.assertTrue(ns.autogo)

        # When autogo is disabled, it finishes the current autogo duration of 1 second.
        ns.autogo = False
        t2 = ns.time
        time.sleep(1)
        self.assertTrue(ns.time <= t2 + 1)
        self.assertFalse(ns.autogo)

    def testRxSensitivity(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.add('router')
        self.assertEqual(['-100'], ns.cmd('rfsim 1 rxsens'))
        self.assertEqual(['-100'], ns.cmd('rfsim 2 rxsens'))

        ns.cmd('rfsim 1 rxsens -85')
        self.assertEqual(['-85'], ns.cmd('rfsim 1 rxsens'))
        self.assertEqual(['-100'], ns.cmd('rfsim 2 rxsens'))

        ns.cmd('rfsim 2 rxsens 23')
        self.assertEqual(['23'], ns.cmd('rfsim 2 rxsens'))

    def testCcaThreshold(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.add('router')
        self.assertEqual(['-75 dBm'], ns.node_cmd(1,'ccathreshold'))
        self.assertEqual(['-75 dBm'], ns.node_cmd(2,'ccathreshold'))
        self.assertEqual(['-75'], ns.cmd('rfsim 1 ccath'))

        ns.node_cmd(2, 'ccathreshold -80')
        self.assertEqual(['-75 dBm'], ns.node_cmd(1,'ccathreshold'))
        self.assertEqual(['-80 dBm'], ns.node_cmd(2,'ccathreshold'))
        self.assertEqual(['-80'], ns.cmd('rfsim 2 ccath'))

        ns.node_cmd(1, 'ccathreshold 42')
        self.assertEqual(['42 dBm'], ns.node_cmd(1,'ccathreshold'))
        self.assertEqual(['-80 dBm'], ns.node_cmd(2,'ccathreshold'))
        self.assertEqual(['42'], ns.cmd('rfsim 1 ccath'))

    def testCmdCommand(self):
        ns: OTNS = self.ns
        output = ns.cmd('autogo') # arbitrary command
        self.assertEqual(1, len(output))
        self.assertEqual('0', output[0])

        output = ns.cmd('') # test empty command (like pressing enter)
        self.assertEqual(0, len(output))

    def testRandomSeedSetting(self):
        self.tearDown()
        nodes = range(1,6)

        # create a new OTNS with 'seed' parameter.
        with OTNS(otns_args=['-log', 'debug', '-seed', '20242025']) as ns:
            self.ns = ns

            mleid_addr = ['']
            for i in nodes:
                ns.add("router")
                mleid_addr.append(str(ns.get_ipaddrs(i, 'mleid')[0]))
            ns.go(150)
            self.assertFormPartitions(1)

            node_states = ['']
            for i in nodes:
                node_states.append(ns.get_state(i))

        # after closing of OTNS, save PCAP file of simulation
        pcap_path = "tmp/unittest_pcap"
        pcap_fn_1 = self.name() + "_session_1.pcap"
        self.ns.save_pcap(pcap_path, pcap_fn_1)

        # create a new OTNS with same 'seed' parameter. By using same seed, it becomes
        # predictable who will become the Leader and what random address a node will use, etc.
        # The PCAP output will also be identical.
        with OTNS(otns_args=['-log', 'debug', '-seed', '20242025']) as ns:
            self.ns = ns

            for i in nodes:
                ns.add("router")
                self.assertEqual(mleid_addr[i], str(ns.get_ipaddrs(i, 'mleid')[0]))
            ns.go(150)
            self.assertFormPartitions(1)

            for i in nodes:
                self.assertNodeState(i, node_states[i])

        # after closing OTNS, save PCAP for second simulation
        pcap_fn_2 = self.name() + "_session_2.pcap"
        self.ns.save_pcap(pcap_path, pcap_fn_2)
        cmp_res = filecmp.cmp(os.path.join(pcap_path, pcap_fn_1), os.path.join(pcap_path, pcap_fn_2), shallow = False)
        self.assertTrue(cmp_res)

    def testKpi(self):
        ns: OTNS = self.ns

        self.assertTrue(ns.kpi())
        ns.add('router')
        ns.add('router')
        ns.add('router')
        ns.go(50)

        ns.kpi_start() # restart KPIs
        self.assertTrue(ns.kpi())
        kpi_data = ns.kpi_save()
        self.assertEqual(0, kpi_data["time_sec"]["duration"])
        self.assertEqual(50, kpi_data["time_sec"]["start"])

        for n in range(0,10):
            ns.ping(1,3)
            ns.go(5)
        ns.kpi_stop()
        self.assertFalse(ns.kpi())

        kpi_data = ns.kpi_save()

        self.assertFalse(ns.kpi())
        self.assertEqual(50, kpi_data["time_sec"]["duration"])
        self.assertEqual(100, kpi_data["time_sec"]["end"])
        self.assertIsNotNone(kpi_data["created"])
        self.assertEqual("ok", kpi_data["status"])
        self.assertIsNotNone(kpi_data["time_us"])
        self.assertIsNotNone(kpi_data["mac"])
        self.assertEqual(3, len(kpi_data["counters"]))

        ns.go(30)
        self.assertFalse(ns.kpi())
        self.assertEqual(50, kpi_data["time_sec"]["duration"])
        self.assertEqual(100, kpi_data["time_sec"]["end"])

        ns.kpi_start()
        self.assertTrue(ns.kpi())
        ns.go(50)
        ns.delete(1)  # delete a node just before KPI collection is done
        ns.go(5)

        # save while running
        kpi_data = ns.kpi_save('tmp/unittest_kpi_test.json')
        self.assertTrue(ns.kpi())
        self.assertEqual("ok", kpi_data["status"])
        self.assertEqual(2, len(kpi_data["counters"]))

    def testLoadYamlTopology(self):
        ns: OTNS = self.ns
        self.assertEqual(0,len(ns.nodes()))
        ns.load('pylibs/test_mesh_topology.yaml')
        self.assertEqual(57,len(ns.nodes()))
        ns.go(1)

    def testSaveYamlTopology(self):
        ns: OTNS = self.ns
        self.assertEqual(0,len(ns.nodes()))
        ns.add('router')
        ns.go(10)
        ns.add('router')
        ns.add('router', version='')
        ns.add('router')
        ns.add('ssed')
        ns.go(25)
        self.assertEqual(5,len(ns.nodes()))
        self.assertFormPartitions(1)

        ns.save('tmp/unittest_save_topology.yaml')
        self.assertEqual(5,len(ns.nodes()))

        ns.delete(1,2,3,4,5)
        self.assertEqual(0,len(ns.nodes()))

        ns.load('tmp/unittest_save_topology.yaml')
        self.assertEqual(5,len(ns.nodes()))
        ns.go(125)
        self.assertFormPartitions(1)

    def testRealtimeMode(self):
        self.tearDown()
        with OTNS(otns_args=['-log', 'debug', '-realtime']) as ns:
            ns.add('router')
            ns.add('router')
            self.assertEqual(True, ns.autogo)
            self.assertEqual(1.0, ns.speed)
            ns.speed = 23
            self.assertEqual(1.0, ns.speed)

    def testClockDriftSetting(self):
        ns: OTNS = self.ns
        ns.add('router')
        ns.add('router')
        ns.add('router')

        ns.set_node_clock_drift(1,  0)
        ns.set_node_clock_drift(2, 20)
        ns.set_node_clock_drift(3, -1)

        ns.go(1000)

        # each node reports a different uptime value, due to their different clock drifts.]
        self.assertEqual(1000.000, ns.get_node_uptime(1))
        self.assertEqual(round(1000.0 * (1 + 20e-6),3), ns.get_node_uptime(2))
        self.assertEqual(round(1000.0 * (1 -  1e-6),3), ns.get_node_uptime(3))

        ns.set_node_clock_drift(2, 0)
        ns.go(1000)
        n2_uptime = round(1000.0 * (1 + 20e-6) + 1000.0, 3)
        self.assertEqual(n2_uptime, ns.get_node_uptime(2))

        # delete other nodes for faster simulation
        ns.delete(1,3)
        ns.go(24*3600) # simulate 1 full day - to test the 'uptime' parsing of day values.
        self.assertEqual(n2_uptime + 24*3600.0, ns.get_node_uptime(2))


if __name__ == '__main__':
    unittest.main()
