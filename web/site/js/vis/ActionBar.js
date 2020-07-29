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
import Button from "./Button";
import {MAX_SPEED, PAUSE_SPEED} from "./consts";
import {Resources} from "./resources";

const {
    OtDeviceRole, NodeMode,
} = require('../proto/visualize_grpc_pb.js');

const TUNE_SPEED_MIN = 0.25;
const TUNE_SPEED_MAX = 1024;

export default class ActionBar extends VObject {
    constructor() {
        super();

        this.root = new PIXI.Container();
        this.root.interactive = true;
        this.root.hitArea = new PIXI.Rectangle(0, 0, 0, 0);

        this._buttons = [];
        this._buttonContext = {};
        this._buttonRequiredAbility = {};
        this._abilities = {};

        this.addButton("<<", "any", "speed", (e) => {
            this.actionSpeedDown()
        });
        this._speedDisplayBtn = this.addButton("1X", "any", "", (e) => {
            this.actionTogglePauseResume()
        }, (button) => {
            let sprite;
            if (this.vis.isPaused()) {
                button.text = "PAUSED";
                sprite = new PIXI.Sprite(Resources().Play32.texture)
            } else if (this.vis.isMaxSpeed()) {
                button.text = "MAX";
                sprite = new PIXI.Sprite(Resources().Pause32.texture)
            } else {
                button.text = this.vis.speed + "X";
                sprite = new PIXI.Sprite(Resources().Pause32.texture)
            }

            sprite.tint = 0x4193F5;
            button.sprite = sprite
        });
        this.addButton(">>", "any", "speed", (e) => {
            this.actionSpeedUp()
        });
        this.addButton("New Router", "any", "add", (e) => {
            this.actionNewRouter(e)
        });
        this.addButton("FED", "any", "add", (e) => {
            this.actionNewFED(e)
        });
        this.addButton("MED", "any", "add", (e) => {
            this.actionNewMED(e)
        });
        this.addButton("SED", "any", "add", (e) => {
            this.actionNewSED(e)
        });
        this.addButton("Clear", "any", "del", (e) => {
            this.actionClear(e)
        });
        // add node context buttons
        this.addButton("Delete", "node", "del", (e) => {
            this.actionDelete(e)
        });
        this.addButton("Radio Off", "node", "radio", (e) => {
            this.actionRadioOff(e)
        });
        this.addButton("Radio On", "node", "radio", (e) => {
            this.actionRadioOn(e)
        })
    }

    setAbilities(abilities) {
        this._abilities = abilities;
        this._resetButtons();
    }

    setSpeed(speed) {
        this.speed = speed;
        this.refresh()
    }

    actionTogglePauseResume() {
        if (this.vis.isPaused()) {
            this.vis.setSpeed(typeof this.vis._resumeSpeed == "undefined" ? 1 : this.vis._resumeSpeed)
        } else {
            this.vis._resumeSpeed = this.vis.speed;
            this.vis.setSpeed(PAUSE_SPEED)
        }
    }

    actionSpeedDown() {
        if (!this.vis.isPaused()) {
            if (this.vis.isMaxSpeed()) {
                this.setNewSpeed(TUNE_SPEED_MAX)
            } else {
                this.setNewSpeed(this.vis.speed / 2)
            }
        }
    }

    actionSpeedUp() {
        if (!this.vis.isPaused() && !this.vis.isMaxSpeed()) {
            if (this.vis.speed >= TUNE_SPEED_MAX) {
                this.setNewSpeed(MAX_SPEED)
            } else {
                this.setNewSpeed(this.vis.speed * 2)
            }
        }
    }

    setNewSpeed(speed) {
        if (speed >= MAX_SPEED) {
            speed = MAX_SPEED
        } else if (speed <= PAUSE_SPEED) {
            speed = PAUSE_SPEED;
        } else {
            if (speed >= 1) {
                speed = Math.round(speed)
            } else {
                speed = Math.fround(speed)
            }
            speed = Math.max(TUNE_SPEED_MIN, speed);
            speed = Math.min(TUNE_SPEED_MAX, speed);
        }
        this.vis.setSpeed(speed)
    }

    actionNewRouter(e) {
        let pos = e.data.getLocalPosition(this.vis._root);
        let mode = new NodeMode();
        mode.setRxOnWhenIdle(true);
        mode.setSecureDataRequests(true);
        mode.setFullThreadDevice(true);
        mode.setFullNetworkData(true);
        this.vis.ctrlAddNode(Math.round(pos.x), Math.round(pos.y - 100), "router")
    }

    actionNewFED(e) {
        let pos = e.data.getLocalPosition(this.vis._root);
        let mode = new NodeMode();
        mode.setRxOnWhenIdle(true);
        mode.setSecureDataRequests(true);
        mode.setFullThreadDevice(true);
        mode.setFullNetworkData(true);
        this.vis.ctrlAddNode(Math.round(pos.x), Math.round(pos.y - 100), "fed")
    }

    actionNewMED(e) {
        let pos = e.data.getLocalPosition(this.vis._root);
        let mode = new NodeMode();
        mode.setRxOnWhenIdle(true);
        mode.setSecureDataRequests(true);
        mode.setFullThreadDevice(false);
        mode.setFullNetworkData(true);
        this.vis.ctrlAddNode(Math.round(pos.x), Math.round(pos.y - 100), "med")
    }

    actionNewSED(e) {
        let pos = e.data.getLocalPosition(this.vis._root);
        let mode = new NodeMode();
        mode.setRxOnWhenIdle(false);
        mode.setSecureDataRequests(true);
        mode.setFullThreadDevice(false);
        mode.setFullNetworkData(true);
        this.vis.ctrlAddNode(Math.round(pos.x), Math.round(pos.y - 100), "sed")
    }

    actionDelete(e) {
        this.vis.deleteSelectedNode()
    }

    actionRadioOff(e) {
        this.vis.setSelectedNodeFailed(true)
    }

    actionRadioOn(e) {
        this.vis.setSelectedNodeFailed(false)
    }

    actionClear() {
        this.vis.clearAllNodes()
    }

    addButton(label, context, requiredAbility, callback, onRefresh) {
        let btn = new Button(this, label, callback, onRefresh);
        this.addChild(btn);
        let btnIndex = this._buttons.length;
        this._buttons.push(btn);
        this._buttonContext[btnIndex] = context;
        this._buttonRequiredAbility[btnIndex] = requiredAbility;
        this._resetButtons();
        return btn
    }

    setContext(obj) {
        let context;
        if (typeof obj == "string") {
            context = obj
        } else {
            context = obj.getActionContext()
        }

        this._currentContext = context;
        this._resetButtons()
    }

    refresh() {
        for (let i in this._buttons) {
            let btn = this._buttons[i];
            btn.refresh()
        }
        this._resetButtons()
    }

    _resetButtons() {
        let x = 0;
        let maxHeight = 0;
        for (let i in this._buttons) {
            let btn = this._buttons[i];
            let btnContext = this._buttonContext[i];
            let btnRequiredAbility = this._buttonRequiredAbility[i];

            let contextMatch = btnContext === "any" || btnContext === this._currentContext;
            let abilityMatch = btnRequiredAbility === "" || this._abilities[btnRequiredAbility];
            if (contextMatch && abilityMatch) {
                btn.visible = true;
                btn.position.set(x + btn.width / 2, btn.height / 2);
                x += btn.width;
                maxHeight = Math.max(maxHeight, btn.height)
            } else {
                btn.visible = false
            }
        }
        this.root.hitArea = new PIXI.Rectangle(0, 0, x, maxHeight)
    }

    onDraggingDone() {

    }
}
