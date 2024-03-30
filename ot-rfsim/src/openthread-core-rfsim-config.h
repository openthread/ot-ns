/*
 *  Copyright (c) 2017-2024, The OpenThread Authors.
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
 *   This file includes the RFSIM platform's compile-time configuration constants
 *   for OpenThread. (All OPENTHREAD_CONFIG_* defines.) It is based on the simulated
 *   radio constants that are found in "radio-parameters.h".
 */

#ifndef OPENTHREAD_CORE_RFSIM_CONFIG_H_
#define OPENTHREAD_CORE_RFSIM_CONFIG_H_

#include "radio-parameters.h"

#ifndef OPENTHREAD_RADIO
#define OPENTHREAD_RADIO 0
#endif

/**
 * @def OPENTHREAD_CONFIG_PLATFORM_INFO
 *
 * The platform-specific string to insert into the OpenThread version string.
 *
 */
#define OPENTHREAD_CONFIG_PLATFORM_INFO "RFSIM"

// In older codebases, some required defines are missing - recreate them here.
#ifndef OT_THREAD_VERSION_1_3
#define OT_THREAD_VERSION_1_3 4
#endif

#ifndef OT_THREAD_VERSION_1_3_1
#define OT_THREAD_VERSION_1_3_1 5
#endif

// Helper macros for vendor SW version string
#define STRINGIZE_HELPER(x) #x
#define STRINGIZE(x) STRINGIZE_HELPER(x)

#ifndef OPENTHREAD_CONFIG_NET_DIAG_VENDOR_SW_VERSION
#define OPENTHREAD_CONFIG_NET_DIAG_VENDOR_SW_VERSION RFSIM_SW_VERSION "/Tv" STRINGIZE(OPENTHREAD_CONFIG_THREAD_VERSION)
#endif

/**
 * @def OPENTHREAD_CONFIG_LOG_OUTPUT
 *
 * Specify where the log output should go.
 *
 */
#ifndef OPENTHREAD_CONFIG_LOG_OUTPUT /* allows command line override */
#define OPENTHREAD_CONFIG_LOG_OUTPUT OPENTHREAD_CONFIG_LOG_OUTPUT_PLATFORM_DEFINED
#endif

#ifndef OPENTHREAD_CONFIG_IP6_SLAAC_ENABLE
#define OPENTHREAD_CONFIG_IP6_SLAAC_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_IP6_MAX_EXT_MCAST_ADDRS
 *
 * Configure a much higher number of externally-configured IPv6 multicast addresses than the default (2).
 * This is done to support OTNS2 'send' command for multicast traffic generation.
 */
#ifndef OPENTHREAD_CONFIG_IP6_MAX_EXT_MCAST_ADDRS
#define OPENTHREAD_CONFIG_IP6_MAX_EXT_MCAST_ADDRS 1024
#endif

#ifndef OPENTHREAD_CONFIG_SRP_SERVER_ENABLE
#define OPENTHREAD_CONFIG_SRP_SERVER_ENABLE 0
#endif

#ifndef OPENTHREAD_CONFIG_SRP_CLIENT_ENABLE
#define OPENTHREAD_CONFIG_SRP_CLIENT_ENABLE (OPENTHREAD_CONFIG_THREAD_VERSION >= OT_THREAD_VERSION_1_3)
#endif

#ifndef OPENTHREAD_CONFIG_DNS_CLIENT_ENABLE
#define OPENTHREAD_CONFIG_DNS_CLIENT_ENABLE (OPENTHREAD_CONFIG_THREAD_VERSION >= OT_THREAD_VERSION_1_3)
#define OPENTHREAD_CONFIG_DNS_CLIENT_OVER_TCP_ENABLE (OPENTHREAD_CONFIG_THREAD_VERSION >= OT_THREAD_VERSION_1_3)
#endif

#ifndef OPENTHREAD_CONFIG_MAC_CSL_RECEIVER_ENABLE
#define OPENTHREAD_CONFIG_MAC_CSL_RECEIVER_ENABLE (OPENTHREAD_CONFIG_THREAD_VERSION >= OT_THREAD_VERSION_1_2)
#endif

