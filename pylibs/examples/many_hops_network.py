#!/usr/bin/env python3
#
# This script simulates a line topology of Routers with a selected number of hops,
# for a range of packet-loss percentages. Log files (.csv) are output that show
# the state of each node over time, to validate that Routers don't lose connectivity
# to the Leader and so will remain Routers.
#
# SIM_MINUTES can be increased to higher (e.g. 90) to get better statistics.
#

import enum
import logging
import math
import os
from typing import Dict

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

SIM_PERIOD = 1
SIM_SPEED = float('inf')
SIM_MINUTES = 12
ROUTER_SPACING = 100
SED_SPACING = 75
ROUTER_COUNT = 11
SED_COUNT = 4

class ThreadState(enum.Enum):
    OFFLINE = 'offline'
    DISABLED = 'disabled'
    DETACHED = 'detached'
    CHILD = 'child'
    ROUTER = 'router'
    LEADER = 'leader'

state_map: Dict[ThreadState, int] = {
    ThreadState.OFFLINE: 0,
    ThreadState.DISABLED: 1,
    ThreadState.DETACHED: 2,
    ThreadState.CHILD: 3,
    ThreadState.ROUTER: 4,
    ThreadState.LEADER: 5
}


def simulate(sim_speed: int = SIM_SPEED,
             packet_loss_ratio: float = 0.0,
             output_file: str = 'node_state.csv',
             key_file: str = 'network_info.txt',
             pcap_file: str = 'current.pcap',
             router_count: int = ROUTER_COUNT,
             sed_count: int = SED_COUNT,
             sim_period: float = SIM_PERIOD,
             sim_minutes: int = SIM_MINUTES,
             title: str = "Test",
             web: bool = True):
    ns = OTNS(otns_args=['-log', 'info', '-watch', 'warn'])
    ns.speed = sim_speed
    ns.radiomodel = 'MIDisc' # Use disc model to enforce line topology of Routers
    ns.packet_loss_ratio = packet_loss_ratio
    ns.set_title(title)
    ns.config_visualization(broadcast_message=False)
    if web:
        ns.web()

    routers = dict()
    seds = dict()

    last_node_added = -1
    time_accum = 0
    epoch = 0
    with open(output_file, 'w') as f:
        while True:
            # Add a new router every 30 seconds of sim until we get enough hops - i
            if last_node_added < 0 or time_accum - last_node_added > 30:
                if len(routers) < router_count:
                    i = len(routers)
                    rx = ROUTER_SPACING * (i + 1)
                    ry = 100
                    name = f'router{i}'
                    router = ns.add("router", rx, ry, radio_range=round(ROUTER_SPACING * 1.8))
                    routers[name] = router
                elif len(seds) < sed_count:
                    i = len(seds)
                    rx = ROUTER_SPACING * router_count
                    ry = 100

                    # Now Some Trig to place SEDs around last Router
                    angle = (180 / (sed_count - 1)) * i
                    rx = rx + round(SED_SPACING * math.sin(math.radians(angle)), 0)
                    ry = ry + round(SED_SPACING * math.cos(math.radians(angle)), 0)

                    name = f'sed{i}'
                    sed = ns.add("sed", int(rx), int(ry), radio_range=round(SED_SPACING * 1.6))
                    seds[name] = sed

                last_node_added = time_accum

            dt = sim_period
            ns.go(dt)
            time_accum += dt
            epoch += 1

            for k, r in routers.items():
                state = ns.get_state(r)
                internal_state = ThreadState(state)
                f.write(f'{epoch},{time_accum},{k},{internal_state},{state_map[internal_state]}\n')

            for k, s in seds.items():
                state = ns.get_state(s)
                internal_state = ThreadState(state)
                f.write(f'{epoch},{time_accum},{k},{internal_state},{state_map[internal_state]}\n')

            f.flush()

            if time_accum >= sim_minutes * 60:
                # Exit After Simulating for the specified duration in minutes
                break

            # Every second, ping the first router
            if len(seds) > 1 and (time_accum % 1 < sim_period):
                idx = math.floor(time_accum % len(seds))
                key = list(seds.keys())[idx]
                source = seds[key]
                target = routers['router0']
                ns.ping(source, target)

    with open(key_file, 'w') as f:
        key = ns.node_cmd(1, 'networkkey')
        f.write(f'key: {key[0]}\n')
        for r in routers.values():
            ips = ns.node_cmd(r, 'ipaddr')
            for ip in ips:
                f.write(f'{r}: {ip}\n')

        for s in seds.values():
            ips = ns.node_cmd(s, 'ipaddr')
            for ip in ips:
                f.write(f'{s}: {ip}\n')

    os.rename('current.pcap', pcap_file)
    ns.close()

def main():
    #logging.getLogger().setLevel(logging.INFO)
    prefix = f'line_topo_{ROUTER_COUNT}_hops_'
    for error_percent in [0,5,10]:
        print(f'Simulating {error_percent}% error')
        file_name = f'{prefix}node_state_{error_percent}-percent.csv'
        key_file = f'{prefix}network_info_{error_percent}-percent.txt'
        pcap_file = f'{prefix}_{error_percent}-percent.pcap'
        simulate(packet_loss_ratio=(error_percent / 100.0),
                 output_file=file_name,
                 key_file=key_file,
                 pcap_file=pcap_file,
                 title=f'Packet Error Rate {error_percent}%')


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
