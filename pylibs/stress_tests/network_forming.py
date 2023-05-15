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
# Network Forming Stress Test:
#   Different number of nodes form networks (a single partition) and measure the network forming delay.
# Topology:
#   1x1 Routers ~ 6x6 Routers
# Fault Injections:
#   None
# Pass Criteria:
#   Network forming time is less than corresponding time limits
#
import os
from typing import Sequence

from BaseStressTest import BaseStressTest

XGAP = 100
YGAP = 100
RADIO_RANGE = int(XGAP * 1.5)

MIN_N = 1
MAX_N = 6

REPEAT = int(os.getenv('STRESS_LEVEL', '1')) * 3

EXPECTED_MERGE_TIME_MAX = [
    None, 10, 14, 30, 60, 130, 190
]


class StressTest(BaseStressTest):
    SUITE = 'network-forming'

    def __init__(self):
        headers = ['Network Size', 'Formation Time 1']
        for i in range(2, REPEAT + 1):
            headers.append(f'FT {i}')

        super(StressTest, self).__init__("Network Formation Test", headers)

    def run(self):
        # self.ns.config_visualization(broadcast_message=False)

        for n in range(MIN_N, MAX_N + 1):
            durations = []
            for i in range(REPEAT):
                secs = self.test_n(n)
                durations.append(secs)

            self.result.append_row(f'{n}x{n}', *['%ds' % d for d in durations])
            avg_dura = self.avg_except_max(durations)
            self.result.fail_if(avg_dura > EXPECTED_MERGE_TIME_MAX[n],
                                f"""{n}x{n} average formation time {avg_dura} > {
                                EXPECTED_MERGE_TIME_MAX[n]}""")

    @staticmethod
    def stdvar(nums: Sequence[float]):
        ex = sum(nums) / len(nums)
        s = 0
        for i in nums:
            s += (i - ex) ** 2
        return float(s) / len(nums)

    def test_n(self, n):
        self.reset()

        for r in range(n):
            for c in range(n):
                self.ns.add("router", 50 + XGAP * c, 50 + YGAP * r, radio_range=RADIO_RANGE)

        secs = 0
        while True:
            self.ns.go(1)
            secs += 1

            pars = self.ns.partitions()
            if len(pars) == 1 and 0 not in pars:
                break

        return secs


if __name__ == '__main__':
    StressTest().run()
