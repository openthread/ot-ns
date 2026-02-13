/*
 *  Copyright (c) 2023-2025, The OpenThread Authors.
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

#include <errno.h>

#include <openthread/tcat.h>
#include <openthread/platform/ble.h>

#include "common/debug.hpp"
#include "lib/platform/exit_code.h"
#include "utils/code_utils.h"

#define OT_BLE_ADV_DELAY_MAX_US 10000
#define OT_BLE_OCTET_DURATION_US 8
#define OT_BLE_CHANNEL 37
#define OT_BLE_TX_POWER_DBM 0
#define OT_BLE_DEFAULT_ATT_MTU 23
#define OT_BLE_OVERHEAD_FACTOR 3

static bool     sEnabled, sConnected, sAdvertising;
static uint64_t sAdvPeriodUs, sNextBleAdvTime;
static uint64_t sNextBleDataTime;
static int      sFd = -1;
static uint8_t  sBleBuffer[8192];

static const uint16_t kPortBase = 10000;
static uint16_t       sPort     = 0;
struct sockaddr_in    sSockaddr;

static void initFds(void)
{
    int                fd  = -1;
    int                one = 1;
    int                flags;
    struct sockaddr_in sockaddr;

    memset(&sockaddr, 0, sizeof(sockaddr));

    sPort               = (uint16_t)(kPortBase + gNodeId);
    sockaddr.sin_family = AF_INET;
    sockaddr.sin_port   = htons(sPort);
    otEXPECT_ACTION(inet_pton(AF_INET, "127.0.0.1", &sockaddr.sin_addr) == 1, /* This should not fail */);

    otEXPECT_ACTION((fd = socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP)) != -1, perror("socket(fd)"));

    otEXPECT_ACTION(setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &one, sizeof(one)) != -1,
                    perror("setsockopt(fd, SO_REUSEADDR)"));
    otEXPECT_ACTION(setsockopt(fd, SOL_SOCKET, SO_REUSEPORT, &one, sizeof(one)) != -1,
                    perror("setsockopt(fd, SO_REUSEPORT)"));

    otEXPECT_ACTION(bind(fd, (struct sockaddr *)&sockaddr, sizeof(sockaddr)) != -1, perror("bind(fd)"));

    // Set the socket to non-blocking mode.
    otEXPECT_ACTION((flags = fcntl(fd, F_GETFL, 0)) != -1, perror("fcntl(fd, F_GETFL)"));
    flags |= O_NONBLOCK;
    otEXPECT_ACTION(fcntl(fd, F_SETFL, flags) != -1, perror("fcntl(fd, F_SETFL)"));

    // Fd is successfully initialized.
    sFd = fd;
    fd  = -1; // Ownership transferred

exit:
    if (fd != -1)
    {
        close(fd);
    }

    if (sFd == -1)
    {
        DieNow(OT_EXIT_FAILURE);
    }
}

static void deinitFds(void)
{
    if (sFd != -1)
    {
        close(sFd);
        sFd = -1;
    }
}

otError otPlatBleEnable(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sEnabled         = true;
    sConnected       = false;
    sAdvertising     = false;
    sNextBleAdvTime  = UNDEFINED_TIME_US;
    sNextBleDataTime = UNDEFINED_TIME_US;
    initFds();
    return OT_ERROR_NONE;
}

otError otPlatBleDisable(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sEnabled     = false;
    sConnected   = false;
    sAdvertising = false;
    deinitFds();
    return OT_ERROR_NONE;
}

otError otPlatBleGetAdvertisementBuffer(otInstance *aInstance, uint8_t **aAdvertisementBuffer)
{
    OT_UNUSED_VARIABLE(aInstance);
    static uint8_t sAdvertisementBuffer[OT_TCAT_ADVERTISEMENT_MAX_LEN];

    *aAdvertisementBuffer = sAdvertisementBuffer;

    return OT_ERROR_NONE;
}

void scheduleNextAdvertisement(void)
{
    // set next BLE advertisement time, after advInterval + advDelay
    // see https://www.bluetooth.com/blog/periodic-advertising-sync-transfer/
    const uint64_t advDelayUs = sAdvPeriodUs + rand() % OT_BLE_ADV_DELAY_MAX_US;
    const uint64_t now        = platformAlarmGetNow();

    sNextBleAdvTime = now + advDelayUs;

    // schedule the next "BLE advertisement" time with the simulator
    otSimSendScheduleNodeEvent(advDelayUs);
}

