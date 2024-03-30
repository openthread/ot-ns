/*
 *  Copyright (c) 2018-2023, The OpenThread Authors.
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

/**
 * @file
 * @brief
 *   This file includes the platform-specific initializers and processing functions.
 */

#include "platform-rfsim.h"

#include <assert.h>
#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#include <openthread/tasklet.h>
#include "common/debug.hpp"

#include "event-sim.h"
#include "utils/uart.h"

#define VERIFY_EVENT_SIZE(X) OT_ASSERT( (payloadLen >= sizeof(X)) && "received event payload too small" );

extern int gSockFd;

uint64_t gLastMsgId = 0;
struct Event gLastRecvEvent;

void platformExit(int exitCode) {
    gTerminate = true;
    otPlatLog(OT_LOG_LEVEL_NOTE,OT_LOG_REGION_PLATFORM,
              "Exiting with exit code %d.", exitCode);
    exit(exitCode);
}

void platformReceiveEvent(otInstance *aInstance)
{
    struct Event event;
    ssize_t      rval = recvfrom(gSockFd, (char *)&event, sizeof(struct EventHeader), 0, NULL, NULL);
    const uint8_t *evData = event.mData;

    if (rval < 0)
    {
        perror("recvfrom");
        platformExit(EXIT_FAILURE);
    }
    else if ((uint16_t)rval < sizeof(struct EventHeader)) {
        fprintf(stderr, "incomplete event received, len=%li", rval);
        platformExit(EXIT_FAILURE);
    }

    // read the rest of data (payload data - optional).
    uint16_t payloadLen = event.mDataLength;
    if (payloadLen > 0) {
        if (payloadLen > sizeof(event.mData)) {
            fprintf(stderr, "too-large event payload detected, len=%u, expected <= %lu", payloadLen, sizeof(event.mData));
            platformExit(EXIT_FAILURE);
        }
        rval = recvfrom(gSockFd, (char *)&event.mData, payloadLen, 0, NULL, NULL);
        if (rval < 0)
        {
            perror("recvfrom");
            platformExit(EXIT_FAILURE);
        }
        else if ((uint16_t)rval < payloadLen) {
            fprintf(stderr, "incomplete event payload received, len=%li, expected=%u", rval, payloadLen);
            platformExit(EXIT_FAILURE);
        }
    }

    gLastRecvEvent = event;
    gLastMsgId = event.mMsgId;

    platformAlarmAdvanceNow(event.mDelay);

    switch (event.mEvent)
    {
    case OT_SIM_EVENT_ALARM_FIRED:
        // Alarm events may be used to wake the node again when some simulated time has passed.
        break;

    case OT_SIM_EVENT_UART_WRITE:
        otPlatUartReceived(event.mData, event.mDataLength);
        break;

    case OT_SIM_EVENT_RADIO_COMM_START:
        VERIFY_EVENT_SIZE(struct RadioCommEventData)
        platformRadioRxStart(aInstance, (struct RadioCommEventData *)evData);
        break;

    case OT_SIM_EVENT_RADIO_RX_DONE:
        VERIFY_EVENT_SIZE(struct RadioCommEventData)
        const size_t sz = sizeof(struct RadioCommEventData);
        platformRadioRxDone(aInstance, evData + sz,
                       event.mDataLength - sz, (struct RadioCommEventData *)evData);
        break;

    case OT_SIM_EVENT_RADIO_TX_DONE:
        VERIFY_EVENT_SIZE(struct RadioCommEventData)
        platformRadioTxDone(aInstance, (struct RadioCommEventData *)evData);
        break;

    case OT_SIM_EVENT_RADIO_CHAN_SAMPLE:
        VERIFY_EVENT_SIZE(struct RadioCommEventData)
        // TODO consider also energy-detect case. This only does CCA now.
        platformRadioCcaDone(aInstance, (struct RadioCommEventData *)evData);
        break;

    case OT_SIM_EVENT_RFSIM_PARAM_GET:
        VERIFY_EVENT_SIZE(struct RfSimParamEventData)
        platformRadioRfSimParamGet(aInstance, (struct RfSimParamEventData *)evData);
        break;

    case OT_SIM_EVENT_RFSIM_PARAM_SET:
        VERIFY_EVENT_SIZE(struct RfSimParamEventData)
        platformRadioRfSimParamSet(aInstance, (struct RfSimParamEventData *)evData);
        platformRadioReportStateToSimulator(true);
        break;

    default:
        OT_ASSERT(false && "Unrecognized event type received");
    }
}

void otPlatOtnsStatus(const char *aStatus)
{
    uint16_t     statusLength = (uint16_t)strlen(aStatus);
    if (statusLength > OT_EVENT_DATA_MAX_SIZE){
        statusLength = OT_EVENT_DATA_MAX_SIZE;
        OT_ASSERT(statusLength <= OT_EVENT_DATA_MAX_SIZE);
    }
    otSimSendOtnsStatusPushEvent(aStatus, statusLength);
}
