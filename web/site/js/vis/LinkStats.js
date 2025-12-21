// Copyright (c) 2025, The OTNS Authors.
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

import VObject from "./VObject";
import * as PIXI from "pixi.js-legacy";
import {Resources} from "./resources";

export default class LinkStats extends VObject {
    constructor(node, peer) {
        super();
        this._root = new PIXI.Container();
        this._node = node;
        this._peer = peer;

        // compute point
        if (peer) {
            let s = new PIXI.Sprite(Resources().FailedNodeMark.texture);
            s.anchor.set(0.5, 0.5);
            s.scale.set(0.1, 0.1);
            s.visible = true;
            this.addChild(s);
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, 32));
        }
    }

    onPositionChange() {
        console.log("onPositionChange() for LinkStats node=" + this._node.id + " peer="+ this._peer.id);
        if (this._node && this._peer && !this._peer.destroyed) {
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, 32));
            console.log(" - LinkStats position set to: " + this.position.x + "," + this.position.y);
            this._peer.onPeerPositionChange(this._node.extAddr);
        }
    }

    onPeerPositionChange() {
        console.log("onPeerPositionChange() for LinkStats node=" + this._node.id + " peer="+ this._peer.id);
        if (this._node && this._peer && !this._peer.destroyed) {
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, 32));
            console.log(" - LinkStats position set to: " + this.position.x + "," + this.position.y);
        }
    }

    update(dt) {
        super.update(dt);
        if (!this._peer || !this._node || this._peer.destroyed || this._node.destroyed) {
            this.destroy();
        }
    }

}

function calcVector(srcPos, destPos, dist) {
    const dx = destPos.x - srcPos.x;
    const dy = destPos.y - srcPos.y;
    const totalDist = Math.sqrt(dx * dx + dy * dy);

    if (totalDist === 0 || dist >= totalDist) {
        return new PIXI.Point(dx, dy);
    }

    const ratio = dist / totalDist;
    return new PIXI.Point(dx * ratio, dy * ratio);
}
