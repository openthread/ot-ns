#!/usr/bin/env python3
# Copyright (c) 2025, The OTNS Authors.
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

# Case study on use of CoAP observe.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS(otns_args=[])
    ns.speed = 1000

    # CoAP server
    srv = ns.add("router")
    ns.node_cmd(srv, "coap start")
    ns.node_cmd(srv, "coap resource test-resource")
    ns.node_cmd(srv, "coap set TestPayload_1")
    srv_addr = ns.get_ipaddrs(srv, 'mleid')[0]

    # CoAP client
    cl = ns.add("router")
    ns.node_cmd(cl, "coap start")

    # form the network
    ns.kpi_start()
    ns.go(10)

    # client sends observe request
    ns.node_cmd(cl, f"coap observe {srv_addr} test-resource")
    ns.go(90)

    # time passes, and resource changes
    ns.node_cmd(srv, "coap set TestPayload_2")
    ns.go(100)

    ns.node_cmd(srv, "coap set TestPayload_3")
    ns.go(100)

    ns.node_cmd(srv, "coap set TestPayload_4")
    ns.go(100)

    # now repeat the observe requests using CON type. This triggers CON notifications to be sent.
    ns.node_cmd(cl, "coap cancel")
    ns.go(100)
    ns.node_cmd(cl, f"coap observe {srv_addr} test-resource con")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_5")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_6")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_7")
    ns.go(100)

    # send a cancel request
    ns.node_cmd(cl, f"coap cancel")
    ns.go(100)

    # server sends an interspersed CON notification after sending 5 NON notifications
    ns.node_cmd(cl, f"coap observe {srv_addr} test-resource")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_8")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_9")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_a")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_b")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_c")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_d")  # this one will be sent CON
    ns.go(0.001)
    # this fast one will be NON. Note that the previous one isn't Ack'ed yet.
    ns.node_cmd(srv, "coap set TestPayload_d2")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_e")
    ns.go(100)
    ns.node_cmd(srv, "coap set TestPayload_e")
    ns.go(100)

    # observe another resource (a non-existing one)
    ns.node_cmd(cl, f"coap observe {srv_addr} resnotfound")
    ns.go(10)

    ns.kpi_stop()
    ns.web_display()

if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