/**
 * @def OPENTHREAD_CONFIG_CSL_RECEIVE_TIME_AHEAD
 *
 * Define radio wake-up time from Sleep to Rx for CSL receiver, in microseconds.
 */
#ifndef OPENTHREAD_CONFIG_CSL_RECEIVE_TIME_AHEAD
#define OPENTHREAD_CONFIG_CSL_RECEIVE_TIME_AHEAD RFSIM_RAMPUP_TIME_US
#endif

/**
 * @def OPENTHREAD_CONFIG_CSL_TRANSMIT_TIME_AHEAD
 *
 * Transmission scheduling and ramp up time needed for the CSL transmitter to be ready, in units of microseconds.
 * This assumes the CSL transmitter goes from Rx (normal) state to Tx state as fast as it can, after having
 * performed the CCA.
 */
#ifndef OPENTHREAD_CONFIG_CSL_TRANSMIT_TIME_AHEAD
#define OPENTHREAD_CONFIG_CSL_TRANSMIT_TIME_AHEAD RFSIM_TURNAROUND_TIME_US
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_CSL_REQUEST_AHEAD_US
 *
 * Define minimum schedule-request ahead time for Tx of CSL transmitter, in microseconds.
 * Includes radio startup Sleep->Tx, SHR time duration, CCA time duration, and Turnaround time.
 */
#ifndef OPENTHREAD_CONFIG_MAC_CSL_REQUEST_AHEAD_US
#define OPENTHREAD_CONFIG_MAC_CSL_REQUEST_AHEAD_US (RFSIM_RAMPUP_TIME_US + 5*32 + 128 + RFSIM_TURNAROUND_TIME_US)
#endif

/**
 * @def OPENTHREAD_CONFIG_MIN_RECEIVE_ON_AFTER
 *
 * The minimum time (in microseconds) after the MHR start that the radio should
 * be in receive state and NOT switched back to sleep state (using
 * `otPlatRadioSleep()`).
 *
 * For the RFSIM platform, it is currently set at 160 us, to include the variable
 * 0-159 quantization error due to CSL phase report in CSL IEs. That's because
 * due to the quant error, the CSL Transmitter may be up to 159 us "late" in
 * transmitting the frame during the CSL sampling window.
 */
#ifndef OPENTHREAD_CONFIG_MIN_RECEIVE_ON_AFTER
#define OPENTHREAD_CONFIG_MIN_RECEIVE_ON_AFTER 160
#endif

#ifndef OPENTHREAD_CONFIG_BORDER_ROUTER_ENABLE
#define OPENTHREAD_CONFIG_BORDER_ROUTER_ENABLE (OPENTHREAD_CONFIG_THREAD_VERSION >= OT_THREAD_VERSION_1_3)
#endif

#ifndef OPENTHREAD_CONFIG_OPERATIONAL_DATASET_AUTO_INIT
#define OPENTHREAD_CONFIG_OPERATIONAL_DATASET_AUTO_INIT 1
#endif

#if OPENTHREAD_RADIO
/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_ACK_TIMEOUT_ENABLE
 *
 * Define to 1 if you want to enable software ACK timeout logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_ACK_TIMEOUT_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_ACK_TIMEOUT_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_ENERGY_SCAN_ENABLE
 *
 * Define to 1 if you want to enable software energy scanning logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_ENERGY_SCAN_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_ENERGY_SCAN_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_RETRANSMIT_ENABLE
 *
 * Define to 1 if you want to enable software retransmission logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_RETRANSMIT_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_RETRANSMIT_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_CSMA_BACKOFF_ENABLE
 *
 * Define to 1 if you want to enable software CSMA-CA backoff logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_CSMA_BACKOFF_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_CSMA_BACKOFF_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_SECURITY_ENABLE
 *
 * Define to 1 if you want to enable software transmission security logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_SECURITY_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_SECURITY_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_TIMING_ENABLE
 *
 * Define to 1 to enable software transmission target time logic.
 *
 */
#ifndef OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_TIMING_ENABLE
#define OPENTHREAD_CONFIG_MAC_SOFTWARE_TX_TIMING_ENABLE 1
#endif
#endif // OPENTHREAD_RADIO

/**
 * @def OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE
 *
 * Define to 1 if you want to support microsecond timer in platform.
 *
 */
