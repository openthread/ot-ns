// Copyright (c) 2023, The OTNS Authors.
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

import * as PIXI from "pixi.js-legacy";
import VObject from "./VObject";
import * as fmt from "./format_text"
import {OtDeviceRole} from "../proto/visualize_grpc_pb";
import {POWER_DBM_INVALID} from "./consts";

export default class NodeWindow extends VObject {
    constructor() {
        super();
        this._root = new PIXI.Container();
        this.showNode(0);
    }

    showNode(node){
        let win = document.getElementById('nodeWindow');
        if (!win) {
            return;
        }
        if (!node) {
            win.hidden = true;
            return;
        }
        win.hidden = false;

        let txt        =
            `Node ${node.id} properties\n\n` +
            `RLOC16    : ${fmt.formatRloc16(node.rloc16)}\n` ;

        let rid = `Router ID : --\n`;
        let cid = `Child ID  : --\n`;
        let pid = `Parent    : --\n\n`;
        if(node.role === OtDeviceRole.OT_DEVICE_ROLE_ROUTER || node.role === OtDeviceRole.OT_DEVICE_ROLE_LEADER) {
            rid = `Router ID : ${node.routerId}\n`;
        }else if(node.role === OtDeviceRole.OT_DEVICE_ROLE_CHILD) {
            cid = `Child ID  : ${node.childId}\n`;
            pid = `Parent    : ${fmt.formatExtAddr(node.parent)}\n` +
                  `            (Node ${node.parentId})\n`;
        }

        txt += rid+cid+pid;

        txt += `ExtAddr   : ${fmt.formatExtAddr(node.extAddr)}\n` +
            `Role      : ${fmt.roleToString(node.role)}\n` +
            `Mode      : ${fmt.modeToString(node.nodeMode)}\n` +
            `Partition : ${fmt.formatPartitionId(node.partition)}\n` +

            `Radio-fail: ${node.failed ? "FAILED (simulated)":"No"}\n` +
            `Position  : (${node.x}, ${node.y})\n` ;

        if (node.txPowerLast != POWER_DBM_INVALID) {
            txt += `Tx-Power  : ${node.txPowerLast} dBm  (last fr)\n`
                +  `Tx-Channel: ${node.channelLast}     (last fr)\n`;
        }

        win.value = txt;
    }
}
