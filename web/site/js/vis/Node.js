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
import {OtDeviceRole} from '../proto/visualize_grpc_pb'
import {Visualizer} from "./PixiVisualizer";
import {Resources} from "./resources";

const NODE_LABEL_FONT_FAMILY = 'Comic Sans MS';

let vis = Visualizer();

export default class Node extends VObject {
    constructor(nodeId, x, y, extAddr, radioRange, nodeMode) {
        super();

        this.id = nodeId;
        this.extAddr = extAddr;
        this.radioRange = radioRange;
        this.nodeMode = nodeMode;
        this.rloc16 = 0xfffe;
        this.role = OtDeviceRole.OT_DEVICE_ROLE_DISABLED;
        this._failed = false;
        this._parent = 0;
        this._partition = 0;
        this._children = {};
        this._neighbors = {};
        this._createTime = this.vis.curTime;
        // this._draggingOffset = null
        this._selected = false;

        this._root = new PIXI.Container();
        this.x = x;
        this.y = y;
        this.position.set(x, y);

        let radius = 20;
        let statusSprite = this._createStatusSprite();
        statusSprite.tint = this.getRoleColor();
        statusSprite.anchor.x = 0.5;
        statusSprite.anchor.y = 0.5;
        statusSprite.scale.x = statusSprite.scale.y = radius * 2 / 32;
        this._root.addChild(statusSprite);
        this._statusSprite = statusSprite;

        let node = this;
        this._root.interactive = true;
        this._root.hitArea = new PIXI.Circle(0, 0, radius);
        this.setOnTouchStart((e) => {
            this.vis.setSelectedNode(node.id);
            e.stopPropagation();
        });
        this.setOnTap((e) => {
            e.stopPropagation();
        });

        this.setDraggable();

        let partitionSprite = new PIXI.Sprite(Resources().WhiteSolidCircle32.texture);
        partitionSprite.anchor.x = 0.5;
        partitionSprite.anchor.y = 0.5;
        partitionSprite.scale.x = partitionSprite.scale.y = radius * 2 / 32 / 1.5;
        partitionSprite.tint = this.vis.getPartitionColor(this._partition);
        this._root.addChild(partitionSprite);
        this._partitionSprite = partitionSprite;

        let label = new PIXI.Text("", {fontFamily: NODE_LABEL_FONT_FAMILY, fontSize: 13, align: 'left'});
        label.position.set(11, 11);
        this._root.addChild(label);
        this.label = label;
        this._updateLabel();

        let failedMask = new PIXI.Sprite(Resources().FailedNodeMark.texture);
        failedMask.anchor.set(0.5, 0.5);
        failedMask.scale.set(0.5, 0.5);
        failedMask.visible = false;
        this._root.addChild(failedMask);
        this._failedMask = failedMask
    }

    get failed() {
        return this._failed
    }

    set failed(v) {
        if (this._failed !== v) {
            this._failed = v;
            this._failedMask.visible = this._failed
        }
    }

    get parent() {
        return this._parent
    }

    set parent(v) {
        this._parent = v
    }

    get partition() {

        return this._partition
    }

    set partition(v) {
        if (v !== this._partition) {
            this._partition = v;
            this._partitionSprite.tint = this.vis.getPartitionColor(this._partition)
        }
    }

    getActionContext() {
        return "node"
    }

    _createStatusSprite() {
        if (this.nodeMode.getFullThreadDevice()) {
            return new PIXI.Sprite(Resources().WhiteSolidCircle32.texture);
        } else if (this.nodeMode.getRxOnWhenIdle()) {
            return new PIXI.Sprite(Resources().WhiteDashed4Circle32.texture);
        } else {
            return new PIXI.Sprite(Resources().WhiteDashed8Circle32.texture);
        }
    }

    setPosition(x, y) {
        this.x = x;
        this.y = y;
        if (!this.isDragging()) {
            this.position.set(x, y)
        }
    }

    _updateLabel() {
        let rloc16 = ('0000' + this.rloc16.toString(16).toUpperCase()).slice(-4);
        this.label.text = this.id.toString() + "|" + rloc16
    }

    setRloc16(rloc16) {
        this.rloc16 = rloc16;
        this._updateLabel()
    }

    setRole(role) {
        this.role = role;
        this._statusSprite.tint = this.getRoleColor()
    }

    getRoleColor() {
        if (this.failed) {
            return 0x757575
        }

        switch (this.role) {
            case OtDeviceRole.OT_DEVICE_ROLE_LEADER:
                return 0xc62828;
            case OtDeviceRole.OT_DEVICE_ROLE_ROUTER:
                return 0x1565c0;
            case OtDeviceRole.OT_DEVICE_ROLE_CHILD:
                return 0x4caf50;
            case OtDeviceRole.OT_DEVICE_ROLE_DETACHED:
                return 0x546e7a;
            case OtDeviceRole.OT_DEVICE_ROLE_DISABLED:
                return 0x757575
        }
        return 0x757575
    }

    addRouterTable(extaddr) {
        this._neighbors[extaddr] = 1
    }

    removeRouterTable(extaddr) {
        delete this._neighbors[extaddr]
    }

    addChildTable(extaddr) {
        this._children[extaddr] = 1
    }

    removeChildTable(extaddr) {
        delete this._children[extaddr]
    }

    onDraggingTimer() {
        let pos = this.position;
        this.vis.ctrlMoveNodeTo(this.id, pos.x, pos.y, (err, resp) => {
        })
    }

    onDraggingDone() {
        let pos = this.position;
        this.vis.ctrlMoveNodeTo(this.id, pos.x, pos.y, (err, resp) => {
            if (err !== null) {
                this.position.set(this.x, this.y)
            }
        })
    }

    update(dt) {
        super.update(dt);
        // this._updateDragging(dt)
    }

    onSelected() {
        this._selected = true;
        if (!this._selbox) {
            const selboxsize = 60;
            let selbox = new PIXI.Sprite(Resources().WhiteRoundedDashedSquare64.texture);
            selbox.tint = 0x2e7d32;
            selbox.alpha = 0.7;
            selbox.scale.set(selboxsize / 64, selboxsize / 64);
            selbox.anchor.set(0.5, 0.5);
            this.root.addChildAt(selbox, 0);
            this._selbox = selbox
        }
    }

    onUnselected() {
        this._selected = false;
        if (this._selbox) {
            this._selbox.destroy();
            delete this._selbox
        }
    }

}