/*
 *  Copyright (c) 2023, The OpenThread Authors.
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

#if OPENTHREAD_CONFIG_BLE_TCAT_ENABLE

#include "platform-rfsim.h"
#include <openthread/tcat.h>
#include <openthread/platform/ble.h>

#define OT_BLE_ADV_DELAY_MAX_US 10000
#define OT_BLE_OCTET_DURATION_US 8
#define OT_BLE_CHANNEL 37
#define OT_BLE_TX_POWER_DBM 0
#define OT_BLE_RX_SENSITIVITY_DBM -76

static bool     sEnabled;
static bool     sAdvertising;
static uint64_t sAdvPeriodUs, sAdvDelayUs, sNextBleEventTime;

static void selectAdvertisementDelay(bool addAdvPeriod);

otError otPlatBleEnable(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sEnabled          = true;
    sNextBleEventTime = UNDEFINED_TIME_US;
    return OT_ERROR_NONE;
}

otError otPlatBleDisable(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sEnabled     = false;
    sAdvertising = false;
    return OT_ERROR_NONE;
}

otError otPlatBleGetAdvertisementBuffer(otInstance *aInstance, uint8_t **aAdvertisementBuffer)
{
    OT_UNUSED_VARIABLE(aInstance);
    static uint8_t sAdvertisementBuffer[OT_TCAT_ADVERTISEMENT_MAX_LEN];

    *aAdvertisementBuffer = sAdvertisementBuffer;

    return OT_ERROR_NONE;
}

otError otPlatBleGapAdvStart(otInstance *aInstance, uint16_t aInterval)
{
    OT_UNUSED_VARIABLE(aInstance);

    if (aInterval < OT_BLE_ADV_INTERVAL_MIN || aInterval > OT_BLE_ADV_INTERVAL_MAX)
    {
        return OT_ERROR_INVALID_ARGS;
    }

    sAdvertising = true;
    sAdvPeriodUs = aInterval * OT_BLE_ADV_INTERVAL_UNIT;
    selectAdvertisementDelay(false);
    return OT_ERROR_NONE;
}

// see https://www.bluetooth.com/blog/periodic-advertising-sync-transfer/
void selectAdvertisementDelay(bool addAdvPeriod)
{
    struct RadioStateEventData stateReport;

    uint64_t now = platformAlarmGetNow();
    sAdvDelayUs  = rand() % OT_BLE_ADV_DELAY_MAX_US;
    if (addAdvPeriod)
        sAdvDelayUs += sAdvPeriodUs;

    stateReport.mChannel       = OT_BLE_CHANNEL;
    stateReport.mTxPower       = OT_BLE_TX_POWER_DBM;
    stateReport.mRxSensitivity = OT_BLE_RX_SENSITIVITY_DBM;
    stateReport.mEnergyState   = OT_RADIO_STATE_INVALID; // TODO: no energy reporting on BLE yet
    stateReport.mSubState      = RFSIM_RADIO_SUBSTATE_READY;
    stateReport.mState         = OT_RADIO_STATE_INVALID; // TODO: no state keeping yet for BLE.
    stateReport.mRadioTime     = now;

    otSimSendRadioStateEvent(&stateReport, sAdvDelayUs);
    sNextBleEventTime = now + sAdvDelayUs;
}

otError otPlatBleGapAdvStop(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sAdvertising = false;
    return OT_ERROR_NONE;
}

otError otPlatBleGapDisconnect(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    return OT_ERROR_NOT_IMPLEMENTED;
}

otError otPlatBleGattMtuGet(otInstance *aInstance, uint16_t *aMtu)
{
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aMtu);
    return OT_ERROR_NOT_IMPLEMENTED;
}

otError otPlatBleGattServerIndicate(otInstance *aInstance, uint16_t aHandle, const otBleRadioPacket *aPacket)
{
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aHandle);
    OT_UNUSED_VARIABLE(aPacket);
    return OT_ERROR_NOT_IMPLEMENTED;
}

void sendBleAdvertisement()
{
    // see https://novelbits.io/maximum-data-bluetooth-advertising-packet-ble/
    // L_PHY = 2 + 4 + L_PDU + 3  // TODO: verify numbers
    // L_PDU = 2  + L_ADV
    // L_ADV = 6 + 31 (max)
    uint64_t frameDurationUs = (2 + 4 + 2 + 6 + 31 + 3) * OT_BLE_OCTET_DURATION_US;

    struct RadioCommEventData txData;
    txData.mChannel  = OT_BLE_CHANNEL;
    txData.mPower    = OT_BLE_TX_POWER_DBM;
    txData.mError    = OT_ERROR_NONE;
    txData.mDuration = frameDurationUs;

    struct RadioMessage aMessage;
    aMessage.mChannel = OT_BLE_CHANNEL;

    size_t msgLen     = 10;   // FIXME test
    aMessage.mPsdu[0] = 0xff; // invalid frame type - so it's not logged in pcap.
    aMessage.mPsdu[1] = 0xff;
    otSimSendRadioCommEvent(&txData, (const uint8_t *)&aMessage, msgLen + offsetof(struct RadioMessage, mPsdu));
}

void platformBleProcess(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    uint64_t now = platformAlarmGetNow();
    if (now >= sNextBleEventTime && sEnabled)
    {
        if (sAdvertising)
        {
            sendBleAdvertisement();

            // set next BLE advertisement time, after advInterval + advDelay
            selectAdvertisementDelay(true);
        }
    }
}

void otPlatBleGetLinkCapabilities(otInstance *aInstance, otBleLinkCapabilities *aBleLinkCapabilities)
{
    OT_UNUSED_VARIABLE(aInstance);
    aBleLinkCapabilities->mGattNotifications = 1;
    aBleLinkCapabilities->mL2CapDirect       = 0;
    aBleLinkCapabilities->mRsv               = 0;
}

otError otPlatBleGapAdvSetData(otInstance *aInstance, uint8_t *aAdvertisementData, uint16_t aAdvertisementLen)
{
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aAdvertisementData);
    OT_UNUSED_VARIABLE(aAdvertisementLen);
    return OT_ERROR_NOT_IMPLEMENTED;
}

bool otPlatBleSupportsMultiRadio(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    return false; // TODO check
}

#endif
