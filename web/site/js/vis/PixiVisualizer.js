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

import * as PIXI from "pixi.js-legacy";
import VObject from "./VObject";
import ActionBar from "./ActionBar";
import {Text} from "./wrapper";
import {
    FRAME_CONTROL_MASK_FRAME_TYPE, FRAME_TYPE_ACK, LOG_WINDOW_FONT_COLOR, MAX_SPEED, PAUSE_SPEED, STATUS_MSG_FONT_FAMILY,
    STATUS_MSG_FONT_SIZE, NODE_ID_INVALID
} from "./consts";
import Node from "./Node"
import {AckMessage, BroadcastMessage, UnicastMessage} from "./message";
import LogWindow, {LOG_WINDOW_WIDTH} from "./LogWindow";
import * as fmt from "./format_text"
import NodeWindow from "./NodeWindow";

const {
    OtDeviceRole, CommandRequest
} = require('../proto/visualize_grpc_pb.js');

var vis = null;
let ticker = PIXI.Ticker.shared;

export function Visualizer() {
    return vis;
}

export default class PixiVisualizer extends VObject {
    constructor(app, grpcServiceClient) {
        super();
        vis = this;

        this.app = app;
        this.grpcServiceClient = grpcServiceClient;
        this.speed = 1;
        this.curTime = 0;
        this.curSpeed = 1;
        this.nodes = {};
        this._messages = {};
        this.newNodePos = null;

        this.root = new PIXI.Container();
        // this.root.width =
        this.root.position.set(0, 20);
        this.root.interactive = true;
        this.root.hitArea = new PIXI.Rectangle(0, 0, 3000, 3000);
        app.stage.addChild(this.root);

        this.setOnTap((e) => {
            this.onTapedStage()
        });

        this.nodeLogColor = {};

        this._bgStage = new PIXI.Container();
        this.addChild(this._bgStage);

        this._broadcastMessagesStage = new PIXI.Container();
        this.addChild(this._broadcastMessagesStage);

        this._logWindowStage = new PIXI.Container();
        this.addChild(this._logWindowStage);

        this._nodesStage = new PIXI.Container();
        this.addChild(this._nodesStage);

        this._unicastMessagesStage = new PIXI.Container();
        this.addChild(this._unicastMessagesStage);

        this.statusMsg = new PIXI.Text("", {
            fontFamily: STATUS_MSG_FONT_FAMILY,
            fontSize: STATUS_MSG_FONT_SIZE,
            fontWeight: "bold"
        });
        this.statusMsg.position.set(0, -this.statusMsg.height);
        this.addChild(this.statusMsg);

        this.actionBar = new ActionBar();
        this.addChild(this.actionBar);
        this.actionBar.position.set(10, 1000);
        this.actionBar.setDraggable();
        this.updateStatusMsg();

        this.nodeWindow = new NodeWindow();
        this.addChild(this.nodeWindow);

        this.otVersion = "";
        this.otCommit = "";
        this.otCommitIdMsg = new Text("OpenThread Version: ", {
            fill: "#0052ff",
            fontFamily: "Verdana",
            fontSize: 13,
            fontStyle: "italic",
            fontWeight: "bolder"
        });
        this.otCommitIdMsg.position.set(this.statusMsg.x, this.statusMsg.y + this.statusMsg.height + 3);
        this.otCommitIdMsg.interactive = true;
        this.otCommitIdMsg.setOnTap((e) => {
            window.open('https://github.com/openthread/openthread/commit/' + this.otCommit, '_blank');
            e.stopPropagation();
        });
        this.addChild(this.otCommitIdMsg);
        this.setOTVersion("", "main");

        this.titleText = new PIXI.Text("", {
            fill: "#e69900",
            fontFamily: "Verdana",
            fontSize: 20,
            fontWeight: "bolder"
        });
        this.titleText.position.set(0, 20);
        this.addChild(this.titleText);

        this.real = false;
        this._applyReal();
        this._resetIdleCheckTimer()
    }

    update(dt) {
        super.update(dt);

        this._drawNodeLinks();

        for (let id in this.nodes) {
            let node = this.nodes[id];
            node.update(dt)
        }

        for (let id in this._messages) {
            let msg = this._messages[id];
            msg.update(dt)
        }
    }

    showLogWindow() {
        if (!this.logWindow) {
            this.logWindow = new LogWindow();
            this._logWindowStage.addChild(this.logWindow._root);
            this._resetLogWindowPosition(window.screen.width, window.screen.height);

            this.log("Log window opened.")
        }
    }

