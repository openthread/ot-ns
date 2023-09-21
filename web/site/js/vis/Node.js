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
import {NodeMode, OtDeviceRole} from '../proto/visualize_grpc_pb'
import {Visualizer} from "./PixiVisualizer";
import {Resources} from "./resources";
import {NODE_ID_INVALID, NODE_LABEL_FONT_FAMILY, NODE_LABEL_FONT_SIZE, POWER_DBM_INVALID,
        EXT_ADDR_INVALID} from "./consts";

const NODE_SHAPE_SCALE = 64;
const NODE_SELECTION_SCALE = 128;
const CIRCULAR_SHAPE_RADIUS = 20;
const HEXAGONAL_SHAPE_RADIUS = 22;

let vis = Visualizer();

export default class Node extends VObject {
    constructor(nodeId, x, y, radioRange) {
        super();

        this.id = nodeId;
        this.extAddr = EXT_ADDR_INVALID;
        this.radioRange = radioRange;
        this.nodeMode = new NodeMode([true, true, true, true]);
        this.rloc16 = 0xfffe;
        this.routerId = NODE_ID_INVALID;
        this.childId = NODE_ID_INVALID;
        this.parentId = NODE_ID_INVALID;
        this.role = OtDeviceRole.OT_DEVICE_ROLE_DISABLED;
        this.txPowerLast = POWER_DBM_INVALID;
        this.channelLast = -1;
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

        let radius = CIRCULAR_SHAPE_RADIUS;
        let statusSprite = this._createStatusSprite();
        statusSprite.tint = this.getRoleColor();
        statusSprite.anchor.x = 0.5;
        statusSprite.anchor.y = 0.5;
        statusSprite.scale.x = statusSprite.scale.y = radius * 2 / NODE_SHAPE_SCALE;
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

        let partitionSprite = this._createPartitionSprite();
        partitionSprite.anchor.x = 0.5;
        partitionSprite.anchor.y = 0.5;
        partitionSprite.scale.x = partitionSprite.scale.y = radius * 2 / NODE_SHAPE_SCALE / 1.5;
        partitionSprite.tint = this.vis.getPartitionColor(this._partition);
        this._root.addChild(partitionSprite);
        this._partitionSprite = partitionSprite;

        this._updateSize();

        let label = new PIXI.Text("", {fontFamily: NODE_LABEL_FONT_FAMILY, fontSize: NODE_LABEL_FONT_SIZE, align: 'left'});
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
        return new PIXI.Sprite(this._getStatusSpriteTexture());
    }

    _createPartitionSprite() {
        return new PIXI.Sprite(this._getPartitionSpriteTexture());
    }

    _getStatusSpriteTexture() {
        switch (this.role) {
            case OtDeviceRole.OT_DEVICE_ROLE_LEADER:
            case OtDeviceRole.OT_DEVICE_ROLE_ROUTER:
                return Resources().WhiteSolidHexagon64.texture;
        }
        if (this.nodeMode.getFullThreadDevice()) {
            return Resources().WhiteSolidCircle64.texture;
        } else if (this.nodeMode.getRxOnWhenIdle()) {
            return Resources().WhiteDashed4Circle64.texture;
        } else {
            return Resources().WhiteDashed8Circle64.texture;
        }
    }

    _getPartitionSpriteTexture() {
        switch (this.role) {
            case OtDeviceRole.OT_DEVICE_ROLE_LEADER:
            case OtDeviceRole.OT_DEVICE_ROLE_ROUTER:
                return Resources().WhiteSolidHexagon64.texture;
            default:
                return Resources().WhiteSolidCircle64.texture;
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

    _updateSize() {
        let radius = CIRCULAR_SHAPE_RADIUS;
        switch (this.role) {
            case OtDeviceRole.OT_DEVICE_ROLE_LEADER:
            case OtDeviceRole.OT_DEVICE_ROLE_ROUTER:
                radius = HEXAGONAL_SHAPE_RADIUS
        }
        this._statusSprite.scale.x = this._statusSprite.scale.y = radius * 2 / NODE_SHAPE_SCALE;
        this._partitionSprite.scale.x = this._partitionSprite.scale.y = radius * 2 / NODE_SHAPE_SCALE / 1.5;
    }

    setRloc16(rloc16) {
        this.rloc16 = rloc16;
        this.routerId = rloc16 >> 10;
        this.childId = rloc16 & 0x01ff;
        this._updateLabel()
    }

    setRole(role) {
        if (role != this.role) {
            this.role = role;
            this._statusSprite.tint = this.getRoleColor();
            this._statusSprite.texture = this._getStatusSpriteTexture();
            this._partitionSprite.texture = this._getPartitionSpriteTexture();
            this._updateSize()
        }
        if (role == OtDeviceRole.OT_DEVICE_ROLE_DISABLED || role == OtDeviceRole.OT_DEVICE_ROLE_DETACHED) {
            this._parent = NODE_ID_INVALID;
            this.parentId = NODE_ID_INVALID;
            this.childId = NODE_ID_INVALID;
            this.routerId = NODE_ID_INVALID;
        }
    }

    setMode(mode) {
        if (mode != this.nodeMode) {
            this.nodeMode = mode;
            this._statusSprite.texture = this._getStatusSpriteTexture();
        }
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
            let selbox = new PIXI.Sprite(Resources().WhiteRoundedDashedSquare128.texture);
            selbox.tint = 0x2e7d32;
            selbox.alpha = 0.7;
            selbox.scale.set(selboxsize / NODE_SELECTION_SCALE, selboxsize / NODE_SELECTION_SCALE);
            selbox.anchor.set(0.5, 0.5);
            this.root.addChildAt(selbox, 0);
            this._selbox = selbox;

            const rangeCircleSize = this.radioRange;
            let rangeCircle = new PIXI.Graphics();
            rangeCircle.beginFill(0x98ee99, 0.2);
            rangeCircle.lineStyle({width: 1, color: 0x338a3e, alpha: 0.7});
            rangeCircle.drawCircle(0, 0, rangeCircleSize);
            rangeCircle.endFill();
            this.root.addChildAt(rangeCircle, 0);
            this._rangeCircle = rangeCircle;
        }
    }

    onUnselected() {
        this._selected = false;
        if (this._selbox) {
            this._selbox.destroy();
            delete this._selbox;

            this._rangeCircle.destroy();
            delete this._rangeCircle;
        }
    }
}