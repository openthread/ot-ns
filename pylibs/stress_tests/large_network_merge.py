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
import time

from BaseStressTest import BaseStressTest
from config import DEBUG

XGAP = 100
YGAP = 100
RADIO_RANGE = int(XGAP * 1.5)

LARGE_N = 8

SIMULATE_TIME = 60
REPEAT = 1 if DEBUG else 100


class StressTest(BaseStressTest):
    SUITE = 'robustness'

    def __init__(self):
        super(StressTest, self).__init__("Large Network Formation Test",
                                         ["Simulation Time", "Execution Time", "Average Partition Count in 60s"])

    def run(self):
        self.ns.packet_loss_ratio = 0.2

        durations = []
        partition_counts = []
        for _ in range(REPEAT):
            dt, par_cnt = self.test_n(LARGE_N)
            durations.append(dt)
            partition_counts.append(par_cnt)

        self.result.append_row('%ds' % (SIMULATE_TIME * REPEAT), '%ds' % sum(durations),
                               '%d' % (sum(partition_counts) / len(partition_counts)))

    def test_n(self, n):
        self.reset()

        for r in range(n):
            for c in range(n):
                self.ns.add("router", 50 + XGAP * c, 50 + YGAP * r, radio_range=RADIO_RANGE)

        t0 = time.time()
        self.ns.go(SIMULATE_TIME)
        dt = time.time() - t0
        return dt, len(self.ns.partitions())


if __name__ == '__main__':
    StressTest().run()
