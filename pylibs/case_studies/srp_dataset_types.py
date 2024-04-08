#!/usr/bin/env python3
# Copyright (c) 2024, The OTNS Authors.
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

# Case study on SRP dataset types and their priority.

import traceback

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_BR = 7
LEADER_ID = 200

TYPE_ANYCAST = 0x5c
TYPE_UNICAST = 0x5d

SERVICE_DATA_LEN_TYPE1 = 19
SERVICE_DATA_LEN_TYPE3 = 1

def get_services(ns, id, tp, service_data_len = 0):
    lines = ns.node_cmd(id, "netdata show")
    # example outputs:
    # 44970 01 28000500000e10 s 7800
    # 44970 5d2920000000000000000000000000123405dc  s 7800
    # 44970 5d fddead00beef0000d88f46624df018b3d11f s a000
    service_id = f'44970 {tp:02x}'
    retval = []
    for line in lines:
        if service_id in line:
            if service_data_len > 0:
                l = len(line.split(' ')[1]) / 2
                if l == service_data_len:
                    retval.append(line)
            else:
                retval.append(line)
    return retval

def expect_count(expected_count, lines):
    actual_count = len(lines)
    if expected_count != actual_count:
        print(f"Expectation FAILED: expected = {expected_count}, actual = {actual_count}")
        traceback.print_stack()

def main():
    ns = OTNS(otns_args=['-seed','84541','-logfile', 'info'])
    ns.speed = 200
    ns.web()

    # Leader
    ns.add("router", id = LEADER_ID)
    ns.go(10)
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE1))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE3))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_ANYCAST))

    # setup of Border Routers
    for i in range(1, NUM_BR+1):
        ns.add("br", id = i)
    ns.go(20)

    # check for services getting added
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE1))
    expect_count(2, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE3))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_ANYCAST))

    # BR 1 adds a preferred unicast service
    ns.node_cmd(1, "netdata publish dnssrp unicast 2920::1234 1500")
    ns.go(20)
    expect_count(1, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE1))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE3))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_ANYCAST))

    # BR 1 changes it to anycast service
    ns.node_cmd(1, "netdata publish dnssrp anycast 1")
    ns.go(35)
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE1))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE3))
    expect_count(1, get_services(ns, LEADER_ID, TYPE_ANYCAST))

    # BR 1 changes it to non-preferred unicast service (ML-EID address)
    ns.node_cmd(1, "netdata publish dnssrp unicast 17175")
    ns.go(20)
    expect_count(0, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE1))
    expect_count(2, get_services(ns, LEADER_ID, TYPE_UNICAST, SERVICE_DATA_LEN_TYPE3))
    expect_count(0, get_services(ns, LEADER_ID, TYPE_ANYCAST))

    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
