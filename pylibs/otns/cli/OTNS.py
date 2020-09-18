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
import os
import shutil
import signal
import subprocess
from typing import List, Union, Optional, Tuple, Dict, Any, Collection

from .errors import OTNSCliError, OTNSExitedError


class OTNS(object):
    """
    OTNS creates and manages an OTNS simulation through CLI.
    """

    MAX_SIMULATE_SPEED = 1000000  # Max simulating speed
    PAUSE_SIMULATE_SPEED = 0

    class GoResult:
        def __init__(self):
            self.packets = []

        def add_packet(self, timestamp: int, nodeid: int, frame: bytes):
            self.packets.append((timestamp, nodeid, frame))

    def __init__(self, otns_path: Optional[str] = None, otns_args: Optional[List[str]] = None):
        self._otns_path = otns_path or self._detect_otns_path()
        self._otns_args = list(otns_args or []) + ['-autogo=false', '-web=false']
        logging.info("otns found: %s", self._otns_path)
        self._launch_otns()
        self._closed = False

    def _launch_otns(self) -> None:
        logging.info("launching otns: %s %s", self._otns_path, ' '.join(self._otns_args))
        self._otns = subprocess.Popen([self._otns_path] + self._otns_args,
                                      bufsize=16384,
                                      stdin=subprocess.PIPE,
                                      stdout=subprocess.PIPE)
        logging.info("otns process launched: %s", self._otns)

    def close(self) -> None:
        """
        Close OTNS simulation.

        :param timeout: timeout for waiting otns process to quit
        """
        if self._closed:
            return

        self._closed = True
        logging.info("waiting for OTNS to close ...")
        self._otns.send_signal(signal.SIGTERM)
        try:
            self._otns.__exit__(None, None, None)
        except BrokenPipeError:
            pass

    def go(self, duration: float = None, speed: float = None, nodeid=None) -> GoResult:
        """
        Continue the simulation for a period of time.

        :param duration: the time duration (in simulating time) for the simulation to continue,
                         or continue forever if duration is not specified.
        :param speed: simulating speed. Use current simulating speed if not specified.
        """
        if duration is None:
            cmd = 'go ever'
        else:
            cmd = f'go {duration}'

        if speed is not None:
            cmd += f' speed {speed}'

        if nodeid is not None:
            cmd += f' node {nodeid}'

        res = OTNS.GoResult()
        lines = self._do_command(cmd)
        for line in lines:
            if line.startswith('DUMP:PACKET:'):
                line = line.strip().split(':')
                timestamp = int(line[2])
                nodeid = int(line[3])
                frame = str(line[4])
                assert len(frame) % 2 == 0
                frame = bytes([int(frame[i:i + 2], 16) for i in range(0, len(frame), 2)])
                res.add_packet(timestamp, nodeid, frame)

        return res

    @property
    def now(self) -> float:
        return self._expect_int(self._do_command('now')) / 1000000.0

    @property
    def speed(self) -> float:
        """
        :return: simulating speed
        """
        speed = self._expect_float(self._do_command(f'speed'))
        if speed >= OTNS.MAX_SIMULATE_SPEED:
            return OTNS.MAX_SIMULATE_SPEED  # max speed
        elif speed <= OTNS.PAUSE_SIMULATE_SPEED:
            return OTNS.PAUSE_SIMULATE_SPEED  # paused
        else:
            return speed

    @speed.setter
    def speed(self, speed: float) -> None:
        """
        Set simulating speed.

        :param speed: new simulating speed
        """
        if speed >= OTNS.MAX_SIMULATE_SPEED:
            speed = OTNS.MAX_SIMULATE_SPEED
        elif speed <= 0:
            speed = OTNS.PAUSE_SIMULATE_SPEED

        self._do_command(f'speed {speed}')

    @staticmethod
    def _detect_otns_path() -> str:
        env_otns_path = os.getenv('OTNS')
        if env_otns_path:
            return env_otns_path

        which_otns = shutil.which('otns')
        if not which_otns:
            raise RuntimeError("otns not found in current directory and PATH")

        return which_otns

    def _do_command(self, cmd: str) -> List[str]:
        logging.info("OTNS <<< %s", cmd)
        try:
            self._otns.stdin.write(cmd.encode('ascii') + b'\n')
            self._otns.stdin.flush()
        except BrokenPipeError:
            self._on_otns_eof()

        output = []
        while True:
            line = self._otns.stdout.readline()
            if line == b'':
                self._on_otns_eof()

            line = line.strip().decode('utf-8')
            logging.info(f"OTNS >>> {line}")
            if line == 'Done':
                return output
            elif line.startswith('Error: '):
                raise OTNSCliError(line[7:])

            output.append(line)

    def add(self, type: str, x: float = None, y: float = None, id=None, radio_range=None, executable=None,
            restore=False, uart_type='auto') -> int:
        """
        Add a new node to the simulation.

        :param type: node type
        :param x: node position X
        :param y: node position Y
        :param id: node ID, or None to use next available node ID
        :param radio_range: node radio range or None for default
        :param executable: specify the executable for the new node, or use default executable if None
        :param restore: whether the node restores network configuration from persistent storage

        :return: added node ID
        """
        cmd = f'add {type}'
        if x is not None:
            cmd = cmd + f' x {x}'
        if y is not None:
            cmd = cmd + f' y {y}'

        if id is not None:
            cmd += f' id {id}'

        if radio_range is not None:
            cmd += f' rr {radio_range}'

        if executable:
            cmd += f' exe "{executable}"'

        if restore:
            cmd += f' restore'

        assert uart_type in ('auto', 'virtual', 'real')
        if uart_type in ['virtual', 'real']:
            cmd += f' uart {uart_type}'

        return self._expect_int(self._do_command(cmd))

    def delete(self, *nodeids: int) -> None:
        """
        Delete nodes from simulation by IDs.

        :param nodeids: node IDs
        """
        cmd = f'del {" ".join(map(str, nodeids))}'
        self._do_command(cmd)

    def move(self, nodeid: int, x: int, y: int) -> None:
        """
        Move node to the target position.

        :param nodeid: target node ID
        :param x: target position X
        :param y: target position Y
        """
        cmd = f'move {nodeid} {x} {y}'
        self._do_command(cmd)

    def ping(self, srcid: int, dst: Union[int, str], addrtype='any', datasize=0) -> None:
        """
        Ping from source node to destination node.

        :param srcid: source node ID
        :param dst: destination node ID or address
        :param addrtype: address type for the destination node (only useful for destination node ID)
        :param datasize: ping data size

        Use pings() to get ping results.
        """
        if isinstance(dst, str):
            addrtype = ''  # addrtype only appliable for dst ID

        cmd = f'ping {srcid} {dst!r} {addrtype} datasize {datasize}'
        self._do_command(cmd)

    @property
    def packet_loss_ratio(self) -> float:
        """
        Get the message drop rate of 128 byte packet.
        Smaller packet has lower drop rate.

        :return: message drop rate (0 ~ 1.0)
        """
        return self._expect_float(self._do_command('plr'))

    @packet_loss_ratio.setter
    def packet_loss_ratio(self, value: float) -> None:
        """
        Set the message drop rate of 128 byte packet.
        Smaller packet has lower drop rate.

        :param value: message drop ratio (0 ~ 1.0)
        """
        self._do_command(f'plr {value}')

    def nodes(self) -> Dict[int, Dict[str, Any]]:
        """
        Get all nodes in simulation

        :return: dict with node IDs as keys and node information as values
        """
        cmd = 'nodes'
        output = self._do_command(cmd)
        nodes = {}
        for line in output:
            nodeinfo = {}
            for kv in line.split():
                k, v = kv.split('=')
                if k in ('id', 'x', 'y'):
                    v = int(v)
                elif k in ('extaddr', 'rloc16'):
                    v = int(v, 16)
                elif k in ('failed',):
                    v = v == 'true'
                elif k in ('ct_interval', 'ct_delay'):
                    v = float(v)
                else:
                    pass

                nodeinfo[k] = v

            nodes[nodeinfo['id']] = nodeinfo

        return nodes

    def partitions(self) -> Dict[int, Collection[int]]:
        """
        Get partitions.

        :return: dict with partition IDs as keys and node list as values
        """
        output = self._do_command('partitions')
        partitions = {}
        for line in output:
            line = line.split()
            assert line[0].startswith('partition=') and line[1].startswith('nodes='), line
            parid = int(line[0].split('=')[1], 16)
            nodeids = list(map(int, line[1].split('=')[1].split(',')))
            partitions[parid] = nodeids

        return partitions

    def radio_on(self, *nodeids: int) -> None:
        """
        Turn on node radio.

        :param nodeids: operating node IDs
        """
        self._do_command(f'radio {" ".join(map(str, nodeids))} on')

    def radio_off(self, *nodeids: int) -> None:
        """
        Turn off node radio.

        :param nodeids: operating node IDs
        """
        self._do_command(f'radio {" ".join(map(str, nodeids))} off')

    def radio_set_fail_time(self, *nodeids: int, fail_time: Optional[Tuple[int, int]]) -> None:
        """
        Set node radio fail time parameters.

        :param nodeids: node IDs
        :param fail_time: fail time (fail_duration, fail_interval) or None for always on.
        """
        fail_duration, period_time = fail_time
        cmd = f'radio {" ".join(map(str, nodeids))} ft {fail_duration} {period_time}'
        self._do_command(cmd)

    def pings(self) -> List[Tuple[int, str, int, float]]:
        """
        Get ping results.

        :return: list of ping results, each of format (node ID, destination address, data size, delay)
        """
        output = self._do_command('pings')
        pings = []
        for line in output:
            line = line.split()
            pings.append((
                int(line[0].split('=')[1]),
                line[1].split('=')[1],
                int(line[2].split('=')[1]),
                float(line[3].split('=')[1][:-2]),
            ))

        return pings

    def joins(self) -> List[Tuple[int, float, float]]:
        """
        Get join results.

        :return: list of join results, each of format (node ID, join time, session time)
        """
        output = self._do_command('joins')
        joins = []
        for line in output:
            line = line.split()
            joins.append((
                int(line[0].split('=')[1]),
                float(line[1].split('=')[1][:-1]),
                float(line[2].split('=')[1][:-1]),
            ))

        return joins

    def counters(self) -> Dict[str, int]:
        """
        Get counters.

        :return: dict of all counters
        """
        output = self._do_command('counters')
        counters = {}
        for line in output:
            name, val = line.split()
            val = int(val)
            counters[name] = val

        return counters

    def prefix_add(self, nodeid: int, prefix: str, preferred=True, slaac=True, dhcp=False, dhcp_other=False,
                   default_route=True, on_mesh=True, stable=True, prf='med') -> None:
        flags = ''
        if preferred:
            flags += 'p'
        if slaac:
            flags += 'a'
        if dhcp:
            flags += 'd'
        if dhcp_other:
            flags += 'c'
        if default_route:
            flags += 'r'
        if on_mesh:
            flags += 'o'
        if stable:
            flags += 's'

        assert flags
        assert prf in ('high', 'med', 'low')

        cmd = f'prefix add {prefix} {flags} {prf}'
        self.node_cmd(nodeid, cmd)
        self.node_cmd(nodeid, 'netdataregister')

    def node_cmd(self, nodeid: int, cmd: str) -> List[str]:
        """
        Run command on node.

        :param nodeid: target node ID
        :param cmd: command to execute

        :return: lines of command output
        """
        cmd = f'node {nodeid} "{cmd}"'
        output = self._do_command(cmd)
        return output

    def get_state(self, nodeid: int) -> str:
        """
        Get node state.

        :param nodeid: node ID
        """
        output = self.node_cmd(nodeid, "state")
        return self._expect_str(output)

    def get_rloc16(self, nodeid: int) -> int:
        """
        Get node RLOC16.

        :param nodeid: node ID
        :return: node RLOC16
        """
        return self._expect_hex(self.node_cmd(nodeid, "rloc16"))

    def get_ipaddrs(self, nodeid: int, addrtype: str = None) -> List[str]:
        """
        Get node ipaddrs.

        :param nodeid: node ID
        :param addrtype: address type (e.x. mleid, rloc, linklocal), or None for all addresses

        :return: list of filtered addresses
        """
        cmd = "ipaddr"
        if addrtype:
            cmd += f' {addrtype}'

        return self.node_cmd(nodeid, cmd)

    def set_network_name(self, nodeid: int, name: str = None) -> None:
        """
        Set network name.

        :param nodeid: node ID
        :param name: network name to set
        """
        name = self._escape_whitespace(name)
        self.node_cmd(nodeid, f'networkname {name}')

    def get_network_name(self, nodeid: int) -> str:
        """
        Get network name.

        :param nodeid: node ID

        :return: network name
        """
        return self._expect_str(self.node_cmd(nodeid, 'networkname'))

    def set_panid(self, nodeid: int, panid: int) -> None:
        """
        Set node pan ID.

        :param nodeid: node ID
        :param panid: pan ID
        """
        self.node_cmd(nodeid, 'panid 0x%04x' % panid)

    def get_panid(self, nodeid: int) -> int:
        """
        Get node pan ID.

        :param nodeid: node ID

        :return: pan ID
        """
        return self._expect_hex(self.node_cmd(nodeid, 'panid'))

    def get_masterkey(self, nodeid: int) -> str:
        """
        Get master key.

        :param nodeid: target node ID

        :return: master key as a hex string
        """
        return self._expect_str(self.node_cmd(nodeid, 'masterkey'))

    def set_masterkey(self, nodeid: int, key: str) -> None:
        """
        Set master key.

        :param nodeid: target node ID
        :param key: master key as a hex string
        """
        self.node_cmd(nodeid, f'masterkey {key}')

    def web(self) -> None:
        """
        Open web browser for visualization.
        """
        self._do_command('web')

    def ifconfig_up(self, nodeid: int) -> None:
        """
        Turn up network interface.

        :param nodeid: target node ID
        """
        self.node_cmd(nodeid, 'ifconfig up')

    def ifconfig_down(self, nodeid: int) -> None:
        """
        Turn down network interface.

        :param nodeid: target node ID
        """
        self.node_cmd(nodeid, 'ifconfig down')

    def thread_start(self, nodeid: int) -> None:
        """
        Start thread.

        :param nodeid: target node ID
        """
        self.node_cmd(nodeid, 'thread start')

    def thread_stop(self, nodeid: int) -> None:
        """
        Stop thread.

        :param nodeid: target node ID
        """
        self.node_cmd(nodeid, 'thread stop')

    def commissioner_start(self, nodeid: int) -> None:
        """
        Start commissioner.

        :param nodeid: target node ID
        """
        self.node_cmd(nodeid, "commissioner start")

    def joiner_start(self, nodeid: int, pwd: str) -> None:
        """
        Start joiner.

        :param nodeid: joiner node ID
        :param pwd: commissioning password
        """
        self.node_cmd(nodeid, f"joiner start {pwd}")

    def commissioner_joiner_add(self, nodeid: int, usr: str, pwd: str, timeout=None) -> None:
        """
        Add joiner to commissioner.

        :param nodeid: commissioner node ID
        :param usr: commissioning user
        :param pwd: commissioning password
        :param timeout: commissioning session timeout
        """
        timeout_s = f" {timeout}" if timeout is not None else ""
        self.node_cmd(nodeid, f"commissioner joiner add {usr} {pwd}{timeout_s}")

    def config_visualization(self, broadcast_message: bool = None, unicast_message: bool = None,
                             ack_message: bool = None, router_table: bool = None, child_table: bool = None) \
            -> Dict[str, bool]:
        """
        Configure the visualization options.

        :param broadcast_message: whether or not to visualize broadcast messages
        :param unicast_message: whether or not to visualize unicast messages
        :param ack_message: whether or not to visualize ACK messages
        :param router_table: whether or not to visualize router tables
        :param child_table: whether or not to visualize child tables

        :return: the active visualization options
        """
        cmd = "cv"
        if broadcast_message is not None:
            cmd += " bro " + ("on" if broadcast_message else "off")

        if unicast_message is not None:
            cmd += " uni " + ("on" if unicast_message else "off")

        if ack_message is not None:
            cmd += " ack " + ("on" if ack_message else "off")

        if router_table is not None:
            cmd += " rtb " + ("on" if router_table else "off")

        if child_table is not None:
            cmd += " ctb " + ("on" if child_table else "off")

        output = self._do_command(cmd)
        vopts = {}
        for line in output:
            line = line.split('=')
            assert len(line) == 2 and line[1] in ('on', 'off'), line
            vopts[line[0]] = (line[1] == "on")

        # convert command options to python options
        vopts['broadcast_message'] = vopts.pop('bro')
        vopts['unicast_message'] = vopts.pop('uni')
        vopts['ack_message'] = vopts.pop('ack')
        vopts['router_table'] = vopts.pop('rtb')
        vopts['child_table'] = vopts.pop('ctb')

        return vopts

    def set_title(self, title: str, x: int = None, y: int = None, font_size: int = None) -> None:
        """
        Set simulation title.

        :param title: title text
        :param x: X coordinate of title
        :param y: Y coordinate of title
        :param font_size: Font size of title
        """
        cmd = f'title "{title}"'

        if x is not None:
            cmd += f' x {x}'

        if y is not None:
            cmd += f' y {y}'

        if font_size is not None:
            cmd += f' fs {font_size}'

        self._do_command(cmd)

    def set_network_info(self, version: str = None, commit: str = None, real: bool = None) -> None:
        """
        Set network info.

        :param version: The OpenThread version.
        :param commit: The OpenThread commit.
        :param real: If the network uses real devices.
        """
        cmd = 'netinfo'

        if version is not None:
            cmd += f' version "{version}"'

        if commit is not None:
            cmd += f' commit "{commit}"'

        if real is not None:
            cmd += f' real {1 if real else 0}'

        self._do_command(cmd)

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def get_router_upgrade_threshold(self, nodeid: int) -> int:
        """
        Get Router upgrade threshold.
        :param nodeid: the node ID
        :return: the Router upgrade threshold
        """
        return self._expect_int(self.node_cmd(nodeid, 'routerupgradethreshold'))

    def set_router_upgrade_threshold(self, nodeid: int, val: int) -> None:
        """
        Set Router upgrade threshold.
        :param nodeid: the node ID
        :param val: the Router upgrade threshold
        """
        self.node_cmd(nodeid, f'routerupgradethreshold {val}')

    def get_router_downgrade_threshold(self, nodeid: int) -> int:
        """
        Get Router downgrade threshold.
        :param nodeid: the node ID
        :return: the Router downgrade threshold
        """
        return self._expect_int(self.node_cmd(nodeid, 'routerdowngradethreshold'))

    def set_router_downgrade_threshold(self, nodeid: int, val: int) -> None:
        """
        Set Router downgrade threshold.
        :param nodeid: the node ID
        :param val: the Router downgrade threshold
        """
        self.node_cmd(nodeid, f'routerdowngradethreshold {val}')

    @staticmethod
    def _expect_int(output: List[str]) -> int:
        assert len(output) == 1, output
        return int(output[0])

    @staticmethod
    def _expect_hex(output: List[str]) -> int:
        assert len(output) == 1, output
        return int(output[0], 16)

    @staticmethod
    def _expect_float(output: List[str]) -> float:
        assert len(output) == 1, output
        return float(output[0])

    @staticmethod
    def _expect_str(output: List[str]) -> str:
        assert len(output) == 1, output
        return output[0].strip()

    @staticmethod
    def _escape_whitespace(s: str) -> str:
        """
        Escape string by replace <whitespace> by \\<whitespace>.

        :param s: string to escape

        :return: the escaped string
        """
        for c in "\\ \t\r\n":
            s = s.replace(c, '\\' + c)
        return s

    def _on_otns_eof(self):
        exit_code = self._otns.wait()
        logging.warning("otns exited: code = %d", exit_code)
        raise OTNSExitedError(exit_code)


if __name__ == '__main__':
    import doctest

    doctest.testmod()
