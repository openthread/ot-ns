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
import math
import os

from otns.cli.errors import OTNSCliError
from BaseStressTest import BaseStressTest

R = 300
RADIO_RANGE = 200
INNER_JOINER_NUM = 4
OUTTER_JOINER_NUM = 10
CENTER_X, CENTER_Y = 500, 500

PASSWORD = "TEST123"

REPEAT = int(os.getenv("STRESS_LEVEL", 1)) * 5


class CommissioningManyStressTest(BaseStressTest):
    SUITE = 'commissioning'

    def __init__(self):
        super(CommissioningManyStressTest, self).__init__("Commissioning Many Test",
                                                          ["Join Count", "Success Percent", "Average Join Time"],
                                                          raw=True)
        self._join_time_accum = 0
        self._join_count = 0
        self._join_fail_count = 0
        self.ns.packet_loss_ratio = 0.3
        # self.ns.config_visualization(broadcast_message=False)

    def run(self):
        for i in range(REPEAT):
            self._press_commissioning_many(i + 1)

        expected_join_count = 100
        total_join_count = self._join_count + self._join_fail_count
        self.result.fail_if(self._join_count != expected_join_count,
                            "Join Count (%d) != %d" % (self._join_count, expected_join_count))
        join_ok_percent = self._join_count * 100 // total_join_count
        avg_join_time = self._join_time_accum / self._join_count if self._join_count else float('inf')
        self.result.append_row(total_join_count, '%d%%' % join_ok_percent,
                               '%.0fs' % avg_join_time)
        self.result.fail_if(join_ok_percent < 90, "Success Percent (%d%%) < 50%%" % join_ok_percent)
        self.result.fail_if(avg_join_time > 20, "Average Join Time (%.0f) > 60s" % avg_join_time)

    def _press_commissioning_many(self, repeat):
        self.reset()
        ns = self.ns

        ns.set_title(f"Commissioning Test Repeat {repeat}")

        # Create the Leader
        commissioner_id = ns.add("router", x=CENTER_X, y=CENTER_Y, radio_range=RADIO_RANGE)
        joiner_ids = []

        for i in range(INNER_JOINER_NUM):
            angle = math.pi * 2 * i / INNER_JOINER_NUM
            joiner_id = ns.add("router", x=CENTER_X + math.cos(angle) * R / 2, y=CENTER_Y + math.sin(angle) * R / 2,
                               radio_range=RADIO_RANGE)
            joiner_ids.append(joiner_id)

        for i in range(OUTTER_JOINER_NUM):
            angle = math.pi * 2 * i / OUTTER_JOINER_NUM
            joiner_id = ns.add("router", x=CENTER_X + math.cos(angle) * R, y=CENTER_Y + math.sin(angle) * R,
                               radio_range=RADIO_RANGE)
            joiner_ids.append(joiner_id)

        # Inject radio failures
        ns.radio_set_fail_time(commissioner_id, fail_time=(120, 600))
        ns.radio_set_fail_time(*joiner_ids, fail_time=(120, 600))

        # Bring up the Commissioner
        ns.node_cmd(commissioner_id, 'dataset init new')
        ns.node_cmd(commissioner_id, 'dataset')
        ns.node_cmd(commissioner_id, 'dataset masterkey 00112233445566778899aabbccddeeff')
        ns.node_cmd(commissioner_id, 'dataset commit active')
        ns.ifconfig_up(commissioner_id)
        ns.thread_start(commissioner_id)
        ns.go(10)
        assert ns.get_state(commissioner_id) == 'leader'

        # Bring up the Joiners
        for joiner_id in joiner_ids:
            ns.ifconfig_up(joiner_id)

        commissioner_session_start_time = 0
        deadline = 7200
        now = 0
        joined = {id: False for id in joiner_ids}
        finished = {id: False for id in joiner_ids}

        while now < deadline and not all(finished.values()):

            # Start all joiners that haven't joined
            for joiner_id in joiner_ids:
                if not joined[joiner_id]:
                    try:
                        ns.joiner_start(joiner_id, PASSWORD)
                    except OTNSCliError as ex:
                        if str(ex).endswith('Busy'):
                            pass

            # Restart Commissioner session every 1000s
            if commissioner_session_start_time == 0 or commissioner_session_start_time + 1000 <= now:
                try:
                    ns.commissioner_start(commissioner_id)
                    ns.commissioner_joiner_add(commissioner_id, "*", PASSWORD, 1000)
                except OTNSCliError as ex:
                    if str(ex).endswith('Already') or str(ex).endswith("InvalidState"):
                        pass

                commissioner_session_start_time = now

            # Start all joined but not started nodes
            for joiner_id, j in joined.items():
                if j and not finished[joiner_id]:
                    ns.thread_start(joiner_id)
                    finished[joiner_id] = True

            ns.go(30)

            joins = ns.joins()
            for nodeid, join_time, session_time in joins:
                if join_time > 0:
                    joined[nodeid] = True

                    self._join_count += 1
                    self._join_time_accum += (now + join_time)
                else:
                    self._join_fail_count += 1

            now += 30

        logging.warning("All Joiners joined successfully in %d seconds", now)


if __name__ == '__main__':
    CommissioningManyStressTest().run()
