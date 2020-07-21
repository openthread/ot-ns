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
import {Visualizer} from "./PixiVisualizer";
import {Resources} from "./resources";
import {COLOR_ACK_MESSAGE} from "./consts";

let nextMessageId = 1;


export class BroadcastMessage extends VObject {
    constructor(src, mvInfo) {
        super();
        this.id = nextMessageId;
        nextMessageId += 1;
        this.mvInfo = mvInfo;

        let beginRadius = 32;
        let sprite = new PIXI.Sprite(Resources().WhiteDashed8Circle128.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(beginRadius * 2 / 128, beginRadius * 2 / 128);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = src.position;
        this._targetRadius = src.radioRange;
        sprite.alpha = 0.1;
        this._root = this.sprite = sprite;
        this._leftFlyTime = 0.7
    }

    getColor() {
        return 0x1565c0
    }

    isBroadcast() {
        return true;
    }

    update(dt) {
        if (this._leftFlyTime <= 0) {
            Visualizer().deleteMessage(this);
            return
        }

        dt = Math.min(dt, this._leftFlyTime);
        this._leftFlyTime -= dt;
        let beginRadius = 32;
        let playRatio = 1.0 - this._leftFlyTime / 0.7;
        let radius = beginRadius + (this._targetRadius - beginRadius) * Math.pow(playRatio, 0.5);
        this.sprite.scale.set(radius * 2 / 128, radius * 2 / 128)
    }
}

export class UnicastMessage extends VObject {

    constructor(src, dst, mvInfo) {
        super();

        this.id = nextMessageId;
        nextMessageId += 1;
        if (dst) {
            this.dstPos = dst.position
        } else {
            this.dstPos = new PIXI.Point(src.position.x, src.position.y + 200)
        }
        this.mvInfo = mvInfo;

        let size = 10;
        let sprite = new PIXI.Sprite(Resources().WhiteSolidHexagon64.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(size / 64, size / 64);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = src.position;
        this._root = this.sprite = sprite;
        this._leftFlyTime = 0.7
    }

    isBroadcast() {
        return false;
    }

    getColor() {
        return 0xff8f00
    }

    update(dt) {
        super.update(dt);

        if (this._leftFlyTime <= 0) {
            Visualizer().deleteMessage(this);
            return
        }

        let leftTime = this._leftFlyTime;
        dt = Math.min(dt, leftTime);
        let mx, my;
        let dx = this.dstPos.x - this.position.x;
        let dy = this.dstPos.y - this.position.y;
        let r = dt / leftTime;
        mx = dx * r;
        my = dy * r;
        this.position.set(this.position.x + mx, this.position.y + my);
        this._leftFlyTime -= dt
    }
}

export class AckMessage extends VObject {

    constructor(src, mvInfo) {
        super();

        this.id = nextMessageId;
        nextMessageId += 1;
        this.dstPos = new PIXI.Point(src.position.x, src.position.y + 50)
        this.mvInfo = mvInfo;

        let size = 10;
        let sprite = new PIXI.Sprite(Resources().WhiteSolidTriangle64.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(size / 64, size / 64);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = src.position;
        this._root = this.sprite = sprite;
        this._leftFlyTime = 0.7
    }

    isBroadcast() {
        return false;
    }

    getColor() {
        return COLOR_ACK_MESSAGE;
    }

    update(dt) {
        super.update(dt);

        if (this._leftFlyTime <= 0) {
            Visualizer().deleteMessage(this);
            return
        }

        let leftTime = this._leftFlyTime;
        dt = Math.min(dt, leftTime);
        let mx, my;
        let dx = this.dstPos.x - this.position.x;
        let dy = this.dstPos.y - this.position.y;
        let r = dt / leftTime;
        mx = dx * r;
        my = dy * r;
        this.position.set(this.position.x + mx, this.position.y + my);
        this._leftFlyTime -= dt
    }
}