    hideLogWindow() {
        if (this.logWindow) {
            this._logWindowStage.removeChild(this.logWindow._root);
            this.logWindow = null
        }
    }

    clearLogWindow() {
        if (this.logWindow) {
            this.logWindow.clear()
        }
    }

    _resetLogWindowPosition(width, height) {
        if (this.logWindow) {
            this.logWindow.position.set(width - LOG_WINDOW_WIDTH, 10);
            this.logWindow.resetLayout(width, height)
        }
    }

    _resetIdleCheckTimer() {
        if (this._idleCheckTimer) {
            this.cancelCallback(this._idleCheckTimer);
            delete this._idleCheckTimer
        }

        this._idleCheckTimer = this.addCallback(10, () => {
            console.error("idle timer fired, reloading ...");
            location.reload()
        })
    }

    visAdvanceTime(ts, speed) {
        this.curTime = ts;
        this.curSpeed = speed;
        this._resetIdleCheckTimer();
        this.updateStatusMsg()
    }

    visHeartbeat() {
        this._resetIdleCheckTimer()
    }

    visStopIdleCheckTimer() {
        if (this._idleCheckTimer) {
            this.cancelCallback(this._idleCheckTimer);
            delete this._idleCheckTimer
        }
    }

    setOTVersion(version, commit) {
        this.otVersion = version;
        this.otCommit = commit;
        this.otCommitIdMsg.text = "OpenThread Version: " + version + " (" + commit + ")";
    }

    setReal(real) {
        if (this.real === real) {
            return;
        }
        this.real = real;
        this._applyReal();
        this.log(`Real devices: ${real ? "ON" : "OFF"}`);
    }

    _applyReal() {
        if (this.real) {
            this.actionBar.setAbilities({})
        } else {
            this.actionBar.setAbilities({
                "speed": true,
                "add": true,
                "del": true,
                "radio": true,
            })
        }
    }

