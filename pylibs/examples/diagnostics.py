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

#
# Network diagnostics example.

import logging
import sys

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

def display(list):
    for line in list:
        print(line, file=sys.stderr)

def main():
    logging.basicConfig(level=logging.WARNING)

    display(["", "diagnostics.py example"])
    ns = OTNS(otns_args=["-log", "warn"])
    ns.radiomodel = 'MIDisc'
    #ns.watch_default('trace')
    ns.set_title("Network Diagnostics Example")
    ns.web()

    nid_cl=ns.add("router", x=600, y=300) #Leader
    ns.go(10)
    nid_srv=ns.add("router", x=600, y=500, version="v131") # Router DUT
    ns.go(10)
    # Add Children to the Router
    ns.add("med")
    ns.add("sed")
    ns.add("sed")
    ns.add("sed")
    ns.add("fed")
    ns.go(10)
    # Add a neighbour Router
    ns.add("router",x=670,y=320)
    ns.go(130)

    display(["","Send DIAG_GET.req to RLOC"])
    a_rloc = ns.get_ipaddrs(nid_srv,'rloc')[0]
    display(ns.node_cmd(nid_cl,f'networkdiagnostic get {a_rloc} 19 23 24 25 26 27 28'))
    display(ns.go(60)) # command runs in the background - this collects the output

    display(["","Send DIAG_GET.req to ML-EID (from Node 8)"])
    a_mleid = ns.get_ipaddrs(nid_srv,'mleid')[0]
    display(ns.node_cmd(8,f'networkdiagnostic get {a_mleid} 5 19 23 24 25 26 27 28 34'))
    display(ns.go(60))

    display(["","Send DIAG_GET.qry to RLOC16"])
    a_rloc16 = ns.get_rloc16(nid_srv)
    display(ns.node_cmd(8,f'meshdiag routerneighbortable {a_rloc16}'))
    display(ns.go(60))

    # allow some time for graphics to be displayed in web GUI.
    ns.speed=0.001
    ns.go(0.001)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
