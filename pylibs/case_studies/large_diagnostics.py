#!/usr/bin/env python3
# Copyright (c) 2023-2025, The OTNS Authors.
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

# Case study on large network diagnostics messages; where the response is
# split into multiple answers using Answer TLV. To see the results,
# use the Wireshark filter 'coap'.
#
# This study REQUIRES a Border Router with extra Child capacity.

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError


def main():
    ns = OTNS(otns_args=['-seed', '53673'])
    ns.loglevel = 'info'
    # ns.watch_default('debug')
    ns.cmd("radiomodel Ideal")
    ns.web()

    n1 = ns.add("router", x=650, y=100)
    n2 = ns.add("br", x=650, y=300, radio_range=600)

    # test if the BR has sufficient child capacity for this study
    if int(ns.node_cmd(n2, "childmax")[0]) < 100:
        print("Error: code REQUIRES a Border Router with extra Child capacity. This can be built as follows:")
        print("   cd ot-rfsim")
        print("   ./script/build_br -DOT_MLE_MAX_CHILDREN=511")
        print("   cd ..")
        exit(1)

    # add n2 to an IPv6 mcast group - a trick to receive diagnostic query mcast messages.
    ns.node_cmd(n2, "ipmaddr add ff02::d1a9")
    ns.go(10)

    # try repeated TLV Type IDs - this will create a large diagnostic response
    ns.node_cmd(n1, f'networkdiagnostic get ff02::d1a9 28 28 28 28 28 28 28 28 28 28 28 25 25 25 25 25 29 29 29 30')
    ns.go(10)

    # add a large number of Children - just to create large set of Child diagnostics.
    for n in range(0, 100):
        ns.add("med", x=200 + 50 * (n % 20), y=350 + 50 * (n // 20), radio_range=600)
        ns.go(0.2)
    ns.go(60)
    ns.save("tmp/large_diagnostics.yaml")

    for n in range(0, 2):
        # try relatively large Child info TLVs: client n1 requests to BR n2.
        ns.node_cmd(n1, f'networkdiagnostic get ff02::d1a9 29')

        # in 2nd iteration - switch off radio for a while just before receiving the rest of diagnostic answer msg.
        # To verify that the responding Thread device will then stop sending further Answer messages until the node
        # is responding again.
        if n == 1:
            ns.go(0.130)
            ns.radio_off(n2)
            ns.go(10)
            ns.radio_on(n2)
            ns.go(89.870)
        else:
            ns.go(100)

    ns.web_display()


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
