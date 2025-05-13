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

import time

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS()
    ns.web()

    ns.radiomodel = 'MIDisc'
    ns.speed = 10.0
    ns.set_title("Interactive simulation OTNS CLI Threaded example - switch to cmdline and type some commands.")

    # add some nodes and let them form network
    ns.add("router")
    ns.go(10)
    ns.add("router")
    ns.go(10)
    ns.add("router")
    ns.go(10)
    ns.add("router")
    ns.go(10)
    ns.add("router")
    ns.go(10)

    # here we call the threaded CLI for the user to type commands. Now the simulation can be manipulated as wanted,
    # using the CLI or GUI commands.
    ns.speed = 1.0
    ns.autogo = True
    if ns.interactive_cli_threaded():
        # if returns True, the threaded CLI was started.
        # in parallel, this script can now act on the simulation.
        for n in range(1, 20):
            print('\nSending Python scripted ping 1 -> 5')
            ns.ping(1, 5)
            time.sleep(5)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
