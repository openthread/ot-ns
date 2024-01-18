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

# Case study on large, 200-node network in office floor topology.

import logging
import resource

from otns.cli import OTNS
from otns.cli.errors import OTNSExitedError

def build_topology(ns, radio_rng):
    # Below code ns.add() lines were built using an office map and ./pylibs/node_clicker.py
    # 34 pixels =~ 2 map grid units =~ 5 feet =~ 1.524 m
    ns.set_radioparam('MeterPerUnit', 1.524/34 )
    node_tp1 = 'router'
    node_tp1_radiorange = radio_rng
    # Nodes 1-70
    ns.add(node_tp1, x=412, y=85, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=446, y=85, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=533, y=83, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=573, y=65, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=609, y=82, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=651, y=65, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=691, y=83, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=689, y=120, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=411, y=119, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=411, y=165, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=451, y=122, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=447, y=163, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=529, y=123, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=533, y=163, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=574, y=140, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=608, y=120, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=610, y=161, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=653, y=141, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=686, y=160, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=687, y=198, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=416, y=236, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=407, y=198, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=450, y=235, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=443, y=196, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=200, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=535, y=237, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=572, y=216, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=614, y=238, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=618, y=201, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=650, y=216, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=275, y=296, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=336, y=293, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=356, y=314, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=385, y=317, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=410, y=297, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=468, y=315, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=511, y=315, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=541, y=306, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=291, y=352, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=295, y=397, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=355, y=356, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=386, y=352, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=451, y=353, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=452, y=390, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=353, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=533, y=391, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=296, y=432, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=294, y=472, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=447, y=437, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=468, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=538, y=429, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=534, y=464, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=298, y=506, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=285, y=537, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=373, y=491, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=334, y=527, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=446, y=510, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=415, y=525, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=534, y=505, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=574, y=525, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=614, y=507, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=652, y=527, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=688, y=512, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=649, y=488, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=284, y=574, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=293, y=607, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=337, y=588, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=367, y=603, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=413, y=587, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=448, y=602, radio_range=node_tp1_radiorange)

    # node 71-200
    ns.add(node_tp1, x=536, y=603, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=574, y=588, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=613, y=605, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=655, y=584, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=54, y=645, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=58, y=680, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=99, y=645, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=153, y=641, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=193, y=643, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=211, y=679, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=291, y=642, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=296, y=681, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=333, y=642, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=391, y=645, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=643, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=683, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=535, y=644, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=680, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=596, y=645, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=631, y=643, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=689, y=643, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=688, y=682, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=58, y=719, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=59, y=758, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=214, y=720, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=212, y=758, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=294, y=720, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=293, y=759, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=448, y=722, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=452, y=758, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=721, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=758, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=687, y=722, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=688, y=758, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=59, y=799, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=98, y=820, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=139, y=759, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=136, y=801, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=176, y=816, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=213, y=799, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=296, y=799, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=331, y=817, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=372, y=762, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=374, y=800, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=413, y=817, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=796, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=537, y=797, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=576, y=817, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=613, y=760, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=614, y=797, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=651, y=816, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=689, y=799, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=58, y=898, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=101, y=872, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=134, y=932, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=138, y=896, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=177, y=876, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=211, y=892, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=296, y=894, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=332, y=877, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=373, y=936, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=377, y=894, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=412, y=876, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=891, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=535, y=894, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=576, y=874, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=616, y=934, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=615, y=892, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=657, y=876, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=691, y=896, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=56, y=937, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=60, y=975, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=210, y=934, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=213, y=972, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=296, y=933, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=298, y=972, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=933, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=451, y=973, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=540, y=935, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=536, y=969, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=692, y=933, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=693, y=972, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=57, y=1014, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=61, y=1048, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=120, y=1049, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=157, y=1048, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=216, y=1011, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=214, y=1049, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=299, y=1009, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=298, y=1047, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=357, y=1050, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=390, y=1050, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=449, y=1011, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=451, y=1050, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=538, y=1010, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=541, y=1052, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=576, y=1049, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=614, y=1046, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=689, y=1013, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=689, y=1047, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=55, y=1089, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=80, y=1106, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=119, y=1086, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=160, y=1108, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=195, y=1086, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=213, y=1105, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=283, y=1109, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=303, y=1087, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=376, y=1088, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=334, y=1107, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=452, y=1085, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=417, y=1107, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=540, y=1086, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=578, y=1107, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=613, y=1087, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=652, y=1105, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=652, y=1068, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=691, y=1088, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=285, y=1166, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=317, y=1182, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=336, y=1164, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=370, y=1181, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=410, y=1161, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=448, y=1180, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=533, y=1181, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=571, y=1163, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=595, y=1183, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=631, y=1185, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=655, y=1161, radio_range=node_tp1_radiorange)
    ns.add(node_tp1, x=690, y=1181, radio_range=node_tp1_radiorange)

def test_ulimit():
    n_files = resource.getrlimit(7)[0] # check RLIMIT_NOFILE, number of open files
    if n_files < 4096:
        print(f'Current open-files limit too low: {n_files}')
        print('Please configure "ulimit -Sn 4096" prior to running this script.')
        exit(1)

def main():
    test_ulimit()

    ns = OTNS()
    ns.logconfig(logging.INFO)
    ns.loglevel = 'info'
    ns.radiomodel = 'MutualInterference'
    # ns.radiomodel = 'Ideal'
    ns.web('main')
    ns.web('stats')

    build_topology(ns, radio_rng=320)
    ns.go(13*60)
    ns.interactive_cli()

    ns.web_display()

if __name__ == '__main__':
    try:
        main()
    except OTNSExitedError as ex:
        if ex.exit_code != 0:
            raise

