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

# Case study on traffic patterns simulation - unicast, multicast.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def build_topology(ns):
    # 34 pixels =~ 2 map grid units =~ 5 feet =~ 1.524 m
    ns.set_radioparam('MeterPerUnit', 1.524 / 34)
    ns.load("etc/mesh-topologies/office_50.yaml")


def main():
    ns = OTNS()
    ns.radiomodel = 'MutualInterference'
    ns.web('stats')
    ns.web('main')

    build_topology(ns)
    ns.go(30)
    ns.set_title('starting data traffic test')
    ns.speed = 1

    # send unicast and multicast traffic over mesh
    ns.kpi_start()
    ns.cmd("send coap con 1 50 ds 64")  # unicast from BR to node
    ns.go(0.002)
    ns.cmd("send coap 48 31-50")  # sensor triggers lights of lower group
    ns.go(5)
    ns.kpi_save('tmp/cs_traffic_patterns.json')
    ns.kpi_save()

    ns.set_title('data traffic test done')
    ns.interactive_cli()
    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
