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

# Case study on multiple runs with network formation in a large, 200-node
# network in office floor topology.

import logging
import os
import resource
import shutil

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

NUM_RUNS = 1
MAX_SIM_TIME = 20*60


def build_topology(ns):
    # 34 pixels =~ 2 map grid units =~ 5 feet =~ 1.524 m
    ns.set_radioparam('MeterPerUnit', 1.524/34 )
    ns.load("etc/mesh-topologies/office_200.yaml")

def test_ulimit():
    n_files = resource.getrlimit(7)[0] # check RLIMIT_NOFILE, number of open files
    if n_files < 4096:
        print(f'Current open-files limit too low: {n_files}')
        print('Please configure "ulimit -Sn 4096" prior to running this script.')
        exit(1)

def run_formation(run_id, sim_time):
    print(f'run_formation: run_id = {run_id}')

    ns = OTNS(otns_args=['-seed',str(2342+run_id)])
    #ns.web('main')
    ns.web('stats')

    build_topology(ns)
    ns.kpi_start()
    ns.go(sim_time)
    ns.kpi_stop()
    ns.web_display()

    ns.kpi_save(f'office_runs/kpi_{run_id}.json')
    shutil.copy('tmp/0_stats.csv', f'office_runs/stats_{run_id}.csv')
    shutil.copy('current.pcap', f'office_runs/pcap_{run_id}.pcap')
    ns.delete_all()

def main():
    test_ulimit()
    try:
        os.mkdir('./office_runs')
    except:
        pass

    for n in range(1,NUM_RUNS+1):
        run_formation(run_id = n, sim_time = MAX_SIM_TIME)


if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise
