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

# Case study on fast key rotation (low Rotation Time field in Security Policy)
# Grep the OT node log outputs to see the KeySeqCntr event happen every 2 hour.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def test_ping(ns):
    # test ping
    ns.ping(1, 2, datasize=48)  # Parent preps ping to SED - waits in buffer
    ns.ping(2, 1, datasize=32)  # SED sends ping to Parent - this also triggers getting buffered ping from parent.
    ns.go(2)
    ns.pings()


def main():
    ns = OTNS()
    ns.loglevel = 'info'
    ns.web()

    # Router/Leader
    ns.add("router", x=300, y=200)
    ns.go(9)

    # make a copy of Active Dataset into the dataset buffer. Change security policy only.
    ns.node_cmd(1, "dataset init active")
    ns.node_cmd(1, "dataset securitypolicy 2 onrc 0")

    # set pending dataset parameters.
    ns.node_cmd(1, "dataset delay 500")
    ns.node_cmd(1, "dataset activetimestamp 1696177379")
    ns.node_cmd(1, "dataset pendingtimestamp 1696177379")

    # commit as the Pending Dataset. Delay timer starts counting down from then on.
    ns.node_cmd(1, "dataset commit pending")

    # wait until Pending Dataset has become active.
    ns.go(1)

    # add a SED
    ns.add("sed", x=300, y=300)
    ns.go(10)

    for i in range(10):
        print(f"Simulating time period {i}")
        #ns.node_cmd(1, "keysequence guardtime 0") # use this to force Router to accept new +1 tKSC value
        test_ping(ns)
        ns.go(7200)  # pass time period for next key rotation

    #ns.interactive_cli() # enable this in case interactive CLI status checking is needed at the end.
    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
