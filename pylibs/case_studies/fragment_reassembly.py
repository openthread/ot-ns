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

# Case study on 6LoWPAN fragment reassembly of long (ping) messages.

import logging
from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_NODES = 4

def ping_test(ns, datasz, count):
    id_src = 1
    id_dst = max(ns.nodes())
    for i in range(0,count):
        ns.ping(id_src, id_dst, datasize=datasz)
        ns.go(6)
    ns.print_pings(ns.pings())

def main():
    ns = OTNS(otns_args=['-seed','550','-logfile', 'trace'])
    ns.speed = 1e6
    ns.radiomodel = 'MutualInterference'
    #ns.radiomodel = 'MIDisc'
    #ns.radiomodel = 'Ideal_Rssi'
    ns.set_radioparam('TimeFadingSigmaMaxDb', 0.0)
    ns.web()

    # setup of line topology test network
    for i in range(0, NUM_NODES):
        nid = ns.add("router", x=100 + 175 * i, y=100)
        if i == 0:
            ns.go(10) # Leader starts first
        print(f'Node {nid}: {ns.get_ipaddrs(nid,"mleid")[0]}' )
    ns.go(300)
    ping_test(ns, datasz=4, count=2) # do the address queries for ML-EID destinations

    # do tests and collect KPIs
    ns.kpi_start()
    ping_test(ns, datasz=1150, count = 100)
    ns.kpi_stop()

    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
