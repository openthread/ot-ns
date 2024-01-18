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
# TCP connection and benchmarking example.

import logging
import sys

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

def display(list):
    for line in list:
        print(line, file=sys.stderr)

def main():
    #logging.basicConfig(level=logging.WARNING)
    logging.basicConfig(format='%(asctime)s - %(levelname)s - %(message)s', level=logging.INFO)

    display(["", "diagnostics.py example"])
    ns = OTNS(otns_args=["-log", "info"])
    ns.set_title("TCP Example")
    ns.watch_default('warn')
    ns.web()

    nid_cl=ns.add("router", x=100, y=300)
    ns.add("router", x=300, y=300)
    ns.add("router", x=500, y=300)
    nid_srv=ns.add("router", x=700, y=300)
    mleid = ns.get_mleid(nid_srv)
    ns.go(130)

    # for TCP commands, see https://github.com/openthread/openthread/blob/main/src/cli/README_TCP.md
    ns.node_cmd(nid_srv,'tcp init')
    ns.node_cmd(nid_srv,'tcp listen :: 30000')
    ns.go(2)
    ns.node_cmd(nid_cl,'tcp init')
    ns.node_cmd(nid_cl,f'tcp connect {mleid} 30000')
    ns.go(2)

    # test connection
    ns.node_cmd(nid_cl,'tcp send hello')
    ns.go(10)

    # benchmark
    ns.node_cmd(nid_cl,'tcp benchmark run')
    ns.go(20)
    ns.node_cmd(nid_cl,'tcp benchmark result')

    # allow some time for graphics to be displayed in web GUI.
    ns.speed=0.001
    ns.go(0.001)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
