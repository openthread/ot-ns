/*
 *  Copyright (c) 2016-2024, The OpenThread Authors.
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

#include <setjmp.h>
#include <unistd.h>

#include "openthread-system.h"
#include <openthread/logging.h>
#include <openthread/platform/misc.h>

#include "common/logging.hpp"

extern jmp_buf      gResetJump;
extern struct Event gLastSentEvent, gLastRecvEvent;

otPlatResetReason   gPlatResetReason = OT_PLAT_RESET_REASON_POWER_ON;
bool                gPlatformPseudoResetWasRequested;
otPlatMcuPowerState gPlatMcuPowerState = OT_PLAT_MCU_POWER_STATE_ON;

void otPlatReset(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

#if OPENTHREAD_PLATFORM_USE_PSEUDO_RESET
    gPlatformPseudoResetWasRequested = true;
    gPlatResetReason                 = OT_PLAT_RESET_REASON_SOFTWARE;

#else // OPENTHREAD_PLATFORM_USE_PSEUDO_RESET
    // Restart the process using execvp.
    otSysDeinit();
    platformUartRestore();

    longjmp(gResetJump, 1);
    assert(false);

#endif // OPENTHREAD_PLATFORM_USE_PSEUDO_RESET
}

#if OPENTHREAD_CONFIG_PLATFORM_ASSERT_MANAGEMENT
void otPlatAssertFail(const char *aFilename, int aLineNumber)
{
    otLogCritPlat("assert failed at %s:%d\n", aFilename, aLineNumber);
    otLogCritPlat("Last sent Event: tp=%i dly=%lu datalen=%u\n", gLastSentEvent.mEvent,
                  (unsigned long)gLastSentEvent.mDelay, gLastSentEvent.mDataLength);
    otLogCritPlat("Last recv Event: tp=%i dly=%lu datalen=%u\n", gLastRecvEvent.mEvent,
                  (unsigned long)gLastRecvEvent.mDelay, gLastRecvEvent.mDataLength);

    fprintf(stderr, "assert failed at %s:%d\n", aFilename, aLineNumber);

    // For debug build, use assert to generate a core dump
    assert(false);
    exit(1);
}
#endif

otPlatResetReason otPlatGetResetReason(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

    return gPlatResetReason;
}

void otPlatWakeHost(void)
{
    // TODO: implement an operation to wake the host from sleep state.
}

otError otPlatSetMcuPowerState(otInstance *aInstance, otPlatMcuPowerState aState)
{
    OT_UNUSED_VARIABLE(aInstance);

    otError error = OT_ERROR_NONE;

    switch (aState)
    {
    case OT_PLAT_MCU_POWER_STATE_ON:
    case OT_PLAT_MCU_POWER_STATE_LOW_POWER:
        gPlatMcuPowerState = aState;
        break;

    default:
        error = OT_ERROR_FAILED;
        break;
    }

    return error;
}

otPlatMcuPowerState otPlatGetMcuPowerState(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

    return gPlatMcuPowerState;
}
