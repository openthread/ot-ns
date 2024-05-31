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

# Case study on forced key rotation by a reference device (Leader), where the
# new tKSC should be followed by the Router DUT.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

def main():
    ns = OTNS()
    ns.speed = 100
    ns.web()

    # Leader
    ns.add("router")
    ns.node_cmd(1, "keysequence counter 127")
    ns.go(10)

    # add Router
    ns.add("router")
    ns.go(10)

    # ping - both src and dst use Key Index 0x79 in Aux Security Header.
    ns.ping(1,2)
    ns.go(10)

    # force tKSC update +1
    ns.node_cmd(1, "keysequence counter 128")

    # ping again - both src and dst must use Key Index 0x01 in Aux Security Header.
    ns.ping(1,2)
    ns.go(10)

    ns.web_display()

if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
