// Copyright (c) 2020-2022, The OTNS Authors.
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
import LVObject from "./LVObject";
import {Resources} from "./resources";
import {COLOR_ACK_MESSAGE} from "./consts";

const BROADCAST_MESSAGE_SCALE = 128;
const UNICAST_MESSAGE_SCALE = 64;

let nextMessageId = 1;

export class BroadcastMessage extends LVObject {
    constructor(src, mvInfo) {
        super();
        this.id = nextMessageId;
        nextMessageId += 1;
        this.mvInfo = mvInfo;
        this.src = src;

        let beginRadius = 32;
        let sprite = new PIXI.Sprite(Resources().WhiteDashed8Circle128.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(beginRadius * 2 / BROADCAST_MESSAGE_SCALE, beginRadius * 2 / BROADCAST_MESSAGE_SCALE);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = src.position;
        this._targetRadius = src.radioRange;
        sprite.alpha = 0.1;
        this._root = this.sprite = sprite;
        this.configureLifetime(mvInfo);
    }

    getColor() {
        return 0x1565c0
    }

    isBroadcast() {
        return true;
    }

    update(dt) {
        super.update(dt);
        let lft = this.getRealLifetimeRemaining();
        if (lft <= 0) {
            this.vis.deleteMessage(this);
            return
        }

        let beginRadius = 32;
        let playRatio = this.getLifetimeProgress();
        let radius = beginRadius + (this._targetRadius - beginRadius) * Math.pow(playRatio, 0.5);
        this.sprite.scale.set(radius * 2 / BROADCAST_MESSAGE_SCALE, radius * 2 / BROADCAST_MESSAGE_SCALE);
        this.position = this.src.position;
    }
}

export class UnicastMessage extends LVObject {

    constructor(src, dst, mvInfo) {
        super();

        this.id = nextMessageId;
        nextMessageId += 1;
        if (dst) {
            this.dst = dst;
        } else {
            this.dstPos = new PIXI.Point(0, 200);
            this.dst = null;
        }
        this.src = src;
        this.mvInfo = mvInfo;

        let size = 10;
        let sprite = new PIXI.Sprite(Resources().WhiteSolidHexagon64.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(size / UNICAST_MESSAGE_SCALE, size / UNICAST_MESSAGE_SCALE);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = this.src.position;
        this._root = this.sprite = sprite;
        this.configureLifetime(mvInfo);
    }

    isBroadcast() {
        return false;
    }

    getColor() {
        return 0xff8f00
    }

    update(dt) {
        super.update(dt);
        let lft = this.getRealLifetimeRemaining();
        if (lft <= 0) {
            this.vis.deleteMessage(this);
            return
        }

        let dstx, dsty;
        if (this.dst != null){  // track the (possibly moving) destination
            dstx = this.dst.x;
            dsty = this.dst.y;
        }else{
            dstx = this.src.x + this.dstPos.x;  // or use a relative dstpos.
            dsty = this.src.y + this.dstPos.y;
        }
        let dx = dstx - this.src.x;
        let dy = dsty - this.src.y;
        let r = this.getLifetimeProgress();
        this.position.set(this.src.x + r * dx, this.src.y + r * dy);
    }
}

export class AckMessage extends LVObject {

    constructor(src, mvInfo) {
        super();

        this.id = nextMessageId;
        nextMessageId += 1;
        this.dstPos = new PIXI.Point(0, 50);
        this.src = src;
        this.mvInfo = mvInfo;

        let size = 10;
        let sprite = new PIXI.Sprite(Resources().WhiteSolidTriangle64.texture);
        sprite.tint = this.getColor();
        sprite.scale.set(size / UNICAST_MESSAGE_SCALE, size / UNICAST_MESSAGE_SCALE);
        sprite.anchor.set(0.5, 0.5);
        sprite.position = src.position;
        this._root = this.sprite = sprite;
        this.configureLifetime(mvInfo);
    }

    isBroadcast() {
        return false;
    }

    getColor() {
        return COLOR_ACK_MESSAGE;
    }

    update(dt) {
        super.update(dt);
        let lft = this.getRealLifetimeRemaining();
        if (lft <= 0) {
            this.vis.deleteMessage(this);
            return
        }

        let r = this.getLifetimeProgress();
        this.position.set(this.src.x + r * this.dstPos.x, this.src.y + r * this.dstPos.y);

    }
}