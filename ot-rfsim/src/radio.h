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

#ifndef PLATFORM_RFSIM_RADIO_H
#define PLATFORM_RFSIM_RADIO_H

#include "radio-parameters.h"
#include <openthread/platform/radio.h>

#define FAILSAFE_TIME_US 1

// platform-specific OT_ERROR status code to indicate an interference Tx
#define OT_TX_TYPE_INTF 192

// IEEE 802.15.4 related parameters. See radio-parameters.h for radio-model-specific parameters.
enum
{
    kMinChannel                     = OT_RADIO_2P4GHZ_OQPSK_CHANNEL_MIN,
    kMaxChannel                     = OT_RADIO_2P4GHZ_OQPSK_CHANNEL_MAX,
    OT_RADIO_LIFS_TIME_US           = 40 * OT_RADIO_SYMBOL_TIME, // From 802.15.4 spec, LIFS
    OT_RADIO_SIFS_TIME_US           = 12 * OT_RADIO_SYMBOL_TIME, // From 802.15.4 spec, SIFS
    OT_RADIO_AIFS_TIME_US           = 12 * OT_RADIO_SYMBOL_TIME, // From 802.15.4 spec, AIFS
    OT_RADIO_CCA_TIME_US            = 8 * OT_RADIO_SYMBOL_TIME,  // From 802.15.4 spec, CCA duration
    OT_RADIO_SHR_DURATION_US        = 5 * OT_RADIO_SYMBOLS_PER_OCTET * OT_RADIO_SYMBOL_TIME, // sync header (SHR)
    OT_RADIO_SHR_PHR_LENGTH_BYTES   = 6, // SHR + PHY header (PHR) length in bytes
    OT_RADIO_SHR_PHR_DURATION_US    = OT_RADIO_SHR_PHR_LENGTH_BYTES * OT_RADIO_SYMBOLS_PER_OCTET * OT_RADIO_SYMBOL_TIME,
    OT_RADIO_MAX_TURNAROUND_TIME_US = 12 * OT_RADIO_SYMBOL_TIME, // specified max turnaround time
    OT_RADIO_MAX_ACK_WAIT_US        = (OT_RADIO_AIFS_TIME_US + (10 * OT_RADIO_SYMBOL_TIME)),
    OT_RADIO_aMaxSifsFrameSize      = 18, // From 802.15.4 spec - max frame size considered 'short'
};

// Wi-Fi 802.11n related parameters. See radio-parameters.h for radio-model-specific Wi-Fi parameters.
enum
{
    OT_RADIO_WIFI_MAX_TXTIME_US =
        5484,                        // https://nl.mathworks.com/help/wlan/gs/packet-size-and-duration-dependencies.html
    OT_RADIO_WIFI_SLOT_TIME_US = 9,  // https://en.wikipedia.org/wiki/DCF_Interframe_Space
    OT_RADIO_WIFI_CCA_TIME_US  = 28, // https://en.wikipedia.org/wiki/DCF_Interframe_Space
    OT_RADIO_WIFI_CWMIN_SLOTS  = 32, // https://wiki.dd-wrt.com/wiki/index.php/WMM
};

OT_TOOL_PACKED_BEGIN
struct RadioMessage
{
    uint8_t mChannel;
    uint8_t mPsdu[OT_RADIO_FRAME_MAX_SIZE];
} OT_TOOL_PACKED_END;

typedef enum
{
    RFSIM_PARAM_RX_SENSITIVITY,
    RFSIM_PARAM_CCA_THRESHOLD,
    RFSIM_PARAM_CSL_ACCURACY,
    RFSIM_PARAM_CSL_UNCERTAINTY,
    RFSIM_PARAM_TX_INTERFERER,
    RFSIM_PARAM_CLOCK_DRIFT,
    RFSIM_PARAM_PHY_BITRATE,
    RFSIM_PARAM_UNKNOWN = 255,
} RfSimParam;

/**
 * The sub-states of the virtual-time simulated radio. Sub-states are shared between
 * all OT radio states.
 */
typedef enum
{
    RFSIM_RADIO_SUBSTATE_READY,
    RFSIM_RADIO_SUBSTATE_IFS_WAIT,
    RFSIM_RADIO_SUBSTATE_TX_CCA,
    RFSIM_RADIO_SUBSTATE_TX_CCA_TO_TX,
    RFSIM_RADIO_SUBSTATE_TX_FRAME_ONGOING,
    RFSIM_RADIO_SUBSTATE_TX_TX_TO_RX,
    RFSIM_RADIO_SUBSTATE_TX_TX_TO_AIFS,
    RFSIM_RADIO_SUBSTATE_TX_AIFS_WAIT,
    RFSIM_RADIO_SUBSTATE_TX_ACK_RX_ONGOING,
    RFSIM_RADIO_SUBSTATE_RX_FRAME_ONGOING,
    RFSIM_RADIO_SUBSTATE_RX_AIFS_WAIT,
    RFSIM_RADIO_SUBSTATE_RX_ACK_TX_ONGOING,
    RFSIM_RADIO_SUBSTATE_RX_TX_TO_RX,
    RFSIM_RADIO_SUBSTATE_RX_ENERGY_SCAN,
    RFSIM_RADIO_SUBSTATE_STARTUP,
    RFSIM_RADIO_SUBSTATE_INVALID,
    RFSIM_RADIO_SUBSTATE_AWAIT_CCA,
    RFSIM_RADIO_SUBSTATE_CW_BACKOFF,
} RadioSubState;

#endif // PLATFORM_RFSIM_RADIO_H
