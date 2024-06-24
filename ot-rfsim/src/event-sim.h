/*
*  Copyright (c) 2020-2024, The OpenThread Authors.
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
*   This file includes simulation-event message definitions, and sending,
*   formatting and parsing functions for events.
*/

#ifndef PLATFORM_RFSIM_EVENT_SIM_H
#define PLATFORM_RFSIM_EVENT_SIM_H

#include "platform-rfsim.h"
#include "radio.h"

/**
 * The event types defined for communication with a simulator and/or with other simulated nodes.
 * Shared for both 'real' and virtual-time event types. Some types are not used in this project
 * (e.g. historic or used by the simulator only.)
 */
enum
{
    OT_SIM_EVENT_ALARM_FIRED         = 0,
    OT_SIM_EVENT_RADIO_RECEIVED      = 1, // legacy
    OT_SIM_EVENT_UART_WRITE          = 2,
    OT_SIM_EVENT_RADIO_SPINEL_WRITE  = 3, // not used?
    OT_SIM_EVENT_POSTCMD             = 4, // not used?
    OT_SIM_EVENT_OTNS_STATUS_PUSH    = 5,
    OT_SIM_EVENT_RADIO_COMM_START    = 6,
    OT_SIM_EVENT_RADIO_TX_DONE       = 7,
    OT_SIM_EVENT_RADIO_CHAN_SAMPLE   = 8,
    OT_SIM_EVENT_RADIO_STATE         = 9,
    OT_SIM_EVENT_RADIO_RX_DONE       = 10,
    OT_SIM_EVENT_EXT_ADDR            = 11,
    OT_SIM_EVENT_NODE_INFO           = 12,
    OT_SIM_EVENT_NODE_DISCONNECTED   = 14, // not used on OT node side
    OT_SIM_EVENT_RADIO_LOG           = 15, // not used on OT node side
    OT_SIM_EVENT_RFSIM_PARAM_GET     = 16,
    OT_SIM_EVENT_RFSIM_PARAM_SET     = 17,
    OT_SIM_EVENT_RFSIM_PARAM_RSP     = 18,
    OT_SIM_EVENT_LOG_WRITE           = 19,
    OT_SIM_EVENT_UDP_TO_HOST         = 20,
    OT_SIM_EVENT_IP6_TO_HOST         = 21,
    OT_SIM_EVENT_UDP_FROM_HOST       = 22,
    OT_SIM_EVENT_IP6_FROM_HOST       = 23,
};

#define OT_EVENT_DATA_MAX_SIZE 2048

OT_TOOL_PACKED_BEGIN
struct EventHeader
{
    uint64_t mDelay;
    uint8_t  mEvent;
    uint64_t mMsgId;
    uint16_t mDataLength;
} OT_TOOL_PACKED_END;

OT_TOOL_PACKED_BEGIN
struct Event
{
    uint64_t mDelay;      // delay in us before execution of the event
    uint8_t  mEvent;      // event type
    uint64_t mMsgId;      // an ever-increasing event message id
    uint16_t mDataLength; // the actual length of following event payload data
    uint8_t  mData[OT_EVENT_DATA_MAX_SIZE];
} OT_TOOL_PACKED_END;

OT_TOOL_PACKED_BEGIN
struct RadioCommEventData
{
    uint8_t  mChannel;    // radio channel number (shared for IEEE 802.15.4 / BLE / ... )
    int8_t   mPower;      // power value (dBm), either RSSI or Tx-power
    uint8_t  mError;      // status code result of radio operation using otError values
    uint64_t mDuration;   // us duration of the radio comm operation
} OT_TOOL_PACKED_END;

OT_TOOL_PACKED_BEGIN
struct RadioStateEventData
{
    uint8_t  mChannel;       // radio channel (see above comments)
    int8_t   mTxPower;       // only valid when mEnergyState == OT_RADIO_STATE_TRANSMIT
    int8_t   mRxSensitivity; // current RX sensitivity in dBm
    uint8_t  mEnergyState;   // energy-state of radio (disabled, sleep, actively Tx, actively Rx)
    uint8_t  mSubState;      // detailed substate of radio, see enum RadioSubState
    uint8_t  mState;         // OT state of radio (disabled, sleep, Tx, Rx)
    uint64_t mRadioTime;     // the radio's time otPlatRadioGetNow()
} OT_TOOL_PACKED_END;

OT_TOOL_PACKED_BEGIN
struct RfSimParamEventData
{
    uint8_t mParam;
    int32_t mValue;
} OT_TOOL_PACKED_END;

