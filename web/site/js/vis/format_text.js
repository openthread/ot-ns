// Copyright (c) 2020-2023, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

const {OtDeviceRole} = require("../proto/visualize_grpc_pb");

export function formatRloc16(rloc16) {
    return ('0000' + rloc16.toString(16).toUpperCase()).slice(-4);
}

export function formatExtAddr(extAddr) {
    return ('0000000000000000' + extAddr.toString(16).toUpperCase()).slice(-16);
}

export function formatPartitionId(parid) {
    return ('00000000' + parid.toString(16).toUpperCase()).slice(-8);
}

export function spacePad(value, numChars) {
    return value.toString().padStart(numChars)
}

export function roleToString(role) {
    switch (role) {
        case OtDeviceRole.OT_DEVICE_ROLE_DISABLED:
            return "Disabled";
        case OtDeviceRole.OT_DEVICE_ROLE_DETACHED:
            return "Detached";
        case OtDeviceRole.OT_DEVICE_ROLE_CHILD:
            return "Child";
        case OtDeviceRole.OT_DEVICE_ROLE_ROUTER:
            return "Router";
        case OtDeviceRole.OT_DEVICE_ROLE_LEADER:
            return "Leader";
    }
}

export function modeToString(mode) {
    if (mode == undefined) {
        return "undefined";
    }
    let s = "";
    if (mode.getRxOnWhenIdle()) {
        s += "r";
    }
    if (mode.getFullThreadDevice()) {
        s += "d";
    }
    if (mode.getFullNetworkData()) {
        s += "n";
    }
    return s;
}

export function threadVersionToString(ver) {
    switch(ver) {
        case 0:
            return "INVALID";
        case 1:
            return "1.0.x";
        case 2:
            return "1.1.x";
        case 3:
            return "1.2.x";
        case 4:
            return "1.3.x";
        case 5:
            return "1.4.x";
        case 6:
            return "1.5.x?";
        case 7:
            return "1.6.x?";
        default:
            return "UNKNOWN";
    }
}