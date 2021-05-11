#!/usr/bin/env python3
# Copyright (c) 2021, The OTNS Authors.
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
import logging
import os
from typing import Collection

import requests

from StressTestResult import StressTestResult


class BotError(Exception):
    pass


class StressReportBot(object):
    LOGIN_NAME = 'ot-stress-report[bot]'

    def __init__(self, installid: int = None, owner: str = None, repo: str = None, issue_number: int = None):
        self.installid = installid
        self.owner = owner
        self.repo = repo
        self.issue_number = issue_number

        if not self.owner:
            self._load_github_env()

    def comment_once(self, content: str) -> dict:
        logging.info("commenting on PR: \n%s\n", content)
        return self.post('/comment-once', content=content)

    def _geturl(self, path: str) -> str:
        return 'https://ot-stress-report.glitch.me/stress-report%s/%s/%s/%s/%s' % (
            path, self.installid, self.owner, self.repo, self.issue_number
        )

    def post(self, path: str, **json) -> dict:
        url = self._geturl(path)
        res = requests.post(url, json=json)
        logging.debug("post %s %s:\n%s\n", url, json, res.content)
        return res.json()

    def _load_github_env(self):
        """
        Load GITHUB environment variables.

        'GITHUB_REPOSITORY': 'openthread/ot-ns'
        'GITHUB_REF': 'refs/pull/11/merge'
        """
        # logging.debug("os.environ = %s", {k: v for k, v in os.environ.items() if k.startswith('GITHUB_')})
        GITHUB_REPOSITORY = os.environ['GITHUB_REPOSITORY']
        GITHUB_REF = os.environ['GITHUB_REF']

        self.owner, self.repo = GITHUB_REPOSITORY.split('/')
        self.issue_number = GITHUB_REF.split('/')[2]
        logging.debug("loaded from GitHub ENV: owner=%s, repo=%s, issue_number=%s", self.owner, self.repo,
                      self.issue_number)

    def submit_suite_results(self, suite_name: str, results: Collection[StressTestResult]):
        results_json = [res.json() for res in results]
        GITHUB_RUN_ID = os.getenv('GITHUB_RUN_ID', 'local')
        reportid = f'ot_stress_tests_report_{GITHUB_RUN_ID}'
        self.post('/submit-suite-results', reportid=reportid, suite=suite_name, results=results_json)

    def report_suite_results(self):
        GITHUB_RUN_ID = os.getenv('GITHUB_RUN_ID', 'local')
        reportid = f'ot_stress_tests_report_{GITHUB_RUN_ID}'
        resp = self.post('/report-suite-results', reportid=reportid)
        if resp['error']:
            raise BotError(resp['error'])

        return resp['report']
