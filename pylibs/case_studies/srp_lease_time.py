#!/usr/bin/env python3
# Copyright (c) 2024-2025, The OTNS Authors.
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

# Case study on SRP registration lease time that differs for two services, and
# clearing of a service (followed by timeout of service on SRP registrar).
# Related to Thread v1.3/1.4 test case 2.16.

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

    ns.node_cmd(cl, f'srp client host name host-test-1')
    ns.node_cmd(cl, 'srp client host address auto')

    ns.node_cmd(cl, "srp client keyleaseinterval 600")
    ns.node_cmd(cl, "srp client leaseinterval 30")
    ns.node_cmd(cl, "srp client service add service-test-1 _thread-test._udp 55555")
    ns.go(2)

    # clear (forget) the service, before reregistration happens.
    ns.node_cmd(cl, "srp client service clear service-test-1 _thread-test._udp")
    ns.go(13)

    # perform new service registration, without reregistering service-test-1
    ns.node_cmd(cl, "srp client leaseinterval 90")
    ns.node_cmd(cl, "srp client service add service-test-2 _thread-test._udp 55556")
    ns.go(14)

    # show registered services on SRP registrar - should still include service-test-1
    ns.node_cmd(srv, "srp server service")

    # let service-test-1 time out - list should not include service-test-1 active
    ns.go(2)
    ns.node_cmd(srv, "srp server service")

    ns.interactive_cli()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
