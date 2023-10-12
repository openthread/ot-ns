#!/usr/bin/env python3
#
# Copyright (c) 2020-2023, The OTNS Authors.
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
import ipaddress
import os
import sys
import time
import traceback
from functools import wraps
from typing import Collection

from otns.cli import OTNS
from otns.cli.errors import UnexpectedError
from StressTestResult import StressTestResult
from errors import UnexpectedNodeAddr, UnexpectedNodeState


class StressTestMetaclass(type):
    def __new__(cls, name, bases, dct):
        assert 'run' in dct, f'run method is not defined in {name}'

        orig_run = dct.pop('run')

        @wraps(orig_run)
        def run_wrapper(self: 'BaseStressTest', report=True):
            try:
                orig_run(self)
            except Exception as ex:
                traceback.print_exc()
                self.result.fail_with_error(ex)
            finally:
                self.stop()

            if report:
                self.report()

        dct['run'] = run_wrapper

        t = super().__new__(cls, name, bases, dct)
        return t


class BaseStressTest(object, metaclass=StressTestMetaclass):
    def __init__(self, name, headers, raw=False, web=True):
        self.name = name
        self._otns_args = ['-log','info','-no-logfile'] # use ['-log', 'debug'] for more debug messages
        if raw:
            self._otns_args.append('-raw')
        self.ns = OTNS(otns_args=self._otns_args)
        self.ns.speed = float('inf')
        if web:
            self.ns.web()

        self.result = StressTestResult(name=name, headers=headers)
        self.result.start()

    def run(self):
        raise NotImplementedError()

    def reset(self):
        nodes = self.ns.nodes()
        if nodes:
            self.ns.delete(*nodes.keys())
        self.ns.speed = float('inf')
        self.ns.go(1)

    def stop(self):
        self.result.stop()
        self.ns.close()

    def expect_node_state(self, nid: int, state: str, timeout: float, go_step: int = 1) -> None:
        while timeout > 0:
            self.ns.go(go_step)
            timeout -= go_step

            if self.ns.get_state(nid) == state:
                return

        raise UnexpectedNodeState(nid, state, self.ns.get_state(nid))

    def report(self):
        try:
            STRESS_RESULT_FILE = os.environ['STRESS_RESULT_FILE']
            stress_result_fd = open(STRESS_RESULT_FILE, 'wt')
        except KeyError:
            stress_result_fd = sys.stdout

        try:
            stress_result_fd.write(
                f"""**[OTNS](https://github.com/openthread/ot-ns) Stress Tests Report Generated at {time.strftime(
                    "%m/%d %H:%M:%S")}**\n""")
            stress_result_fd.write(self.result.format())
        finally:
            if stress_result_fd is not sys.stdout:
                stress_result_fd.close()

        if self.result.failed:
            raise RuntimeError("Stress test failed: \n" + "\n".join("\t" + msg for msg in self.result._fail_msgs))

    def avg_except_max(self, vals: Collection[float]) -> float:
        assert len(vals) >= 2
        max_val = max(vals)
        max_idxes = [i for i in range(len(vals)) if vals[i] >= max_val]
        assert max_idxes
        rmidx = max_idxes[0]
        vals[rmidx:rmidx + 1] = []
        return self.avg(vals)

    def avg(self, vals: Collection[float]) -> float:
        assert len(vals) > 0
        return sum(vals) / len(vals)

    def expect_all_nodes_become_routers(self, timeout: int = 1000) -> None:
        all_routers = False

        while timeout > 0 and not all_routers:
            self.ns.go(10)
            timeout -= 10

            nodes = (self.ns.nodes())

            all_routers = True
            print(nodes)
            for nid, info in nodes.items():
                if info['state'] not in ['leader', 'router']:
                    all_routers = False
                    break

            if all_routers:
                break

        if not all_routers:
            raise UnexpectedError("not all nodes are Routers: %s" % self.ns.nodes())

    def expect_node_addr(self, nodeid: int, addr: str, timeout=100):
        addr = ipaddress.IPv6Address(addr)

        found_addr = False
        while timeout > 0:
            if addr in map(ipaddress.IPv6Address, self.ns.get_ipaddrs(nodeid)):
                found_addr = True
                break

            self.ns.go(1)

        if not found_addr:
            raise UnexpectedNodeAddr(f'Address {addr} not found on node {nodeid}')

    def expect_node_mleid(self, nodeid: int, timeout: int):
        while True:
            mleid = self.ns.get_mleid(nodeid)
            if mleid:
                return mleid

            self.ns.go(1)
            timeout -= 1
            if timeout <= 0:
                raise UnexpectedNodeAddr(f'MLEID not found on node {nodeid}')
