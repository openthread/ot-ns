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

/**
 * @file
 * @brief
 *   This file includes the platform-specific OT system initializers and processing
 *   for the OT-RFSIM simulation platform.
 */

#include "platform-rfsim.h"

#include <errno.h>
#include <libgen.h>
#include <sys/socket.h>
#include <sys/un.h>

#include <openthread/tasklet.h>
#include <openthread/udp.h>

#include "common/debug.hpp"

extern void platformReceiveEvent(otInstance *aInstance);
extern bool gPlatformPseudoResetWasRequested;

static void socket_init(char *socketFilePath);
static void handleSignal(int aSignal);

volatile bool   gTerminate          = false;
uint32_t        gNodeId             = 0;
int             gSockFd             = 0;
static uint16_t sIsInstanceInitDone = false;

void otSysInit(int argc, char *argv[])
{
    char   *endptr;
    int32_t randomSeed = 0;

    if (gPlatformPseudoResetWasRequested)
    {
        gPlatformPseudoResetWasRequested = false;
        return;
    }

    signal(SIGTERM, &handleSignal);
    signal(SIGHUP, &handleSignal);

    if (argc < 3 || argc > 4)
    {
        fprintf(stderr, "Usage: %s <NodeId> <OTNS-Unix-socket-file> [<random-seed>]\n", basename(argv[0]));
        platformExit(EXIT_FAILURE);
    }

    long nodeIdParam = strtol(argv[1], &endptr, 0);
    if (*endptr != '\0' || nodeIdParam < 1 || nodeIdParam >= UINT32_MAX)
    {
        fprintf(stderr, "Invalid NodeId: %s (must be >= 1 and < UINT32_MAX )\n", argv[1]);
        platformExit(EXIT_FAILURE);
    }
    gNodeId = (uint32_t)nodeIdParam;

    if (argc == 4)
    {
        long randomSeedParam = strtol(argv[3], &endptr, 0);
        if (*endptr != '\0' || randomSeedParam >= INT32_MAX || randomSeedParam <= INT32_MIN)
        {
            fprintf(stderr, "Invalid random-seed integer: %s (must be > INT32_MIN and < INT32_MAX)\n", argv[3]);
            platformExit(EXIT_FAILURE);
        }
        randomSeed = (int32_t)randomSeedParam;
    }

    platformLoggingInit(argv[0]);
    platformRandomInit(randomSeed);
    socket_init(argv[2]);
    platformAlarmInit();
    platformRadioInit();
    platformRfsimInit();

    otSimSendNodeInfoEvent(gNodeId);
}

bool otSysPseudoResetWasRequested(void) { return gPlatformPseudoResetWasRequested; }

void otSysDeinit(void)
{
    close(gSockFd);
    gSockFd = 0;
}

void otSysProcessDrivers(otInstance *aInstance)
{
    fd_set read_fds;
    fd_set write_fds;
    fd_set error_fds;
    int    max_fd;
    int    rval;

    if (gTerminate)
    {
        platformExit(EXIT_SUCCESS);
    }

    // on the first call, perform any init that requires the aInstance.
    if (!sIsInstanceInitDone)
    { // TODO move to own function
#if OPENTHREAD_CONFIG_UDP_FORWARD_ENABLE && OPENTHREAD_CONFIG_BORDER_ROUTING_ENABLE
        otUdpForwardSetForwarder(aInstance, handleUdpForwarding, aInstance);
#endif
        platformNetifSetUp(aInstance);
        sIsInstanceInitDone = true;
    }

    FD_ZERO(&read_fds);
    FD_ZERO(&write_fds);
    FD_ZERO(&error_fds);

    FD_SET(gSockFd, &read_fds);
    max_fd = gSockFd;

    if (!otTaskletsArePending(aInstance) && platformAlarmGetNext() > 0 &&
        (!platformRadioIsTransmitPending() || platformRadioIsBusy()))
    {
        // report my final radio state at end of this time instant, then go to sleep.
        platformRadioReportStateToSimulator(false);
        otSimSendSleepEvent();

        // wake up by reception of socket event from simulator.
        rval = select(max_fd + 1, &read_fds, &write_fds, &error_fds, NULL);

        if ((rval < 0) && (errno != EINTR))
        {
            perror("select");
            platformExit(EXIT_FAILURE);
        }

        if (rval > 0 && FD_ISSET(gSockFd, &read_fds))
        {
            platformReceiveEvent(aInstance);
        }
    }

    platformAlarmProcess(aInstance);
    platformRadioProcess(aInstance);
    platformRadioInterfererProcess(aInstance);
#if OPENTHREAD_CONFIG_BLE_TCAT_ENABLE
    platformBleProcess(aInstance);
#endif
}

/**
 * Initialises the client socket used for communication with the
 * simulator. The port number is calculated based on environment vars (if set)
 * or else defaults.
 */
static void socket_init(char *socketFilePath)
{
    struct sockaddr_un sockaddr;
    memset(&sockaddr, 0, sizeof(struct sockaddr_un));
    sockaddr.sun_family = AF_UNIX;
    size_t strLen       = strlen(socketFilePath);
    OT_ASSERT(strLen < sizeof(sockaddr.sun_path));
    memcpy(sockaddr.sun_path, socketFilePath, strLen);

    gSockFd = socket(AF_UNIX, SOCK_STREAM, 0);

    if (gSockFd == -1)
    {
        perror("socket");
        platformExit(EXIT_FAILURE);
    }

    if (connect(gSockFd, (struct sockaddr *)&sockaddr, sizeof(sockaddr)) == -1)
    {
        gTerminate = true;
        fprintf(stderr, "Unable to open Unix socket to OT-NS at: %s\n", sockaddr.sun_path);
        perror("bind");
        platformExit(EXIT_FAILURE);
    }
}

static void handleSignal(int aSignal)
{
    OT_UNUSED_VARIABLE(aSignal);

    gTerminate = true;
}
