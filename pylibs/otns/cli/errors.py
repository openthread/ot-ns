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

"""
This file contains definitions of OTNS CLI errors
"""

from otns.errors import OTNSError


class OTNSCliError(OTNSError):
    def __init__(self, error: str):
        super(OTNSCliError, self).__init__(error)

class OTNSExitedError(OTNSCliError):
    """
    OTNS exited.
    """

    def __init__(self, exit_code: int):
        super(OTNSExitedError, self).__init__(f"exited: {exit_code}")
        self.exit_code = exit_code


class OTNSCommandInterruptedError(OTNSExitedError):
    """
    Command was interrupted due to OTNS exiting.
    """

    def __init__(self):
        super(OTNSExitedError, self).__init__("command interrupted")
        self.exit_code = 0


class UnexpectedError(OTNSError):
    def __init__(self, error: str):
        super(UnexpectedError, self).__init__(error)


def create_otns_cli_error(error_line: str):
    """
    Create an OTNSCliError basd on the error line reported by OTNS.
    :param error_line: the error line
    :return: OTNSCliError or subclass error (see types in otns.cli.errors)
    """
    if error_line.startswith("Error: command interrupted"):
        return OTNSCommandInterruptedError()
    if error_line.startswith("Error: "):
        return OTNSCliError(error_line[7:])
    return OTNSCliError("Error: " + error_line)