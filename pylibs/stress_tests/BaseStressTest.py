#!/usr/bin/env python3
#
# Copyright (c) 2020, The OTNS Authors.
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
import os
import sys
import time
import traceback
from functools import wraps

from StressTestResult import StressTestResult
from otns.cli import OTNS


class StressTestMetaclass(type):
    def __new__(cls, name, bases, dct):
        assert 'run' in dct, f'run method is not defined in {name}'

        orig_run = dct.pop('run')

        @wraps(orig_run)
        def run_wrapper(self: 'BaseStressTest', report=True):
            try:
                orig_run(self)
            except Exception as ex:
                traceback.print_exc()
                self.result.fail_with_error(ex)
            finally:
                self.stop()

            if report:
                self.report()

        dct['run'] = run_wrapper

        t = super().__new__(cls, name, bases, dct)
        return t


class BaseStressTest(object, metaclass=StressTestMetaclass):
    def __init__(self, name, headers, raw=False):
        self.name = name
        self._otns_args = []
        if raw:
            self._otns_args.append('-raw')
        self.ns = OTNS(otns_args=self._otns_args)
        self.ns.speed = float('inf')
        self.ns.web()

        self.result = StressTestResult(name=name, headers=headers)
        self.result.start()

    def run(self):
        raise NotImplementedError()

    def reset(self):
        nodes = self.ns.nodes()
        if nodes:
            self.ns.delete(*nodes.keys())

    def stop(self):
        self.result.stop()
        self.ns.close()

    def report(self):
        try:
            STRESS_RESULT_FILE = os.environ['STRESS_RESULT_FILE']
            stress_result_fd = open(STRESS_RESULT_FILE, 'wt')
        except KeyError:
            stress_result_fd = sys.stdout

        with stress_result_fd:
            stress_result_fd.write(
                f"""**[OTNS](https://github.com/openthread/ot-ns) Stress Tests Report Generated at {time.strftime(
                    "%m/%d %H:%M:%S")}**\n""")
            stress_result_fd.write(self.result.format())
