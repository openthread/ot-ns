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
# Rekeying the entire mesh network example.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS(otns_args=["-log", "debug"])
    ns.speed = 32
    ns.radiomodel = 'MIDisc'
    ns.set_title("Network Key Rekeying Example - Setting up network")
    ns.web()

    ns.add("router", x=200, y=300) #Leader
    ns.go(10)
    ns.add("router")
    ns.add("router")
    ns.add("router")
    ns.add("router", x=200, y=450)
    ns.add("router")
    ns.add("router")
    ns.add("router")
    ns.add("router")
    ns.add("router", x=200, y=600)
    ns.add("router")
    ns.add("router")
    ns.add("router")
    ns.add("fed")
    ns.add("fed")
    ns.add("sed")
    ns.add("router")
    ns.go(10)
    ns.add("med")
    ns.add("sed")
    ns.add("sed")
    ns.add("sed")
    ns.go(120)

    # print Active Dataset
    ns.set_title("Network Key Rekeying Example - Waiting on delay timer to set new Network Key and Channel")
    print(*ns.node_cmd(1,"dataset active"), sep='\n')

    # make a copy of Active Dataset into the dataset buffer. Change network key & timestamp & channel.
    ns.node_cmd(1, "dataset init active")
    ns.node_cmd(1, "dataset networkkey 506071d8391be671569e080d52870fd5")
    ns.node_cmd(1, "dataset activetimestamp 1696177379")
    ns.node_cmd(1, "dataset channel 16")

    # set pending dataset parameters.
    ns.node_cmd(1, "dataset delay 200000")
    ns.node_cmd(1,"dataset pendingtimestamp 1696177379")

    # commit as the Pending Dataset. Delay timer starts counting down from then on.
    ns.node_cmd(1, "dataset commit pending")

    # wait until Pending Dataset has been distributed and all nodes' delay timers have counted to 0.
    ns.go(200)

    # simulate some time with network using the new network key.
    ns.set_title("Network Key Rekeying Example - New Network Key and Channel are active")
    ns.go(500)

    # allow some time for graphics to be displayed in web GUI.
    ns.speed=0.001
    ns.go(0.001)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
