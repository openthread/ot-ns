#!/usr/bin/env python3
# Copyright (c) 2023-2024, The OTNS Authors.
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

import math
import random
import unittest

from OTNSTestCase import OTNSTestCase
from otns.cli import OTNS


class TopologiesTests(OTNSTestCase):

    def testLargeNetwork(self):
        n = 1024 # network size (nodes)
        ns = self.ns
        ns.loglevel = 'info'
        ns.add("router")
        ns.go(10)
        self.assertEqual("leader", ns.get_state(1))

        for i in range(n-1):
            ns.add("router")
            ns.go(0.001)

        ns.go(1)
        self.assertEqual(n,len(ns.nodes()))
        self.assertEqual("leader", ns.get_state(1))

    def testDenseNetwork(self):
        nn = 144 # number of nodes
        x0 = 100 # start coordinate (x,y) (pixels)
        dx = 100 # coordinate delta x/y between nodes (pixels)
        rr = 700 # radio range (pixels)
        probability_med = 0.20 # probability that new node is a MED

        ns = self.ns
        ns.loglevel = 'info'
        x = x0
        y = x0
        for i in range(nn):
            if random.random() < probability_med:
                ns.add("med", x, y, radio_range=rr)
            else:
                ns.add("router", x, y, radio_range=rr)
            ns.go(1)
            x += dx
            if x >= (x0+math.sqrt(nn)*dx):
                x = x0
                y += dx

        ns.go(20)
        self.assertEqual(nn,len(ns.nodes()))

    def testMultiChannel(self):
        self.ns.close()
        # use ot-script option to prevent standard network data init script to run.
        self.ns = OTNS(otns_args=['-ot-script','none'])
        ns = self.ns
        ns.loglevel = 'info'
        ns.watch_default('warn') # show errors+warnings from all OT nodes

        n_netw_group = [2, 3] # number of different network-groups [rows,cols]
        n_netw_groups = n_netw_group[0] * n_netw_group[1]
        n_node_group = [4, 4] # nodes per Thread Network (i.e. channel) [rows,cols]
        gdx = 500
        gdy = 400
        ndx = 70
        ndy = 70
        ofs_x = 100
        ofs_y = 100
        ng = 1 # number of group (network) a node is in.
        for rg in range(0,n_netw_group[0]):
            for cg in range(0,n_netw_group[1]):
                for rn in range(0,n_node_group[0]):
                    for cn in range(0,n_node_group[1]):
                        nid = ns.add('router', x=ofs_x+cg*gdx+cn*ndx, y=ofs_y+rg*gdy+rn*ndy)
                        self.setup_node_for_group(nid, ng)
                ng += 1

        ns.go(10)
        self.assertTrue(len(ns.partitions()) > n_netw_groups)
        ns.go(120)
        self.assertFormPartitions(n_netw_groups)

    # executes a startup script on each node, params depending on each group (ngrp)
    def setup_node_for_group(self, nid, ngrp):
        chan = ngrp-1 + 11
        self.ns.set_network_name(nid,f"Netw{ngrp}_Chan{chan}")
        self.ns.set_panid(nid,ngrp)
        self.ns.set_extpanid(nid,ngrp)
        self.ns.set_networkkey(nid,f"{ngrp:#0{34}x}"[2:])
        self.ns.set_channel(nid,chan)
        self.ns.ifconfig_up(nid)
        self.ns.thread_start(nid)


if __name__ == '__main__':
    unittest.main()
