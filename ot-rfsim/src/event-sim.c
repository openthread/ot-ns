/*
 *  Copyright (c) 2022-2024, The OpenThread Authors.
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
 *   This file includes simulation-event message formatting, sending and parsing functions.
 */

#include "common/debug.hpp"

#include "platform-rfsim.h"
#include "event-sim.h"

// socket communication parameters for events
extern int      gSockFd;

struct Event gLastSentEvent;

void otSimSendSleepEvent(void)
{
    OT_ASSERT(platformAlarmGetNext() > 0);
    struct Event event;

    event.mDelay      = platformAlarmGetNext();
    event.mEvent      = OT_SIM_EVENT_ALARM_FIRED;
    event.mDataLength = 0;

    otSimSendEvent(&event);
}

void otSimSendRadioCommEvent(struct RadioCommEventData *aEventData, const uint8_t *aPayload, size_t aLenPayload)
{
    OT_ASSERT(aLenPayload <= OT_EVENT_DATA_MAX_SIZE);
    struct Event event;

    event.mEvent = OT_SIM_EVENT_RADIO_COMM_START;
    memcpy(event.mData, aEventData, sizeof(struct RadioCommEventData));
    memcpy(event.mData + sizeof(struct RadioCommEventData), aPayload, aLenPayload);
    event.mDataLength = sizeof(struct RadioCommEventData) + aLenPayload;

    otSimSendEvent(&event);
}

void otSimSendRadioCommInterferenceEvent(struct RadioCommEventData *aEventData)
{
    struct Event event;

    event.mEvent = OT_SIM_EVENT_RADIO_COMM_START;
    memcpy(event.mData, aEventData, sizeof(struct RadioCommEventData));
    event.mData[sizeof(struct RadioCommEventData)] = aEventData->mChannel; // channel is stored twice TODO
    event.mDataLength = sizeof(struct RadioCommEventData) + 1;

    otSimSendEvent(&event);
}

void otSimSendRadioChanSampleEvent(struct RadioCommEventData *aChanData)
{
    struct Event event;

    event.mEvent = OT_SIM_EVENT_RADIO_CHAN_SAMPLE;
    event.mDelay = 0;
    memcpy(event.mData, aChanData, sizeof(struct RadioCommEventData));
    event.mDataLength = sizeof(struct RadioCommEventData);

    otSimSendEvent(&event);
}

void otSimSendRadioStateEvent(struct RadioStateEventData *aStateData, uint64_t aDeltaUntilNextRadioState)
{
    struct Event event;

    event.mEvent = OT_SIM_EVENT_RADIO_STATE;
    event.mDelay = aDeltaUntilNextRadioState;
    memcpy(event.mData, aStateData, sizeof(struct RadioStateEventData));
    event.mDataLength = sizeof(struct RadioStateEventData);

    otSimSendEvent(&event);
}

void otSimSendUartWriteEvent(const uint8_t *aData, uint16_t aLength) {
    OT_ASSERT(aLength <= OT_EVENT_DATA_MAX_SIZE);
    struct Event event;

    event.mEvent      = OT_SIM_EVENT_UART_WRITE;
    event.mDelay      = 0;
    event.mDataLength = aLength;
    memcpy(event.mData, aData, aLength);

    otSimSendEvent(&event);
}

void otSimSendLogWriteEvent(const uint8_t *aData, uint16_t aLength) {
    OT_ASSERT(aLength <= OT_EVENT_DATA_MAX_SIZE);

    struct Event event;
    event.mEvent      = OT_SIM_EVENT_LOG_WRITE;
    event.mDelay      = 0;
    event.mDataLength = aLength;
    memcpy(event.mData, aData, aLength);

    otSimSendEvent(&event);
}

void otSimSendOtnsStatusPushEvent(const char *aStatus, uint16_t aLength) {
    OT_ASSERT(aLength <= OT_EVENT_DATA_MAX_SIZE);
    struct Event event;

    memcpy(event.mData, aStatus, aLength);
    event.mEvent      = OT_SIM_EVENT_OTNS_STATUS_PUSH;
    event.mDelay      = 0;
    event.mDataLength = aLength;

    otSimSendEvent(&event);
}

void otSimSendExtAddrEvent(const otExtAddress *aExtAddress) {
    OT_ASSERT(aExtAddress != NULL);
    struct Event event;

    memcpy(event.mData, aExtAddress, sizeof(otExtAddress));
    event.mEvent      = OT_SIM_EVENT_EXT_ADDR;
    event.mDelay      = 0;
    event.mDataLength = sizeof(otExtAddress);

    otSimSendEvent(&event);
}

void otSimSendNodeInfoEvent(uint32_t nodeId) {
    struct Event event;
    OT_ASSERT(nodeId > 0);

    memcpy(event.mData, &nodeId, sizeof(uint32_t));
    event.mEvent      = OT_SIM_EVENT_NODE_INFO;
    event.mDelay      = 0;
    event.mDataLength = sizeof(uint32_t);

    otSimSendEvent(&event);
}

void otSimSendRfSimParamRespEvent(uint8_t param, int32_t value) {
    struct Event event;

    event.mData[0] = param;
    memcpy(event.mData + 1, &value, sizeof(int32_t));
    event.mEvent      = OT_SIM_EVENT_RFSIM_PARAM_RSP;
    event.mDelay      = 0;
    event.mDataLength = sizeof(int32_t) + 1;

    otSimSendEvent(&event);
}

void otSimSendMsgToHostEvent(uint8_t evType, struct MsgToHostEventData *aEventData, uint8_t *aMsgBytes, size_t aMsgLen) {
    const size_t evDataSz = sizeof(struct MsgToHostEventData);
    struct Event event;
    OT_ASSERT(aMsgLen <= OT_EVENT_DATA_MAX_SIZE - evDataSz);

    event.mEvent = evType;
    event.mDelay = 0;
    memcpy(event.mData, aEventData, evDataSz);
    memcpy(event.mData + evDataSz, aMsgBytes, aMsgLen);
    event.mDataLength = evDataSz + aMsgLen;

    otSimSendEvent(&event);
}

void otSimSendEvent(struct Event *aEvent)
{
    ssize_t rval;

    aEvent->mMsgId = gLastMsgId;
    gLastSentEvent = *aEvent;

    if (gSockFd == 0)   // don't send events if socket invalid.
        return;

    // send header and data.
    rval = write(gSockFd, aEvent, offsetof(struct Event, mData) + aEvent->mDataLength);

    if (rval < 0)
    {
        perror("write");
        platformExit(EXIT_FAILURE);
    }
}
