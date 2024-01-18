#!/usr/bin/env python3
# Copyright (c) 2020, The OTNS Authors.
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

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

RADIO_RANGE = 460


def main():
    logging.basicConfig(format='%(asctime)s - %(levelname)s - %(message)s', level=logging.DEBUG)

    ns = OTNS()
    ns.set_title("Ping Example")
    ns.web()
    ns.radiomodel = 'Ideal'

    ns.speed = 6

    def add_node(*args, **kwargs):
        nid = ns.add(*args, **kwargs, radio_range=RADIO_RANGE)
        ns.node_cmd(nid, 'routerselectionjitter 1')
        return nid

    add_node("fed", 100, 100)
    add_node("fed", 100, 300)
    add_node("fed", 100, 500)
    add_node("fed", 100, 700)
    add_node("fed", 100, 900)

    add_node("router", 450, 100)
    add_node("router", 550, 300)
    add_node("router", 450, 500)
    add_node("router", 550, 700)
    add_node("router", 450, 900)

    add_node("fed", 1800, 100)
    add_node("fed", 1800, 300)
    add_node("fed", 1800, 500)
    add_node("fed", 1800, 700)
    add_node("fed", 1800, 900)

    add_node("router", 1450, 100)
    add_node("router", 1350, 300)
    add_node("router", 1450, 500)
    add_node("router", 1350, 700)
    add_node("router", 1450, 900)

    C1 = add_node("router", 950, 300)
    C2 = add_node("router", 800, 700)
    C3 = add_node("router", 1100, 700)

    def ping(src: int, dst: int, duration: float):
        while duration > 0:
            ns.ping(src, dst)
            ns.go(1)
            duration -= 1

    while True:
        ping(1, 11, 30)
        c1_rlocs = ns.get_ipaddrs(C1, "rloc")
        if c1_rlocs:
            for i in range(4):
                for id in (6, 7, 8, 9, 16, 17, 18, 19, C2, C3):
                    ns.ping(id, c1_rlocs[0])

        ns.delete(C1)
        ping(1, 11, 30)
        ns.delete(C2)
        ns.delete(C3)
        ns.go(130)

        add_node("router", 950, 300, id=C1)
        add_node("router", 800, 700, id=C2)
        add_node("router", 1100, 700, id=C3)
        ns.go(10)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
