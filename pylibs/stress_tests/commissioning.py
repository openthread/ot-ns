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

import os
import random
import time
from typing import List, Dict, Tuple

from BaseStressTest import BaseStressTest
from otns.cli.errors import OTNSCliError

XGAP = 100
YGAP = 100

PASSWORD = "TEST123"

REPEAT = int(os.getenv("STRESS_LEVEL", 1)) * 10
N = 5


class CommissioningStressTest(BaseStressTest):
    SUITE = 'commissioning'

    def __init__(self):
        super(CommissioningStressTest, self).__init__("Commissioning Test",
                                                      ["Join Count", "Success Percent", "Average Join Time"],
                                                      raw=True)
        self._join_time_accum = 0
        self._join_count = 0
        self._join_fail_count = 0
        self.ns.packet_loss_ratio = 0.05

    def run(self):
        for _ in range(REPEAT):
            self._press_commissioning(N, N, max_joining_count=2)

        expected_join_count = REPEAT * (N * N - 1)
        total_join_count = self._join_count + self._join_fail_count
        self.result.fail_if(self._join_count < expected_join_count,
                            "Join Count (%d) < %d" % (self._join_count, expected_join_count))
        join_ok_percent = self._join_count * 100 // total_join_count
        avg_join_time = self._join_time_accum / self._join_count if self._join_count else float('inf')
        self.result.append_row(total_join_count, '%d%%' % join_ok_percent, '%.0fs' % avg_join_time)
        self.result.fail_if(join_ok_percent < 90, "Success Percent (%d%%) < 90%%" % join_ok_percent)
        self.result.fail_if(avg_join_time > 20, "Average Join Time (%.0f) > 20s" % avg_join_time)

    def _press_commissioning(self, R: int, C: int, max_joining_count: int = 2):
        self.reset()

        ns = self.ns

        G: List[List[int]] = [[-1] * C for _ in range(R)]
        RC: Dict[int, Tuple[int, int]] = {}

        for r in range(R):
            for c in range(C):
                device_role = 'router'
                if R >= 3 and C >= 3 and (r in (0, R - 1) or c in (0, C - 1)):
                    device_role = random.choice(['fed', 'med', 'sed'])
                G[r][c] = ns.add(device_role, x=c * XGAP + XGAP, y=YGAP + r * YGAP)
                RC[G[r][c]] = (r, c)

        joined: List[List[bool]] = [[False] * C for _ in range(R)]
        started: List[List[bool]] = [[False] * C for _ in range(R)]

        # choose and setup the commissioner
        cr, cc = R // 2, C // 2
        ns.node_cmd(G[cr][cc], 'dataset init new')
        ns.node_cmd(G[cr][cc], 'dataset')
        ns.node_cmd(G[cr][cc], 'dataset networkkey 00112233445566778899aabbccddeeff')
        ns.node_cmd(G[cr][cc], 'dataset commit active')
        ns.ifconfig_up(G[cr][cc])
        ns.thread_start(G[cr][cc])
        ns.go(15)
        assert ns.get_state(G[cr][cc]) == 'leader'
        started[cr][cc] = joined[cr][cc] = True

        # bring up all nodes
        for r in range(R):
            for c in range(C):
                ns.ifconfig_up(G[r][c])

        join_order = [(r, c) for r in range(R) for c in range(C)]
        join_order = sorted(join_order, key=lambda rc: abs(rc[0] - cr) + abs(rc[1] - cc))

        joining = {}
        now = 0

        commissioner_session_start_time = 0
        deadline = R * C * 100

        while now < deadline and not all(started[r][c] for r in range(R) for c in range(C)):
            if commissioner_session_start_time == 0 or commissioner_session_start_time + 1000 <= now:
                try:
                    ns.commissioner_start(G[cr][cc])
                except OTNSCliError as ex:
                    if str(ex).endswith('Already'):
                        pass

                ns.commissioner_joiner_add(G[cr][cc], "*", PASSWORD, 1000)
                commissioner_session_start_time = now

            # start all joined but not started nodes
            for r, c in join_order:
                if joined[r][c] and not started[r][c]:
                    ns.thread_start(G[r][c])
                    started[r][c] = True

            # choose `max_joining_count` nodes to join
            for r, c in join_order:
                if len(joining) >= max_joining_count:
                    break

                if joined[r][c]:
                    continue

                if (r, c) in joining:
                    continue

                ns.joiner_start(G[r][c], PASSWORD)
                joining[r, c] = 0

                ns.go(0.1)  # small delay to avoid lockstep sync'ed behavior of new started joiner nodes.
                now += 0.1

            # make sure the joining nodes are joining
            for (r, c), ts in joining.items():
                if ts == 0 or ts + 10 < now:
                    try:
                        ns.joiner_start(G[r][c], PASSWORD)
                    except OTNSCliError as ex:
                        if str(ex).endswith("Busy"):
                            pass

                    joining[(r, c)] = now

            ns.go(1)
            now += 1

            joins = ns.joins()
            for nodeid, join_time, session_time in joins:
                if join_time > 0:
                    r, c = RC[nodeid]
                    joining.pop((r, c), None)
                    joined[r][c] = True

                    self._join_count += 1
                    self._join_time_accum += join_time
                else:
                    self._join_fail_count += 1

        # (typically) all nodes are now joined and started. Simulate some time to see the full partition form.
        ns.go(200)
        time.sleep(2)  # in GUI case, allow display to catch up before exiting.


if __name__ == '__main__':
    CommissioningStressTest().run()
