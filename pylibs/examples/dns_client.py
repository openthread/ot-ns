#!/usr/bin/env python3
# Copyright (c) 2023, The OTNS Authors.
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
# DNS client example - the PCAP file created can be checked in Wireshark to
# see the DNS messages.

from otns.cli import OTNS
from otns.cli.errors import OTNSCliError, OTNSExitedError


def main():
    ns = OTNS(otns_args=["-log", "debug"])
    ns.radiomodel = 'MIDisc'
    ns.set_title("DNS Client Example - BR server = 1, Client = 3")
    ns.web()

    idBr = ns.add("router", x=200, y=300)
    ns.add("router", x=400, y=300)
    idCl = ns.add("fed", x=600, y=300)
    ns.go(200)

    # try to send a DNS query - doesn't work, as it cannot find a route to the default DNS server address.
    cmd = 'dns resolve messagenotsent.example.com'
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # change DNS client config on client node idCl: use BR as server - 'forced' setting.
    # see https://github.com/openthread/openthread/blob/main/src/cli/README.md#dns-config
    server_ip = ns.get_ipaddrs(idBr, 'mleid')[0]
    cmd = 'dns config %s' % server_ip
    ns.node_cmd(idCl, cmd)

    # send AAAA query for Internet name
    cmd = 'dns resolve namenotfound.example.com'
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # different DNS client config
    max_tx_attempts = 1
    service_mode = 'srv'
    timeout_ms = 6000
    cmd = 'dns config %s 53 %d %d 0 %s' % (server_ip, timeout_ms, max_tx_attempts, service_mode)
    ns.node_cmd(idCl, cmd)

    # send SRV query
    cmd = 'dns service MyExampleService _thread-test._udp.default.service.arpa'
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # send PTR query
    cmd = 'dns browse _thread-test._udp.default.service.arpa'
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # send TXT query
    service_mode = 'txt'
    cmd = 'dns config %s 53 %d %d 0 %s' % (server_ip, timeout_ms, max_tx_attempts, service_mode)
    ns.node_cmd(idCl, cmd)
    cmd = 'dns service MyExampleService _thread-test._udp.default.service.arpa'
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # send AAAA query for local SRP host name
    # The name has a space, and is escaped in Python once and in Go another time. This results in a single slash
    # being sent to the OT node eventually.
    cmd = "dns resolve Example\\\\ host.default.service.arpa"
    ns.node_cmd(idCl, cmd)
    ns.go(20)

    # allow some time for graphics to be displayed in web GUI.
    ns.speed = 0.001
    ns.go(0.001)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