#define OPENTHREAD_CONFIG_PLATFORM_USEC_TIMER_ENABLE 1

/**
 * @def OPENTHREAD_CONFIG_PLATFORM_FLASH_API_ENABLE
 *
 * Define to 1 to enable otPlatFlash* APIs to support non-volatile storage.
 *
 * When defined to 1, the platform MUST implement the otPlatFlash* APIs instead of the otPlatSettings* APIs.
 *
 */
#define OPENTHREAD_CONFIG_PLATFORM_FLASH_API_ENABLE 1

/**
 * @def CLI_COAP_SECURE_USE_COAP_DEFAULT_HANDLER
 *
 * Define to 1 to use DefaultHandler for unhandled requests
 *
 */
#define CLI_COAP_SECURE_USE_COAP_DEFAULT_HANDLER 1

/**
 * @def OPENTHREAD_CONFIG_LOG_PLATFORM
 *
 * Define to enable platform region logging.
 *
 */
#ifndef OPENTHREAD_CONFIG_LOG_PLATFORM
#define OPENTHREAD_CONFIG_LOG_PLATFORM 1
#endif

/**
 * @def OPENTHREAD_CONFIG_CLI_MAX_LINE_LENGTH
 *
 * The maximum size of the CLI line in bytes.
 *
 */
#ifndef OPENTHREAD_CONFIG_CLI_MAX_LINE_LENGTH
#define OPENTHREAD_CONFIG_CLI_MAX_LINE_LENGTH 640
#endif

/**
 * @def OPENTHREAD_CONFIG_CLI_UART_RX_BUFFER_SIZE
 *
 * The size of CLI UART RX buffer in bytes.
 *
 */
#ifndef OPENTHREAD_CONFIG_CLI_UART_RX_BUFFER_SIZE
#define OPENTHREAD_CONFIG_CLI_UART_RX_BUFFER_SIZE 640
#endif

/**
 * @def OPENTHREAD_CONFIG_MLE_MAX_CHILDREN
 *
 * The maximum number of children for a Router. Default is the minimum required for a
 * certified device, which is 10.
 *
 */
#ifndef OPENTHREAD_CONFIG_MLE_MAX_CHILDREN
#define OPENTHREAD_CONFIG_MLE_MAX_CHILDREN 10
#endif

/**
 * @def OPENTHREAD_CONFIG_UPTIME_ENABLE
 *
 * Define to 1 to enable tracking the uptime of OpenThread instance.
 *
 */
#ifndef OPENTHREAD_CONFIG_UPTIME_ENABLE
#define OPENTHREAD_CONFIG_UPTIME_ENABLE 1
#endif

/**
 * @def OPENTHREAD_CONFIG_LOG_PREPEND_UPTIME
 *
 * Define as 1 to prepend the current uptime to all log messages.
 *
 */
#ifndef OPENTHREAD_CONFIG_LOG_PREPEND_UPTIME
#define OPENTHREAD_CONFIG_LOG_PREPEND_UPTIME 1
#endif

/**
 * @def OPENTHREAD_PLATFORM_USE_PSEUDO_RESET
 *
 * Define as 1 to enable pseudo-reset.
 *
 */
#ifndef OPENTHREAD_PLATFORM_USE_PSEUDO_RESET
#define OPENTHREAD_PLATFORM_USE_PSEUDO_RESET 1
#endif

/**
 * @def OPENTHREAD_CONFIG_PLATFORM_ASSERT_MANAGEMENT
 *
 * Define as 1 to enable custom log message printing upon OT_ASSERT by the OpenThread stack.
 * Define as 0 to enable the default C library 'assert' function to handle OT_ASSERT.
 */
#ifndef OPENTHREAD_CONFIG_PLATFORM_ASSERT_MANAGEMENT
#define OPENTHREAD_CONFIG_PLATFORM_ASSERT_MANAGEMENT 1
#endif

/**
 * @def OPENTHREAD_CONFIG_OTNS_ENABLE
 *
 * This setting configures OT-NS simulator support. It MUST be enabled
 * for the RFSIM platform.
 *
 */
#define OPENTHREAD_CONFIG_OTNS_ENABLE 1

#endif // OPENTHREAD_CORE_RFSIM_CONFIG_H_
