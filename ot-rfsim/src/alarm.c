/*
 *  Copyright (c) 2018-2024, The OpenThread Authors.
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

#include <stdbool.h>

#include <openthread/platform/alarm-micro.h>
#include <openthread/platform/alarm-milli.h>

#define US_PER_MS 1000
#define US_PER_S 1000000
#define PS_PER_US 1000000

static uint64_t sNow           = 0; // node time in microseconds
static int16_t  sClockDriftPpm = 0; // clock drift parameter, in PPM, can be <0, 0 or >0
static int64_t  sDriftPicoSec  = 0; // current drift on sNow that happened, in picoseconds

static bool     sIsMsRunning = false;
static uint32_t sMsAlarm     = 0;

static bool     sIsUsRunning = false;
static uint32_t sUsAlarm     = 0;

void platformAlarmInit()
{
    sNow = 0;
    sDriftPicoSec = 0;
    sClockDriftPpm = 0;
}

uint64_t platformAlarmGetNow(void)
{
    return sNow;
}

void platformAlarmAdvanceNow(uint64_t aDelta)
{
    int64_t adjust;

    sNow += aDelta;

    // additional clock drift computed in picosec precision.
    sDriftPicoSec += (int64_t)sClockDriftPpm * (int64_t)aDelta;
    if (sDriftPicoSec >= PS_PER_US || sDriftPicoSec <= -PS_PER_US) { // time to adjust the microsec resolution clock?
        adjust = sDriftPicoSec / PS_PER_US;
        sNow += adjust;
        sDriftPicoSec -= adjust * PS_PER_US;
    }
}

int16_t platformAlarmGetClockDrift()
{
    return sClockDriftPpm;
}

void platformAlarmSetClockDrift(int16_t aDrift)
{
    sClockDriftPpm = aDrift;
}

uint32_t otPlatAlarmMilliGetNow(void)
{
    return (uint32_t)(sNow / US_PER_MS);
}

void otPlatAlarmMilliStartAt(otInstance *aInstance, uint32_t aT0, uint32_t aDt)
{
    OT_UNUSED_VARIABLE(aInstance);

    sMsAlarm     = aT0 + aDt;
    sIsMsRunning = true;
}

void otPlatAlarmMilliStop(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

    sIsMsRunning = false;
}

uint32_t otPlatAlarmMicroGetNow(void)
{
    return (uint32_t)sNow;
}

void otPlatAlarmMicroStartAt(otInstance *aInstance, uint32_t aT0, uint32_t aDt)
{
    OT_UNUSED_VARIABLE(aInstance);

    sUsAlarm     = aT0 + aDt;
    sIsUsRunning = true;
}

void otPlatAlarmMicroStop(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

    sIsUsRunning = false;
}

uint64_t platformAlarmGetNext(void)
{
    uint64_t remaining = INT64_MAX;

    if (sIsMsRunning)
    {
        int32_t milli = (int32_t)(sMsAlarm - otPlatAlarmMilliGetNow());

        if (milli < 0)
        {
            remaining = 0;
        }
        else
        {
            remaining = (uint64_t)milli;
            remaining *= US_PER_MS;
        }
    }

#if OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE
    if (sIsUsRunning)
    {
        int32_t micro = (int32_t)(sUsAlarm - otPlatAlarmMicroGetNow());

        if (micro < 0)
        {
            remaining = 0;
        }
        else if (remaining > ((uint64_t)micro))
        {
            remaining = (uint64_t)micro;
        }
    }
#endif

    return remaining;
}

void platformAlarmUpdateTimeout(struct timeval *aTimeout)
{
    int64_t  remaining = INT32_MAX;
    uint64_t now       = platformAlarmGetNow();

    assert(aTimeout != NULL);

    if (sIsMsRunning)
    {
        remaining = (int32_t)(sMsAlarm - (uint32_t)(now / US_PER_MS));
        if(remaining <= 0) {
            goto exit;
        }
        remaining *= US_PER_MS;
        remaining -= (now % US_PER_MS);
    }

#if OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE
    if (sIsUsRunning)
    {
        int32_t usRemaining = (int32_t)(sUsAlarm - (uint32_t)now);

        if (usRemaining < remaining)
        {
            remaining = usRemaining;
        }
    }
#endif // OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE

exit:
    if (remaining <= 0)
    {
        aTimeout->tv_sec  = 0;
        aTimeout->tv_usec = 0;
    }
    else
    {
        if (remaining < (int64_t)(aTimeout->tv_sec) * US_PER_S + (int64_t)(aTimeout->tv_usec))
        {
            aTimeout->tv_sec  = (time_t)(remaining / US_PER_S);
            aTimeout->tv_usec = (suseconds_t)(remaining % US_PER_S);
        }
    }
}

void platformAlarmProcess(otInstance *aInstance)
{
    int32_t remaining;

    if (sIsMsRunning)
    {
        remaining = (int32_t)(sMsAlarm - otPlatAlarmMilliGetNow());

        if (remaining <= 0)
        {
            sIsMsRunning = false;

#if OPENTHREAD_CONFIG_DIAG_ENABLE

            if (otPlatDiagModeGet())
            {
                otPlatDiagAlarmFired(aInstance);
            }
            else
#endif
            {
                otPlatAlarmMilliFired(aInstance);
            }
        }
    }

#if OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE

    if (sIsUsRunning)
    {
        remaining = (int32_t)(sUsAlarm - otPlatAlarmMicroGetNow());

        if (remaining <= 0)
        {
            sIsUsRunning = false;
            otPlatAlarmMicroFired(aInstance);
        }
    }

#endif // OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE
}

uint64_t otPlatTimeGet(void)
{
    return platformAlarmGetNow();
}

#if OPENTHREAD_CONFIG_TIME_SYNC_ENABLE
uint16_t otPlatTimeGetXtalAccuracy(void)
{
    return 0;
}
#endif

