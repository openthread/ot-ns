// Copyright (c) 2020, The OTNS Authors.
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

import * as PIXI from "pixi.js";
import VObject from "./VObject";
import ActionBar from "./ActionBar";
import {Text} from "./wrapper";
import {FRAME_CONTROL_MASK_FRAME_TYPE, FRAME_TYPE_ACK, MAX_SPEED, PAUSE_SPEED} from "./consts";
import Node from "./Node"
import {AckMessage, BroadcastMessage, UnicastMessage} from "./message";

const {
    VisualizeRequest, VisualizeEvent, OtDeviceRole, NodeMode,
    MoveNodeToRequest, DeleteNodeRequest, AddNodeRequest, SetNodeFailedRequest,
    SetSpeedRequest,
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

        this.root = new PIXI.Container();
        // this.root.width =
        this.root.position.set(0, 20);
        this.root.interactive = true;
        this.root.hitArea = new PIXI.Rectangle(0, 0, 3000, 3000);
        app.stage.addChild(this.root);

        this.setOnTap((e) => {
            this.onTapedStage()
        });

        this._bgStage = new PIXI.Container();
        this.addChild(this._bgStage);

        this._broadcastMessagesStage = new PIXI.Container();
        this.addChild(this._broadcastMessagesStage);

        this._nodesStage = new PIXI.Container();
        this.addChild(this._nodesStage);

        this._unicastMessagesStage = new PIXI.Container();
        this.addChild(this._unicastMessagesStage);

        this.statusMsg = new PIXI.Text("", {
            fontFamily: "Verdana",
            fontSize: 13,
            fontWeight: "bolder"
        });
        this.statusMsg.position.set(0, -this.statusMsg.height);
        this.addChild(this.statusMsg);

        this.actionBar = new ActionBar();
        this.addChild(this.actionBar);
        this.actionBar.position.set(10, 1000);
        this.actionBar.setDraggable();
        this.updateStatusMsg();

        this.otCommitIdMsg = new Text("OpenThread Commit: " + this.OT_COMMIT_ID, {
            fill: "#0052ff",
            fontFamily: "Verdana",
            fontSize: 13,
            fontStyle: "italic",
            fontWeight: "bolder"
        });
        this.otCommitIdMsg.position.set(this.statusMsg.x, this.statusMsg.y + this.statusMsg.height + 3);
        this.otCommitIdMsg.interactive = true;
        this.otCommitIdMsg.setOnTap((e) => {
            window.open('https://github.com/openthread/openthread/commit/' + this.OT_COMMIT_ID, '_blank');
            e.stopPropagation();
        });
        this.addChild(this.otCommitIdMsg);

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

    updateStatusMsg() {
        this.statusMsg.text = "OTNS-Web | FPS=" + Math.round(ticker.FPS) + " | "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_LEADER) + " leaders "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_ROUTER) + " routers "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_CHILD) + " EDs "
            + this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_DETACHED) + " detached"
            + " | SPEED=" + Math.round(this.curSpeed * 10) / 10 + " | TIME=" + this.formatTime();
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

    visAddNode(nodeId, x, y, radioRange, nodeMode) {
        let node = new Node(nodeId, x, y, radioRange, nodeMode);
        this.nodes[nodeId] = node;
        this._nodesStage.addChild(node._root);
        this.setSelectedNode(nodeId)
    }

    visSetNodeRloc16(nodeId, rloc16) {
        this.nodes[nodeId].setRloc16(rloc16)
    }

    visSetNodeRole(nodeId, role) {
        this.nodes[nodeId].setRole(role)
    }

    visDeleteNode(nodeId) {
        let node = this.nodes[nodeId];
        delete this.nodes[nodeId];
        node.destroy();
        if (nodeId === this._selectedNodeId) {
            this.setSelectedNode(0)
        }
    }

    visSetSpeed(speed) {
        this.speed = speed;
        this.actionBar.setSpeed(speed)
    }

    isPaused() {
        return this.speed <= PAUSE_SPEED
    }

    isMaxSpeed() {
        return this.speed >= MAX_SPEED
    }

    visSetNodePos(nodeId, x, y) {
        this.nodes[nodeId].setPosition(x, y)
    }

    visOnExtAddrChange(nodeId, extAddr) {
        this.nodes[nodeId].extAddr = extAddr
    }

    visOnNodeFail(nodeId) {
        this.nodes[nodeId].failed = true
    }

    visOnNodeRecover(nodeId) {
        this.nodes[nodeId].failed = false
    }

    visSetParent(nodeId, extAddr) {
        this.nodes[nodeId].parent = extAddr
    }

    visSend(srcId, dstId, mvInfo) {
        if (document.visibilityState !== "visible") {
            return
        }

        let src = this.nodes[srcId];

        let frameType = mvInfo.getFrameControl() & FRAME_CONTROL_MASK_FRAME_TYPE;
        if (frameType === FRAME_TYPE_ACK) {
            // ACK
            this.createAckMessage(src, mvInfo)
        } else if (dstId == -1) {
            // broadcast
            this.createBroadcastMessage(src, mvInfo)
        } else {
            let dst = this.nodes[dstId];
            this.createUnicastMessage(src, dst, mvInfo)
        }
    }

    visSetNodePartitionId(nodeId, partitionId) {
        this.nodes[nodeId].partition = partitionId
    }

    visShowDemoLegend(x, y, title) {
        console.log("ShowDemoLegend not implemented")
    }

    visCountDown(durationMs, title) {
        console.log("CountDown not implemented")
    }

    ctrlAddNode(x, y, isRouter, mode, cb) {
        let req = new AddNodeRequest();
        req.setX(x);
        req.setY(y);
        req.setIsRouter(isRouter);
        req.setMode(mode);
        this.grpcServiceClient.ctrlAddNode(req, {}, cb)
    }

    ctrlMoveNodeTo(nodeId, x, y, cb) {
        let req = new MoveNodeToRequest();
        req.setNodeId(nodeId);
        req.setX(x);
        req.setY(y);
        this.grpcServiceClient.ctrlMoveNodeTo(req, {}, cb)
    }

    ctrlDeleteNode(nodeId, cb) {
        let req = new DeleteNodeRequest();
        req.setNodeId(nodeId);
        this.grpcServiceClient.ctrlDeleteNode(req, {}, cb)
    }

    ctrlSetNodeFailed(nodeId, failed, cb) {
        let req = new SetNodeFailedRequest();
        req.setNodeId(nodeId);
        req.setFailed(failed);
        this.grpcServiceClient.ctrlSetNodeFailed(req, {}, cb)
    }

    ctrlSetSpeed(speed, cb) {
        let req = new SetSpeedRequest();
        req.setSpeed(speed);
        this.grpcServiceClient.ctrlSetSpeed(req, {}, cb)
    }

    getPartitionColor(parid) {
        if (parid === 0) {
            return 0x000000
        }

        return parid
    }


    setSelectedNode(id) {
        if (id === this._selectedNodeId) {
            return
        }

        let old_sel = this.nodes[this._selectedNodeId];
        if (old_sel) {
            old_sel.onUnselected()
        }
        delete this._selectedNodeId;

        let new_sel = this.nodes[id];
        if (new_sel) {
            this._selectedNodeId = id;
            new_sel.onSelected()
        }

        this.actionBar.setContext(new_sel || "any")
    }

    newNode(x, y, isRouter, mode) {
        this.ctrlAddNode(x, y, isRouter, mode, (err, resp) => {
        })
    }

    setSpeed(speed) {
        this.ctrlSetSpeed(speed, (err, resp) => {
        })
    }

    deleteSelectedNode() {
        let sel = this.nodes[this._selectedNodeId];
        if (sel) {
            this.ctrlDeleteNode(sel.id, (err, resp) => {
            })
        }
    }

    setSelectedNodeFailed(failed) {
        let sel = this.nodes[this._selectedNodeId];
        if (!sel) {
            return
        }

        this.ctrlSetNodeFailed(sel.id, failed, (err, resp) => {
        })
    }

    clearAllNodes() {
        for (let id in this.nodes) {
            this.ctrlDeleteNode(id, (err, resp) => {
            })
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
        graphics.lineStyle(linkLineWidth, 0x8bc34a, 1);

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
                    graphics.moveTo(node.position.x, node.position.y);
                    graphics.lineTo(child.position.x, child.position.y)
                }
            }
        }
        graphics.endFill();

        graphics.beginFill(0x1976d2);
        graphics.lineStyle(linkLineWidth, 0x1976d2, 1);

        for (let nodeid in this.nodes) {
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
        this.nodes[nodeId].addRouterTable(extaddr)
    }

    visRemoveRouterTable(nodeId, extaddr) {
        this.nodes[nodeId].removeRouterTable(extaddr)
    }

    visAddChildTable(nodeId, extaddr) {
        this.nodes[nodeId].addChildTable(extaddr)
    }

    visRemoveChildTable(nodeId, extaddr) {
        this.nodes[nodeId].removeChildTable(extaddr)
    }

    formatTime() {
        let secs = Math.floor(this.curTime / 1000000);
        let d = Math.floor(secs / 86400);
        secs = secs % 86400;
        let h = Math.floor(secs / 3600);
        secs = secs % 3600;
        let m = Math.floor(secs / 60);
        secs = secs % 60;
        return d + "d" + h + "h" + m + "m" + secs + "s"
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
        this.actionBar.position.set(10, height - this.actionBar.height - 20 - 10)
    }

}
