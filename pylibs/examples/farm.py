#!/usr/bin/env python3
# Copyright (c) 2020-2024, The OTNS Authors.
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

# This script simulates a farm where sensors are installed on horses.
# 6 Routers are installed at the borders of the farm which has a transmission range of 300m.
# One of the Routers is selected as the Gateway.
# A SED sensor is installed on each horse to collect and transmit information to the Gateway.
# The horses moves randomly in the farm, thus SED sensors lose connectivity to their parents constantly
# and reattaches to new parents in range.
# This example shows that SED devices can handle parent loss gracefully to retain connectivity.
# The messages dropped due to parent connectivity loss can also be observed in the simulation.
#
# Inspired by https://www.threadgroup.org/Farm-Jenny

import logging
import math
import random

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

R = 3 # screen-pixels per meter
RECEIVER_TX_POWER = 10 # dBm, integer - router
HORSE_TX_POWER = -20 #dBm, integer - sensor
HORSE_NUM = 10
HORSE_MAX_SPEED_MPS = 3 # horse max speed in m/sec
FARM_RECT = [20 * R, 20 * R, 440 * R, 260 * R] # number in meters


def main():
    ns = OTNS(otns_args=['-no-logfile'])

    if False: # Optional forcing of random-seed for OTNS and Python. This gives exact reproducable simulation.
        # The pcap parameter is to select another PCAP type that includes channel info.
        random_seed = 2142142
        ns = OTNS(otns_args=['-seed', f'{random_seed}', '-pcap', 'wpan-tap'])
        random.seed(random_seed)

    ns.loglevel = 'info'
    ns.watch_default('note')
    ns.logconfig(logging.INFO)
    ns.speed = 4
    ns.radiomodel = 'Outdoor'
    ns.set_radioparam('MeterPerUnit', 1/R )
    ns.set_radioparam('ShadowFadingSigmaDb', 0.0)
    ns.set_radioparam('TimeFadingSigmaMaxDb', 0.0)

    ns.set_title("Farm Example")
    ns.config_visualization(broadcast_message=False)
    ns.web()

    gateway = ns.add("router", FARM_RECT[0], FARM_RECT[1])
    ns.add("router", FARM_RECT[0], FARM_RECT[3], txpower=RECEIVER_TX_POWER)
    ns.add("router", FARM_RECT[2], FARM_RECT[1], txpower=RECEIVER_TX_POWER)
    ns.add("router", FARM_RECT[2], FARM_RECT[3], txpower=RECEIVER_TX_POWER)
    ns.add("router", (FARM_RECT[0] + FARM_RECT[2]) // 2, FARM_RECT[1], txpower=RECEIVER_TX_POWER)
    ns.add("router", (FARM_RECT[0] + FARM_RECT[2]) // 2, FARM_RECT[3], txpower=RECEIVER_TX_POWER)

    horse_pos = {}
    horse_move_dir = {}

    for i in range(HORSE_NUM):
        rx = random.randint(FARM_RECT[0] + 20, FARM_RECT[2] - 20)
        ry = random.randint(FARM_RECT[1] + 20, FARM_RECT[3] - 20)
        sid = ns.add("ssed", rx, ry, txpower=HORSE_TX_POWER)
        horse_pos[sid] = (rx, ry)
        horse_move_dir[sid] = random.uniform(0, math.pi * 2)

    def blocked(sid, x, y):
        if not (FARM_RECT[0] + 20 < x < FARM_RECT[2] - 20) or not (FARM_RECT[1] + 20 < y < FARM_RECT[3] - 20):
            return True

        for oid, (ox, oy) in horse_pos.items():
            if oid == sid:
                continue

            dist2 = (x - ox) ** 2 + (y - oy) ** 2
            if dist2 <= 1600:
                return True

        return False

    time_accum = 0
    sid_last_ping = 0

    while True:
        dt = 1
        ns.go(dt)
        time_accum += dt

        for sid, (sx, sy) in horse_pos.items():

            for i in range(10):
                mdist = random.uniform(0, HORSE_MAX_SPEED_MPS * R * dt)

                sx = int(sx + mdist * math.cos(horse_move_dir[sid]))
                sy = int(sy + mdist * math.sin(horse_move_dir[sid]))

                if blocked(sid, sx, sy):
                    horse_move_dir[sid] += random.uniform(0, math.pi * 2)
                    continue

                sx = min(max(sx, FARM_RECT[0]), FARM_RECT[2])
                sy = min(max(sy, FARM_RECT[1]), FARM_RECT[3])
                ns.move(sid, sx, sy)

                horse_pos[sid] = (sx, sy)
                break

        if time_accum >= HORSE_NUM+1:
            ns.print_pings(ns.pings())
            found = False
            for sid in horse_pos:
                if sid > sid_last_ping:
                    ns.ping(sid, gateway)
                    sid_last_ping = sid
                    found = True
                    break
            if not found:
                sid_last_ping = 0
                time_accum = 0

    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
