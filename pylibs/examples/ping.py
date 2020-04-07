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

logging.basicConfig(format='%(asctime)s - %(levelname)s - %(message)s', level=logging.DEBUG)

ns = OTNS()
ns.web()

RADIO_RANGE = 460

ns.speed = 4
ns.demo_legend(title="Legend: Node Types & Links", x=850, y=10)


def add_node(*args, **kwargs):
    return ns.add(*args, **kwargs, radio_range=RADIO_RANGE)


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
    ns.countdown(30, "Ping from node 1 to node 11 ... %v left")
    ping(1, 11, 30)
    c1_rlocs = ns.get_ipaddrs(C1, "rloc")
    if c1_rlocs:
        for i in range(4):
            for id in (6, 7, 8, 9, 16, 17, 18, 19, C2, C3):
                ns.ping(id, c1_rlocs[0])

    ns.delete(C1)
    ns.countdown(30, "Switch routing path after removing router 21 ... %v left")
    ping(1, 11, 30)
    ns.delete(C2)
    ns.delete(C3)
    ns.countdown(130, "Waiting for network to form 2 partitions ... %v left")
    ns.go(130)

    add_node("router", 950, 300, id=C1)
    add_node("router", 800, 700, id=C2)
    add_node("router", 1100, 700, id=C3)
    ns.countdown(10, "Restore 3 routers, wait for network to stabilize - %v left")
    ns.go(10)
