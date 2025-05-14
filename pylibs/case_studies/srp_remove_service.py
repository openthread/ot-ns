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

# Case study on SRP removal of a service.
# Related to Thread v1.3/1.4 test case 2.8

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS(otns_args=['-seed', '34541'])
    ns.web()
    ns.watch_all('info')

    srv = ns.add('br')
    ns.go(10)

    cl = ns.add('router')
    ns.go(10)

    # print version info
    ns.node_cmd(cl, 'version')
    ns.node_cmd(srv, 'version')

    # init srp config
    ns.node_cmd(cl, f'srp client host name host-test-1')
    ns.node_cmd(cl, 'srp client host address auto')
    ns.node_cmd(cl, "srp client keyleaseinterval 1200")
    ns.node_cmd(cl, "srp client leaseinterval 600")
    ns.node_cmd(cl, "srp client autostart enable")

    # init dns client config
    ns.node_cmd(cl, "dns config :: 0 0 0 0 srv_txt")

    # register 2 services
    ns.node_cmd(cl, "srp client service add service-test-1 _thread-test._udp 55555")
    ns.go(3)

    ns.node_cmd(cl, "srp client autostart enable")
    ns.node_cmd(cl, "srp client service add service-test-2 _thread-test._udp 55555")
    ns.go(3)

    # show registered services on SRP registrar - should include both
    ns.node_cmd(srv, "srp server service")
    ns.go(30)

    # let client query - should get it
    ns.node_cmd(cl, "dns service service-test-1 _thread-test._udp.default.service.arpa")
    ns.go(30)

    # remove the service.
    ns.node_cmd(cl, "srp client service remove service-test-1 _thread-test._udp")
    ns.node_cmd(cl, "netdata register")  #try to reproduce what ref-device did
    ns.node_cmd(cl, "srp client keyleaseinterval 600")  #try to reproduce what ref-device did
    ns.go(15)

    # show registered services on SRP registrar - should not include service-test-1
    ns.node_cmd(srv, "srp server service")

    # let client query - should not get 1
    ns.node_cmd(cl, "dns service service-test-1 _thread-test._udp.default.service.arpa")
    ns.go(30)

    # let client query - should get 2
    ns.node_cmd(cl, "dns service service-test-2 _thread-test._udp.default.service.arpa")
    ns.go(30)

    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