void scheduleNextDataPacket(uint16_t prevPacketLength)
{
    // BLE data packet duration includes any inter-packet wait times, overhead, etc - rough model
    const uint64_t now        = platformAlarmGetNow();
    const uint64_t durationUs = prevPacketLength * OT_BLE_OCTET_DURATION_US * OT_BLE_OVERHEAD_FACTOR;

    sNextBleDataTime = now + durationUs;

    // schedule the next "BLE data packet" time with the simulator
    otSimSendScheduleNodeEvent(durationUs);
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
    scheduleNextAdvertisement();

    return OT_ERROR_NONE;
}

// TODO: use advertisement data length to adapt the length (bytes) of the transmitted BLE packet
otError otPlatBleGapAdvUpdateData(otInstance *aInstance, uint8_t *aAdvertisementData, uint16_t aAdvertisementLen)
{
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aAdvertisementData);
    OT_UNUSED_VARIABLE(aAdvertisementLen);

    return OT_ERROR_NONE;
}

otError otPlatBleGapAdvStop(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sAdvertising    = false;
    sNextBleAdvTime = UNDEFINED_TIME_US;
    return OT_ERROR_NONE;
}

otError otPlatBleGapDisconnect(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    sConnected = false;
    return OT_ERROR_NONE;
}

otError otPlatBleGattMtuGet(otInstance *aInstance, uint16_t *aMtu)
{
    OT_UNUSED_VARIABLE(aInstance);
    *aMtu = OT_BLE_DEFAULT_ATT_MTU;
    return OT_ERROR_NONE;
}

otError otPlatBleGattServerIndicate(otInstance *aInstance, uint16_t aHandle, const otBleRadioPacket *aPacket)
{
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aHandle);

    ssize_t rval;
    otError error = OT_ERROR_NONE;

    otEXPECT_ACTION(sFd != -1, error = OT_ERROR_INVALID_STATE);
    otEXPECT_ACTION(sSockaddr.sin_port != 0, error = OT_ERROR_INVALID_STATE); // No destination address known

    rval = sendto(sFd, (const char *)aPacket->mValue, aPacket->mLength, 0, (struct sockaddr *)&sSockaddr,
                  sizeof(sSockaddr));
    if (rval == -1)
    {
        perror("BLE simulation sendto failed.");
        error = OT_ERROR_INVALID_STATE;
    }

    scheduleNextDataPacket(aPacket->mLength);

exit:
    return error;
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
    txData.mError    = OT_TX_TYPE_BLE_ADV;
    txData.mDuration = frameDurationUs;

    // TODO: could send actual bytes of BLE message for capture in Wireshark. Now sent as simple interference.
    otSimSendRadioCommInterferenceEvent(&txData);
}

void platformBleProcess(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);

    uint64_t  now = platformAlarmGetNow();
    socklen_t len = sizeof(sSockaddr);

    // process BLE advertisements
    if (sEnabled && sAdvertising && now >= sNextBleAdvTime)
    {
        sendBleAdvertisement();
        scheduleNextAdvertisement();
    }

    // process simulated BLE link with e.g. a TCAT Commissioner - receive non-blocking via UDP.
    if (sEnabled && now >= sNextBleDataTime)
    {
        OT_ASSERT(sFd != -1);
        ssize_t rval = recvfrom(sFd, sBleBuffer, sizeof(sBleBuffer), 0, (struct sockaddr *)&sSockaddr, &len);

        if (rval > 0)
        {
            if (!sConnected)
            {
                otPlatBleGapOnConnected(aInstance, 0);
                sConnected = true;
            }
            otBleRadioPacket myPacket;
            myPacket.mValue  = sBleBuffer;
            myPacket.mLength = (uint16_t)rval;
            myPacket.mPower  = 0;
            otPlatBleGattServerOnWriteRequest(
                aInstance, 0,
                &myPacket); // TODO consider passing otPlatBleGattServerOnWriteRequest as a callback function

            scheduleNextDataPacket(myPacket.mLength);
        }
        else if (rval == 0)
        {
            // socket is closed, which should not happen in the UDP case
            assert(false);
        }
        else if (rval == -1)
        {
            if (errno != EINTR && errno != EAGAIN && errno != EWOULDBLOCK)
            {
                perror("recvfrom BLE simulation failed");
                DieNow(OT_EXIT_FAILURE);
            }
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
    // FIXME: save the advertisement data locally, and use it to send out on BLE channel
    OT_UNUSED_VARIABLE(aInstance);
    OT_UNUSED_VARIABLE(aAdvertisementData);
    OT_UNUSED_VARIABLE(aAdvertisementLen);
    return OT_ERROR_NONE;
}

bool otPlatBleSupportsMultiRadio(otInstance *aInstance)
{
    OT_UNUSED_VARIABLE(aInstance);
    return true; // Support both Thread and BLE at the same time
}

#endif
