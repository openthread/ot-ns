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

# Case study on SRP reregistrations that happen when the OMR prefix changes.

import logging
from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_BR = 2
NUM_NODES = 10
BR_ID_OFFSET = 100

SCRIPT_BR = """
# OT BR CLI script to configure a BR to use its own OMR prefix, if needed.
# There is no DHCPv6-PD available to get a prefix from.

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
srp server enable
br init 1 1
br enable
"""


def print_services(srv):
    for line in srv:
        if ':' in line:
            print('\t', line)
        else:
            print(line)


def main():
    ns = OTNS(otns_args=['-seed', '34541', '-logfile', 'info'])
    ns.speed = 200
    ns.radiomodel = 'MutualInterference'
    ns.web()

    # Leader
    ns.add("router", id=200)
    ns.go(10)

    # setup of Border Routers
    for i in range(1, NUM_BR + 1):
        ns.add("br", id=BR_ID_OFFSET + i, script=SCRIPT_BR)
    ns.go(10)

    # setup of Router nodes - each with service registration using SRP
    for i in range(1, NUM_NODES + 1):
        nid = ns.add("router", id=i)
        host = f'host{i}'
        port = 18000 + i
        ns.node_cmd(nid, f'srp client host name {host}')
        ns.node_cmd(nid, 'srp client host address auto')
        ns.node_cmd(nid, f'srp client service add instance{i} _test._udp {port}')
        ns.go(10)
    ns.go(20)

    for i in range(1, NUM_BR + 1):
        nid = BR_ID_OFFSET + i
        print(f"Services registered on BR node {nid}:")
        print_services(ns.node_cmd(nid, 'srp server service'))

    ns.kpi_start()
    ns.delete(102)  # delete the BR that provides winning OMR prefix
    ns.go(700)  # timeout removed BR info and let re-registrations occur
    ns.kpi_stop()

    for i in range(1, NUM_BR + 1):
        nid = BR_ID_OFFSET + i
        if nid == 102:
            continue
        print(f"Services registered on BR node {nid} - after SRP re-registrations:")
        print_services(ns.node_cmd(nid, 'srp server service'))

    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