    updateStatusMsg() {
        this.statusMsg.text = "OTNS-Web | FPS=" + Math.round(ticker.FPS).toString().padStart(3, " ") + " | "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_LEADER) + " leaders "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_ROUTER) + " routers "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_CHILD) + " EDs "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_DETACHED) + " detached"
            + " | SPEED=" + this.formatSpeed()
            + " | TIME=" + this.formatTime();
    }

    getNodeCountByRole(role) {
        let count = 0;
        for (let nodeid in this.nodes) {
            let node = this.nodes[nodeid];
            if (node.role === role) {
                count += 1
            }
        }
        return count
    }

    log(text, color = LOG_WINDOW_FONT_COLOR) {
        if (this.logWindow) {
            this.logWindow.addLog(text, color)
        }
    }

    visAddNode(nodeId, x, y, radioRange) {
        let node = new Node(nodeId, x, y, radioRange);
        this.nodes[nodeId] = node;
        this._nodesStage.addChild(node._root);
        this.setSelectedNode(nodeId);

        let msg = `Added at (${x},${y})`;
        if (!this.real) {
            msg += `, radio range ${radioRange}`
        }
        this.logNode(nodeId, msg)
        this.onNodeUpdate(nodeId);
    }

    visSetNodeRloc16(nodeId, rloc16) {
        let node = this.nodes[nodeId];
        let oldRloc16 = node.rloc16;
        node.setRloc16(rloc16);
        if (oldRloc16 != rloc16) {
            this.logNode(nodeId, `RLOC16 changed from ${fmt.formatRloc16(oldRloc16)} to ${fmt.formatRloc16(rloc16)}`)
            this.onNodeUpdate(nodeId);
        }
    }

    visSetNodeRole(nodeId, role) {
        let oldRole = this.nodes[nodeId].role;
        this.nodes[nodeId].setRole(role);
        if (oldRole != role) {
            this.logNode(nodeId, `Role changed from ${fmt.roleToString(oldRole)} to ${fmt.roleToString(role)}`)
            this.onNodeUpdate(nodeId);
        }
    }

    visSetNodeMode(nodeId, mode) {
        let oldMode = this.nodes[nodeId].nodeMode;
        this.nodes[nodeId].setMode(mode);
        if (oldMode != mode) {
            this.logNode(nodeId, `Mode changed from ${fmt.modeToString(oldMode)} to ${fmt.modeToString(mode)}`);
            this.onNodeUpdate(nodeId);
        }
    }

    visSetNetworkInfo(version, commit, real) {
        let oldVersion = this.otVersion;
        let oldCommit = this.otCommit;
        this.setOTVersion(version, commit);
        this.setReal(real);

        if (oldVersion != this.otVersion) {
            this.log(`OpenThread Version: ${version}`);
        }
        if (oldCommit != this.otCommit) {
            this.log(`OpenThread Commit: ${commit}`);
        }
    }

    visDeleteNode(nodeId) {
        let node = this.nodes[nodeId];
        delete this.nodes[nodeId];
        node.destroy();
        if (nodeId === this._selectedNodeId) {
            this.setSelectedNode(0);
        }
        this.logNode(nodeId, "Deleted")
        this.onNodeUpdate(nodeId);
    }

    visSetSpeed(speed) {
        this.speed = speed;
        this.actionBar.setSpeed(speed);
        this.log(`Speed set to ${speed}`);
        this.updateStatusMsg();
    }

    isPaused() {
        return this.speed <= PAUSE_SPEED
    }

    isMaxSpeed() {
        return this.speed >= MAX_SPEED
    }

    visSetNodePos(nodeId, x, y) {
        this.nodes[nodeId].setPosition(x, y);
        this.logNode(nodeId, `Moved to (${x},${y})`)
        this.onNodeUpdate(nodeId);
    }

    visOnExtAddrChange(nodeId, extAddr) {
        this.nodes[nodeId].extAddr = extAddr;
        this.logNode(nodeId, `Extended Address set to ${fmt.formatExtAddr(extAddr)}`)
        this.onNodeUpdate(nodeId);
    }

    visOnNodeFail(nodeId) {
        this.nodes[nodeId].failed = true;
        this.logNode(nodeId, "Radio is OFF")
        this.onNodeUpdate(nodeId);
    }

    visOnNodeRecover(nodeId) {
        this.nodes[nodeId].failed = false;
        this.logNode(nodeId, "Radio is ON")
        this.onNodeUpdate(nodeId);
    }

    visSetParent(nodeId, extAddr) {
        let parent = this.findNodeByExtAddr(extAddr);
        this.nodes[nodeId].parent = extAddr;
        if (parent) {
            this.nodes[nodeId].parentId = parent.id;
        }else {
            this.nodes[nodeId].parentId = NODE_ID_INVALID;
        }
        this.logNode(nodeId, `Parent set to ${this.formatExtAddrPretty(extAddr)}`)
        this.onNodeUpdate(nodeId);
    }

    visSetTitle(title, x, y, fontSize) {
        let oldTitleText = this.titleText.text;
        this.titleText.text = title;
        this.titleText.x = x;
        this.titleText.y = y;
        this.titleText.style.fontSize = fontSize;

        if (oldTitleText !== title) {
            this.log(`Title set to "${title}", position (${x},${y}), font size ${fontSize}`);
        }
    }

    visSend(srcId, dstId, mvInfo) {
        if (document.visibilityState !== "visible") {
            return;
        }

        let src = this.nodes[srcId];
        if (src == null) return;

        let frameType = mvInfo.getFrameControl() & FRAME_CONTROL_MASK_FRAME_TYPE;
        if (frameType === FRAME_TYPE_ACK) {
            // ACK
            this.createAckMessage(src, mvInfo);
        } else if (dstId == -1) {
            // broadcast
            this.createBroadcastMessage(src, mvInfo);
        } else {
            let dst = this.nodes[dstId];
            this.createUnicastMessage(src, dst, mvInfo);
        }

        if (src.txPowerLast != mvInfo.getPowerDbm() || src.channelLast != mvInfo.getChannel()) {
            src.txPowerLast = mvInfo.getPowerDbm();
            src.channelLast = mvInfo.getChannel();
            this.onNodeUpdate(srcId);
        }
    }

    visSetNodePartitionId(nodeId, partitionId) {
        let oldPartitionId = this.nodes[nodeId].partition;
        this.nodes[nodeId].partition = partitionId;
        this.logNode(nodeId, `Partition changed from ${fmt.formatPartitionId(oldPartitionId)} to ${fmt.formatPartitionId(partitionId)}`)
        this.onNodeUpdate(nodeId);
    }

    visShowDemoLegend(x, y, title) {
        console.log("ShowDemoLegend not implemented")
    }

    visCountDown(durationMs, title) {
        console.log("CountDown not implemented")
    }

    ctrlAddNode(type) {
        if (this.newNodePos == null) {
            this.runCommand("add " + type);
        }else{
            this.runCommand("add " + type + " x " + this.newNodePos.x + " y " + this.newNodePos.y);
            this.newNodePos = null
        }
    }

    ctrlMoveNodeTo(nodeId, x, y, cb) {
        x = Math.floor(x);
        y = Math.floor(y);
        this.runCommand("move " + nodeId + " " + x + " " + y, cb);
    }

    ctrlDeleteNode(nodeId) {
        this.runCommand("del " + nodeId);
    }

    ctrlSetNodeFailed(nodeId, failed) {
        this.runCommand("radio " + nodeId + " " + (failed ? "off" : "on"))
    }

    ctrlSetSpeed(speed) {
        this.runCommand("speed " + speed)
    }

    runCommand(cmd, callback) {
        let req = new CommandRequest();
        req.setCommand(cmd);
        this.log(`> ${cmd}`);
        console.log(`> ${cmd}`);

        this.grpcServiceClient.command(req, {}, (err, resp) => {
                if (err !== null) {
                    this.log("Error: " + err.toLocaleString());
                    console.error("Error: " + err.toLocaleString());
                    if (callback) {
                        callback(err, [])
                    }
                }

                let output = resp.getOutputList();
                for (let i in output) {
                    console.log(output[i]);
                }

                if (callback) {
                    let errmsg = output.pop();

                    if (errmsg !== "Done") {
                        callback(new Error(errmsg), output)
                    } else {
                        callback(null, output)
                    }
                }
            }
        )
    }

    getPartitionColor(parid) {
        if (parid === 0) {
            return 0x000000
        }

        return parid
    }

    setSelectedNode(id) {
        if (id === this._selectedNodeId) {
            return;
        }

        let old_sel = this.nodes[this._selectedNodeId];
        if (old_sel) {
            old_sel.onUnselected();
        }
        delete this._selectedNodeId;

        let new_sel = this.nodes[id];
        if (new_sel) {
            this._selectedNodeId = id;
            new_sel.onSelected();
        }

        this.nodeWindow.showNode(new_sel);

        this.actionBar.setContext(new_sel || "any");
    }

    setSpeed(speed) {
        if (this.real) {
            console.error("setSpeed() not available in real mode");
            return
        }

        this.ctrlSetSpeed(speed)
    }

    deleteSelectedNode() {
        let sel = this.nodes[this._selectedNodeId];
        if (sel) {
            this.ctrlDeleteNode(sel.id)
        }
    }

    setSelectedNodeFailed(failed) {
        let sel = this.nodes[this._selectedNodeId];
        if (!sel) {
            return
        }

        this.ctrlSetNodeFailed(sel.id, failed)
    }

    clearAllNodes() {
        for (let id in this.nodes) {
            this.ctrlDeleteNode(id)
        }
    }

    onTapedStage() {
        this.setSelectedNode(0)
    }

    _drawNodeLinks() {
        let linkLineWidth = 1;
        // this._bgStage.removeChildAt(0)
        this._bgStage.removeChildren().forEach(child => child.destroy());

        const graphics = new PIXI.Graphics();
        graphics.beginFill(0x8bc34a);

        for (let nodeid in this.nodes) {
            let node = this.nodes[nodeid];
            if (node.parent) {
                let parent = this.findNodeByExtAddr(node.parent);
                if (parent !== null) {
                    graphics.moveTo(node.position.x, node.position.y);
                    graphics.lineTo(parent.position.x, parent.position.y)
                }
            }
            for (let extaddr in node._children) {
                let child = this.findNodeByExtAddr(extaddr);
                if (child) {
                    graphics.lineStyle(nodeid == this._selectedNodeId || child.id == this._selectedNodeId ? linkLineWidth * 3 : linkLineWidth, 0x8bc34a, 1);

                    graphics.moveTo(node.position.x, node.position.y);
                    graphics.lineTo(child.position.x, child.position.y)
                }
            }
        }
        graphics.endFill();

        graphics.beginFill(0x1976d2);

        for (let nodeid in this.nodes) {
            graphics.lineStyle(nodeid == this._selectedNodeId ? linkLineWidth * 3 : linkLineWidth, 0x1976d2, 1);

            let node = this.nodes[nodeid];
            for (let extaddr in node._neighbors) {
                let neighbor = this.findNodeByExtAddr(extaddr);
                if (neighbor) {
                    graphics.moveTo(node.position.x, node.position.y);
                    graphics.lineTo(neighbor.position.x, neighbor.position.y)
                }
            }
        }
        graphics.endFill();
        this._bgStage.addChild(graphics)
    }

    /**
     * find a Node by extended address
     * @param extaddr
     * @returns Node
     */
    findNodeByExtAddr(extaddr) {
        for (let nodeid in this.nodes) {
            let node = this.nodes[nodeid];
            if (node.extAddr == extaddr) {
                return node
            }
        }
        return null
    }

    visAddRouterTable(nodeId, extaddr) {
        this.nodes[nodeId].addRouterTable(extaddr);
        this.logNode(nodeId, `Router table added: ${this.formatExtAddrPretty(extaddr)}`)
        this.onNodeUpdate(nodeId);
    }

    visRemoveRouterTable(nodeId, extaddr) {
        this.nodes[nodeId].removeRouterTable(extaddr);
        this.logNode(nodeId, `Router table removed: ${this.formatExtAddrPretty(extaddr)}`)
        this.onNodeUpdate(nodeId);
    }

    visAddChildTable(nodeId, extaddr) {
        this.nodes[nodeId].addChildTable(extaddr);
        this.logNode(nodeId, `Child table added: ${this.formatExtAddrPretty(extaddr)}`)
        let child = this.findNodeByExtAddr(extaddr);
        if (child && this.nodes[child.id]) {
            let extAddrParent = this.nodes[nodeId].extAddr;
            this.visSetParent(child.id,extAddrParent); // call from here because 'parent' push event is not emitted by OT.
        }
        this.onNodeUpdate(nodeId);
    }

    visRemoveChildTable(nodeId, extaddr) {
        this.nodes[nodeId].removeChildTable(extaddr);
        this.logNode(nodeId, `Child table removed: ${this.formatExtAddrPretty(extaddr)}`)
        this.onNodeUpdate(nodeId);
    }

    logNode(nodeId, msg) {
        let color = this.nodeLogColor[nodeId];
        if (typeof color == "undefined") {
            color = this.randomColor();
            this.nodeLogColor[nodeId] = color;
        }

        this.log(`Node ${nodeId}: ${msg}`, color)
    }

    onNodeUpdate(nodeId) {
        if(this._selectedNodeId==nodeId && nodeId > 0){
            this.nodeWindow.showNode(this.nodes[nodeId])
        }
    }

    randomColor() {
        let hue = Math.floor(Math.random() * 360);
        let color = `hsl(${hue}deg, 92%, 23%)`;
        return color;
    }

    formatExtAddrPretty(extAddr) {
        let node = this.findNodeByExtAddr(extAddr);
        if (node) {
            return `Node ${node.id}(${fmt.formatExtAddr(extAddr)})`
        } else {
            return fmt.formatExtAddr(extAddr);
        }
    }

    formatTime() {
        let us = this.curTime % 1000;
        let ms = Math.floor((this.curTime % 1000000) / 1000);
        let secs = Math.floor(this.curTime / 1000000);
        let d = Math.floor(secs / 86400);
        secs = secs % 86400;
        let h = Math.floor(secs / 3600);
        secs = secs % 3600;
        let m = Math.floor(secs / 60);
        secs = secs % 60;

        let str = "";
        if (d > 0) {
            str += d + "d";
        }
        str += h + "h" +
            m.toString().padStart(2, "0") + "m" +
            secs.toString().padStart(2, "0") + "s" +
            "  " + ms.toString().padStart(3," ") + "ms" +
            " " + us.toString().padStart(3," ") + "us";
        return str;
    }

    formatSpeed() {
        let s = this.curSpeed;
        if (s >= 0.9995) {
            return this.curSpeed.toFixed(1).toString().padStart(7, " ") + "     ";
        }else if (s >= 0.0009995) {
            return "    " + this.curSpeed.toFixed(3).toString() + "   ";
        }else {
            return "    " + this.curSpeed.toFixed(6).toString();
        }
    }

    createUnicastMessage(src, dst, mvInfo) {
        let msg = new UnicastMessage(src, dst, mvInfo);
        this._unicastMessagesStage.addChild(msg._root);
        this._messages[msg.id] = msg;
    }

    deleteMessage(msg) {
        delete this._messages[msg.id];
        msg._root.destroy()
    }

    createBroadcastMessage(src, mvInfo) {
        let msg = new BroadcastMessage(src, mvInfo);
        this._broadcastMessagesStage.addChild(msg._root);
        this._messages[msg.id] = msg;
    }

    createAckMessage(src, mvInfo) {
        let msg = new AckMessage(src, mvInfo);
        this._unicastMessagesStage.addChild(msg._root);
        this._messages[msg.id] = msg;
    }

    onResize(width, height) {
        console.log("window resized to " + width + "," + height);
        this.actionBar.position.set(10, height - this.actionBar.height - 20 - 10);
        this._resetLogWindowPosition(width, height);
    }

}
