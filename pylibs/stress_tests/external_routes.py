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

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

MAX_SIMULATE_SPEED = 1000000


class ExternalRoutesTestbench(object):
    """
    Runs multiple simulation tests where some nodes act as BRs that publish external prefixes, and
    common network data items. The total data size of network data is tracked, as a function of
    the number of border routers (Nbr), number of GUA prefixes (Nglobals) and ULA prefixes (Nulas) that
    are present on the AIL.
    """
    aGlobalPrefixes = ['2001:db8:1234::/64', '2002:db9:5678::/64', '2002:db8:5678::/64', '2002:db7:5678::/64',
                       '2001:4567:5678::/64']
    aUlaPrefixes = ['fd12:3456::/64', 'fc00:abcd::/64', 'fd34:1234:5678:abcd:abcd::/80', 'fdab:1234:1234::/64',
                    'fd0a:c394:a051:8d93::/64']

    ns = None
    sz = []
    netdata = []

    def display_tab_separated(self, disp):
        s = ""
        for item in disp:
            s += str(item) + "\t"
        print(s)

    def setup_otns(self):
        self.ns = OTNS(otns_args=["-log", "warn"])
        # self.ns.web()
        self.ns.speed = MAX_SIMULATE_SPEED
        return self.ns

    def remove_all_nodes(self):
        ids = self.ns.nodes().keys()
        for id in ids:
            self.ns.delete(id)

    def add_route(self, nodeId, prefix):
        self.ns.node_cmd(nodeId, 'netdata publish route %s s med' % prefix )

    def track_netdata_size(self):
        self.ns.go(10)  # Time for Leader to collect new data.
        self.sz.append(len(self.ns.node_cmd(1, 'netdata show -x')[0])/2)  # Count datalen based on hex string.
        self.netdata.append( '\n'.join(self.ns.node_cmd(1, 'netdata show')) )

    def run_topology(self, Nbr=1, Nglobals=1, Nulas=1):
        ns = self.setup_otns()
        self.sz = []
        self.netdata = []

        # Leader
        ns.add("router", x=300, y=300)
        ns.go(10)

        # Border Routers
        for k in range(0, Nbr):
            id = ns.add("router", x=300 + 100 + k * 50, y=300)
            ns.node_cmd(id, 'netdata publish prefix fd00:dead:beef::/64 paosr med') # OMR prefix
            ns.node_cmd(id, 'netdata publish dnssrp anycast 1') # DNS/SRP Anycast Dataset
            ns.node_cmd(id, 'netdata publish dnssrp unicast 2001:db8:1234::1234 53') # DNS/SRP Unicast Dataset (external srv)
            ns.node_cmd(id, 'netdata publish route 64:ff9b::/96 sn med') # RFC 6146 address
            ns.go(10)

        self.track_netdata_size()

        # Set ext routes
        N = max(Nglobals, Nulas)
        for j in range(0, N):
            for k in range(0, Nbr):
                if j < Nglobals:
                    self.add_route(k+1, self.aGlobalPrefixes[j])
                    self.track_netdata_size()
                if j < Nulas:
                    self.add_route(k+1, self.aUlaPrefixes[j])
                    self.track_netdata_size()

        ns.go(1)
        self.remove_all_nodes()
        ns.go(100)
        ns.close()
        return self.sz, self.netdata

    def run_tests(self):
        self.display_tab_separated( ('Nbr', 'Nglob', 'Nula', 'Size[B]', 'MaxSz[B]') )
        for Nbr in range(1, 13):
            for Nglobals in range(0, 4):
                for Nulas in range(0, 5):
                    sz, netdata = self.run_topology(Nbr, Nglobals, Nulas)
                    for k in range(0, len(sz)):
                        self.display_tab_separated( (Nbr, Nglobals, Nulas, round(sz[k]), round(max(sz))) )
                        #print(netdata[k])


if __name__ == '__main__':
    try:
        ExternalRoutesTestbench().run_tests()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