OT_TOOL_PACKED_BEGIN
struct MsgToHostEventData
{
    uint16_t mSrcPort;
    uint16_t mDstPort;
    uint8_t  mSrcIp6[OT_IP6_ADDRESS_SIZE];
    uint8_t  mDstIp6[OT_IP6_ADDRESS_SIZE];
} OT_TOOL_PACKED_END;

/**
 * Send a generic simulation event to the simulator. Event fields are
 * updated to the values that were used for sending the event.
 *
 * @param[in,out]   aEvent  A pointer to the simulation event to update, and send.
 */
void otSimSendEvent(struct Event *aEvent);

/**
 * Send a sleep event to the simulator. The amount of time to sleep
 * for this node is determined by the alarm timer, by calling platformAlarmGetNext().
 *
 */
void otSimSendSleepEvent(void);

/**
 * Sends a RadioComm (Tx) simulation event to the simulator.
 *
 * @param[in]       aEventData A pointer to specific data for RadioComm event.
 * @param[in]       aPayload     A pointer to the data payload (radio frame) to send.
 * @param[in]       aLenPayload  Length of aPayload data.
 */
void otSimSendRadioCommEvent(struct RadioCommEventData *aEventData,  const uint8_t *aPayload, size_t aLenPayload);

/**
 * Sends a RadioComm (Tx) simulation event to the simulator for transmitting non-802.15.4
 * interference signals.
 *
 * @param[in] aEventData A pointer to specific data for the event.
 */
void otSimSendRadioCommInterferenceEvent(struct RadioCommEventData *aEventData);

/**
 * Send a Radio State simulation event to the simulator. It reports radio state
 * and indicates for how long the current radio-state will last until next state-change.
 *
 * @param[in]  aStateData                 A pointer to specific data for Radio State event.
 * @param[in]  aDeltaUntilNextRadioState  Time (us) until next radio-state change event, or UNDEFINED_TIME_US.
 */
void otSimSendRadioStateEvent(struct RadioStateEventData *aStateData, uint64_t aDeltaUntilNextRadioState);

/**
 * Send a channel sample simulation event to the simulator. It is used both
 * for CCA and energy scanning on channels.
 *
 * @param[in]  aChanData    A pointer to channel-sample data instructing what to sample.
 */
void otSimSendRadioChanSampleEvent(struct RadioCommEventData *aChanData);

/**
 * Send a UART data event to the simulator.
 *
 * @param[in]   aData       A pointer to the UART data.
 * @param[in]   aLength     Length of UART data.
 */
void otSimSendUartWriteEvent(const uint8_t *aData, uint16_t aLength);

/**
 * Send an OT log-write event to the simulator, containing a single log item.
 *
 * @param[in]   aData       A pointer to the UART data.
 * @param[in]   aLength     Length of UART data.
 */
void otSimSendLogWriteEvent(const uint8_t *aData, uint16_t aLength);

/**
 * Send status push data event to the OT-NS simulator.
 *
 * @param[in]   aStatus     A pointer to the status string data.
 * @param[in]   aLength     Length of status string data.
 */
void otSimSendOtnsStatusPushEvent(const char *aStatus, uint16_t aLength);

/**
 * Send Extended Address change event to the simulator.
 * It differs from an OTNS Status Push 'extaddr' event in being not
 * encoded as a string, but binary.
 *
 * @param aExtAddress    The (new) Extended Address of the node.
 */
void otSimSendExtAddrEvent(const otExtAddress *aExtAddress);

/**
 * Send OT node information to the simulator. This helps the simulator
 * to identify a new socket connection made by the node.
 *
 * @param nodeId  id of the sending OT node
 */
void otSimSendNodeInfoEvent(uint32_t nodeId);

// TODO
void otSimSendRfSimParamRespEvent(uint8_t param, int32_t value);

/**
 * Send an OT message (e.g. UDP, or IPv6 datagram, etc.) to the simulator to be handled
 * by the "host" of the node. This host could be a local process/script, or an AIL network
 * interface that can further forward the message to its destination.
 *
 * @param evType     the event type to use
 * @param aEventData the event data containing the message's metadata
 * @param aMsgBytes  the bytes of the message itself (e.g. UDP packet bytes or IPv6 datagram bytes)
 * @param aMsgLen    the length of the message
 */
void otSimSendMsgToHostEvent(uint8_t evType, struct MsgToHostEventData *aEventData, uint8_t *aMsgBytes, size_t aMsgLen);

#endif // PLATFORM_RFSIM_EVENT_SIM_H
