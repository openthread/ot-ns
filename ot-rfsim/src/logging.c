/*
 *  Copyright (c) 2016-2023, The OpenThread Authors.
 *  All rights reserved.
 *
 *  Redistribution and use in source and binary forms, with or without
 *  modification, are permitted provided that the following conditions are met:
 *  1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *  2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *  3. Neither the name of the copyright holder nor the
 *     names of its contributors may be used to endorse or promote products
 *     derived from this software without specific prior written permission.
 *
 *  THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 *  AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 *  IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
 *  ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
 *  LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
 *  CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
 *  SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
 *  INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
 *  CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
 *  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
 *  POSSIBILITY OF SUCH DAMAGE.
 */

#include "platform-rfsim.h"
#include <openthread-core-config.h>
#include <openthread/config.h>

#include <libgen.h>
#include <stdarg.h>
#include <stdint.h>
#include <stdio.h>
#include <syslog.h>

#include <openthread/platform/logging.h>
#include <openthread/platform/toolchain.h>
#include "common/debug.hpp"

#if (OPENTHREAD_CONFIG_LOG_OUTPUT == OPENTHREAD_CONFIG_LOG_OUTPUT_PLATFORM_DEFINED)

// specify up to which syslog log level message will still be handled. Normally we rely on log messages being sent
// in events to the simulator; so we don't need everything to go to syslog.
//#define SYSLOG_LEVEL LOG_DEBUG
#define SYSLOG_LEVEL LOG_WARNING

static int convertOtLogLevelToSyslogLevel(otLogLevel otLevel);

void platformLoggingInit(char *processName)
{
    openlog(basename(processName), LOG_PID, LOG_USER);
    setlogmask(setlogmask(0) & LOG_UPTO(SYSLOG_LEVEL));
    syslog(LOG_NOTICE, "Started process for ot-rfsim node ID: %d", gNodeId);
}

OT_TOOL_WEAK void otPlatLog(otLogLevel aLogLevel, otLogRegion aLogRegion, const char *aFormat, ...)
{
    OT_UNUSED_VARIABLE(aLogRegion);

    char    logString[512];
    int     strLen;
    va_list args;

    va_start(args, aFormat);
    strLen = vsnprintf(&logString[0], sizeof(logString) - 2, aFormat, args);
    va_end(args);
    OT_ASSERT(strLen >= 0);

    syslog(convertOtLogLevelToSyslogLevel(aLogLevel), "%s", logString);

    // extend logString with newline, and then log this string in an event.
    if (!gTerminate)
    {
        logString[strLen]     = '\n';
        logString[strLen + 1] = '\0';
        otSimSendLogWriteEvent((const uint8_t *)&logString[0], strLen + 1);
    }
}

int convertOtLogLevelToSyslogLevel(otLogLevel otLevel)
{
    switch (otLevel)
    {
    case OT_LOG_LEVEL_CRIT:
        return LOG_CRIT;
    case OT_LOG_LEVEL_WARN:
        return LOG_WARNING;
    case OT_LOG_LEVEL_NOTE:
        return LOG_NOTICE;
    case OT_LOG_LEVEL_INFO:
        return LOG_INFO;
    case OT_LOG_LEVEL_DEBG:
        return LOG_DEBUG;
    default:
        return LOG_CRIT;
    }
}
#endif
