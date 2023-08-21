// Copyright (c) 2022-2023, The OTNS Authors.
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
import Button from "./Button";
import {
    MAX_SPEED,
    PAUSE_SPEED,
    TUNE_SPEED_SETTINGS,
    NODE_SPACING_ABOVE_ACTIONBAR_PX
} from "./consts";
import {Resources} from "./resources";

const {
    OtDeviceRole, NodeMode,
} = require('../proto/visualize_grpc_pb.js');

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
        this._tuneSpeedIndex = TUNE_SPEED_SETTINGS.indexOf(1);

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
        this._speedDisplayBtn.minWidth = 88;
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
        // add node context buttons
        this.addButton("Delete", "node", "del", (e) => {
            this.actionDelete(e)
        });
        this.addButton("Radio Off", "node", "radio", (e) => {
            this.actionRadioOff(e)
        });
        this.addButton("Radio On", "node", "radio", (e) => {
            this.actionRadioOn(e)
        });
        this._logClearButton = this.addButton("Clear Log", "any", "", (e) => {
            this.actionClearLog()
        });
        this._logOnOffButton = this.addButton("Show Log", "any", "", (e) => {
            this.actionToggleLogWindow()
        });
        this._energyChartOnOffButton = this.addButton("Energy-stats", "any", "", (e) => {
            this.actionOpenEnergyWindow()
        });
        this.addButton("Delete All", "any", "del", (e) => {
            this.actionClear(e)
        });
    }

    setAbilities(abilities) {
        this._abilities = abilities;
        this._resetButtons();
    }

    setSpeed(speed) {
        let oldspeed = this.speed;
        this.speed = speed;
        if (Math.round(this.speed / oldspeed * 100) != 1.0) {
            this.refresh()
        }
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
        if (this.vis.isPaused())
            return;
        this._tuneSpeedIndex--;
        if (this._tuneSpeedIndex < 0){
            this._tuneSpeedIndex = 0;
        }else{
            this.setNewSpeed(TUNE_SPEED_SETTINGS[this._tuneSpeedIndex]);
        }
    }

    actionSpeedUp() {
        if (this.vis.isPaused())
            return;
        this._tuneSpeedIndex++;
        if (this._tuneSpeedIndex >= TUNE_SPEED_SETTINGS.length){
            this._tuneSpeedIndex = TUNE_SPEED_SETTINGS.length-1;
        }else{
            this.setNewSpeed(TUNE_SPEED_SETTINGS[this._tuneSpeedIndex]);
        }
    }

    setNewSpeed(speed) {
        if (speed >= MAX_SPEED) {
            speed = MAX_SPEED
        } else if (speed <= PAUSE_SPEED) {
            speed = PAUSE_SPEED;
        }
        this.vis.setSpeed(speed);
    }

    actionNewRouter(e) {
        this.vis.ctrlAddNode("router")
    }

    actionNewFED(e) {
        this.vis.ctrlAddNode("fed")
    }

    actionNewMED(e) {
        this.vis.ctrlAddNode("med")
    }

    actionNewSED(e) {
        this.vis.ctrlAddNode("sed")
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

    actionClearLog() {
        this.vis.clearLogWindow()
    }

    actionToggleLogWindow() {
        let vis = this.vis;
        if (!vis.logWindow) {
            this.vis.showLogWindow();
            this._logOnOffButton.text = "Hide Log"
        } else {
            this.vis.hideLogWindow();
            this._logOnOffButton.text = "Show Log"
        }
        this.vis.actionBar.refresh()
    }
    
    actionOpenEnergyWindow() {
        window.open(document.location.href.replace("/visualize","/energyViewer"), '_blank');
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
        //
    }
}
