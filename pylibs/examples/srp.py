#!/usr/bin/env python3
# Copyright (c) 2023-2024, The OTNS Authors.
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
# SRP client and server example.

import time
from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS(otns_args=['-log', 'debug'])
    ns.radiomodel = 'MIDisc'
    ns.set_title("SRP Example - BR server = 1, Client = 3")
    ns.web()

    # start an SRP server on a BR (non-BRs don't have SRP server build flag enabled)
    id_br = ns.add('br', x=200, y=300)
    # starting the server manually is not really needed, due to the autostart feature for BR.
    # But it's left here to show the command to manually enable/disable.
    ns.node_cmd(id_br,'srp server enable')
    ns.go(10)

    ns.add('router', x=400, y=300)
    ns.go(10)

    # start an SRP client
    id_cl = ns.add('fed', x=600, y=300)
    ns.go(200) # form network
    ns.node_cmd(id_cl,'srp client host name MyExampleHost')
    ns.node_cmd(id_cl,'srp client host address auto')

    # client registers an SRP service
    ns.node_cmd(id_cl,'srp client service add MyExampleInstance _otns-test._udp 8080 1 2')
    ns.node_cmd(id_cl,'srp client autostart enable')
    ns.go(50)

    # client: check status
    ns.node_cmd(id_cl,'srp client host')
    ns.node_cmd(id_cl,'srp client service')

    # server: check status
    ns.node_cmd(id_br,'srp server host')
    ns.node_cmd(id_br,'srp server service')

    ns.go(100)

    # register another service
    ns.node_cmd(id_cl,'srp client service add TestService _thread-test._udp 8081 0 0')
    ns.go(50)

    # client: check status
    ns.node_cmd(id_cl,'srp client host')
    ns.node_cmd(id_cl,'srp client service')

    # server: check status
    ns.node_cmd(id_br,'srp server host')
    ns.node_cmd(id_br,'srp server service')

    # client: remove host and all services
    ns.node_cmd(id_cl, 'srp client host remove')
    ns.go(10)

    # server: check status
    ns.node_cmd(id_br,'srp server host')
    ns.node_cmd(id_br,'srp server service')

    # allow some time for graphics to be displayed in web GUI.
    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
