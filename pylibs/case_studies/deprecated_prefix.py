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

# Case study on routing when a prefix becomes deprecated. Requires loading current.pcap
# into Wireshark to see the results.

import logging
from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS()
    ns.logconfig(logging.INFO)
    ns.loglevel = 'info'
    ns.radiomodel = 'MIDisc'
    ns.web()

    # BR - the deprecated prefix is not in network data anymore.
    ns.add("router", x=300, y=100)
    ns.node_cmd(1, 'prefix add fd00:db8::/64 paros med')
    ns.go(10)
    ns.node_cmd(1, 'netdata register')

    # rest of network
    ns.add("router", x=300, y=300)  #
    ns.add("router", x=300, y=500)
    ns.add("fed", x=300, y=700)
    ns.add("med", x=350, y=675)
    ns.add("fed", x=400, y=675)

    # FED / MED have a deprecated (not anymore on-mesh) address
    ns.node_cmd(4, 'ipaddr add 2001:db8::1234')
    ns.node_cmd(5, 'ipaddr add 2001:db8::5678')
    ns.go(20)

    # show IP addresses of children
    ns.node_cmd(4, 'ipaddr -v')
    ns.node_cmd(4, 'netdata show')
    ns.node_cmd(5, 'ipaddr -v')
    ns.node_cmd(6, 'ipaddr -v')

    # BR pings deprecated-prefix address of FED - it fails
    omr_fed = ns.get_ipaddrs(4)[0]
    ns.node_cmd(1, 'ping async 2001:db8::1234')
    ns.go(10)
    ns.node_cmd(1, f'ping async {omr_fed}')
    ns.go(10)
    ns.node_cmd(1, 'eidcache')

    # BR sends addr-query of non-existing on-mesh addr - used once to find the byte sequence for ADDR_QRY UDP payload
    #ns.node_cmd(1,'ping async fd00:db8::1234')
    #ns.go(10)

    # BR sends addr-query to find 2001:db8::1234 FED
    ns.node_cmd(1, 'udp send ff03::2 61631 -x 5202a1e4efb1b161026171ff001020010db8000000000000000000001234')
    ns.go(20)

    # BR sends addr-query to find 2001:db8::5678 MED
    ns.node_cmd(1, 'udp send ff03::2 61631 -x 5202a1e4efb1b161026171ff001020010db8000000000000000000005678')
    ns.go(20)

    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
