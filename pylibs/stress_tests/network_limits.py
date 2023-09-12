#!/usr/bin/env python3
#
# Copyright (c) 2022-2023, The OTNS Authors.
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
import logging
import math
import random

from BaseStressTest import BaseStressTest

PARENT_X = 500
PARENT_Y = 500
MAX_DISTANCE = 250
CHILDREN_N = 10


class StressTest(BaseStressTest):
    SUITE = 'network-limits'

    TIME_LIMIT = {
        'fed': 1,
        'med': 1,
        'sed': 1,
    }

    def __init__(self):
        super(StressTest, self).__init__("Parent with max Children count",
                                         [])

    def run(self):
        self.ns.speed = 30 # speed is lowered to see the visualization, when run locally.
        self.test('fed')
        self.test('med')
        self.test('sed')
        self.test('med')
        self.test('fed')

    def test(self, child_type: str):
        self.reset()
        self.ns.add("router", PARENT_X, PARENT_Y)
        self.ns.go(7)

        time_limit = StressTest.TIME_LIMIT[child_type]
        all_children = []

        for i in range(CHILDREN_N):
            angle = math.pi * 2 * i / CHILDREN_N
            d = random.randint(0, MAX_DISTANCE * MAX_DISTANCE) ** 0.5
            child_x = int(PARENT_X + d * math.cos(angle))
            child_y = int(PARENT_Y + d * math.sin(angle))
            child = self.ns.add(child_type, child_x, child_y)
            all_children.append(child)
            self.ns.go(random.uniform(0.001, 0.1))

        for i in range(time_limit):
            self.ns.go(60)
            n_children=0
            for child in all_children:
                if self.ns.get_state(child) == 'child':
                    n_children += 1
            if n_children == CHILDREN_N:
                logging.info("All %s children has attached successfully within %d minutes.", child_type, i + 1)
                break

        self.ns.speed = 0.01
        self.ns.go(0.01) # trick to ensure final topology is briefly shown in web UI

        if n_children < CHILDREN_N:
            raise Exception("Not all %s children attached within time limit of %d minutes." % (child_type, time_limit))

if __name__ == '__main__':
    StressTest().run()
