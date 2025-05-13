#!/usr/bin/env python3
# Copyright (c) 2020-2023, The OTNS Authors.
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

XGAP = 150
YGAP = 100


def main():
    ns = OTNS(otns_args=["-log", "debug", '-logfile', 'none'])
    ns.set_title("Form Partition Example")
    ns.web()
    ns.speed = float('inf')

    while True:
        # wait until next time
        for n in (2, 3, 4, 5, 6, 7, 8):
            test_nxn(ns, n)
            time.sleep(1)


def test_nxn(ns, n):
    nodes = ns.nodes()
    for id in nodes:
        ns.delete(id)

    for r in range(n):
        for c in range(n):
            ns.add("router", 100 + XGAP * c, 100 + YGAP * r)

    secs = 0
    while True:
        ns.go(1)
        secs += 1

        partitions = ns.partitions()
        if len(partitions) == 1 and 0 not in partitions:
            # all nodes converged into one partition
            break


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
