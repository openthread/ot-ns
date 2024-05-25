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

# Case study on SRP reregistrations that happen when the Anycast Dataset
# sequence number changes, with a large number of SRP clients.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_BR = 1
NUM_NODES = 50
BR_ID_OFFSET = 100
DX = 150  # pixels spacing

SCRIPT_BR="""
# OT BR CLI script specific to this test scenario

networkname OTSIM
networkkey 00112233445566778899aabbccddeeff
panid 0xface
channel 11
ifconfig up
thread start

# BR wants to become Router early on; never downgrade; always upgrade.
routerselectionjitter 1
routerdowngradethreshold 33
routerupgradethreshold 33

# automatic route to ULA-based AIL
netdata publish route fc00::/7 s med

bbr enable

# use anycast dataset to advertise SRP
srp server addrmode anycast
srp server enable
br init 1 1
br enable
"""

def print_services(srv):
    for line in srv:
        if ':' in line:
            print('\t',line)
        else:
            print(line)

def main():
    ns = OTNS(otns_args=['-seed','34541'])
    ns.speed = 200
    ns.radiomodel = 'MutualInterference'
    ns.web()

    # setup of Border Routers
    for i in range(1, NUM_BR+1):
        nid = ns.add("br", x = i*100, y = 50, id = BR_ID_OFFSET+i, script = SCRIPT_BR)
    nid_br = nid
    ns.go(10)

    # setup of Router nodes - each with service registration using SRP
    cx = 100
    cy = DX
    for i in range(1, NUM_NODES+1):
        nid = ns.add("router", id=i, x = cx, y = cy)
        host = f'EAAFA10F49B12F3{i}'
        ns.node_cmd(nid, f'srp client host name {host}')
        ns.node_cmd(nid, 'srp client host address auto')
        ns.node_cmd(nid, f'srp client service add 15077FD8184910A6-00320000B330{i} _matter._tcp.default.service.arpa,_I1F097FD112451046 18001 0 0 085349493d31303030085341493d31303030085341543d3430303003543d30')
        ns.node_cmd(nid, f'srp client service add 25077FD8184910A6-00320000B330{i} _matter._tcp.default.service.arpa,_I2F097FD112451046 18002 0 0 085449493d31303030085341493d31303030085341543d3430303003543d30')

        ns.go(10)

        cx += DX
        if cx > 1500:
            cx = 100
            cy += DX

    ns.go(290)

    # test anycast dataset seqnum change event
    ns.kpi_start()
    seq = ns.node_cmd(nid_br, "srp server seqnum")
    ns.node_cmd(nid_br, "srp server disable")
    seq = int(seq[0]) + 1
    ns.node_cmd(nid_br, f'srp server seqnum {seq}')
    ns.node_cmd(nid_br, "srp server enable")
    ns.go(100)  # let re-registrations occur
    ns.kpi_stop()

    # check service state from client's viewpoint
    for i in range(1, NUM_NODES+1):
        lines = ns.node_cmd(i, "srp client service")
        for line in lines:
            print(f'{i}: {line}')

    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
