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

import * as PIXI from "pixi.js-legacy";

import VObject from "./VObject";

export default class LinkStats extends VObject {
    constructor(node, peer, textLabel, distanceRelative, textStyle) {
        super();
        this._root = new PIXI.Container();
        this._node = node;
        this._peer = peer;
        this._text = null;
        this._distanceRel = distanceRelative;
        this._minPixelsFromNode = textStyle.distanceFromNode;

        if (node && peer) {
            this._text = new PIXI.Text(textLabel, textStyle);
            this._text.anchor.set(0.5, 0.5);
            this.visible = true;
            this._root.visible = true;
            this.addChild(this._text);
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, this._distanceRel, this._minPixelsFromNode));
        }
    }

    setTextLabel(textLabel) {
        this._text.text = textLabel;
    }

    onPositionChange() {
        if (this._node && this._peer && !this._peer.destroyed) {
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, this._distanceRel, this._minPixelsFromNode));
            this._peer.onPeerPositionChange(this._node.extAddr);
        }
    }

    onPeerPositionChange() {
        if (this._node && this._peer && !this._peer.destroyed) {
            this.position.copyFrom(calcVector(this._node.position, this._peer.position, this._distanceRel, this._minPixelsFromNode));
        }
    }

    update(dt) {
        super.update(dt);
    }

}

function calcVector(srcPos, destPos, distanceRel, minPixelsFromNode) {
    const dx = destPos.x - srcPos.x;
    const dy = destPos.y - srcPos.y;
    const totalDist = Math.sqrt(dx * dx + dy * dy);

    if (totalDist === 0 || totalDist <= 2 * minPixelsFromNode) {
        return new PIXI.Point(dx/2.0, dy/2.0); // if really tight, pick middle ground
    }

    const unitX = dx / totalDist;
    const unitY = dy / totalDist;
    const availableDist = totalDist - 2 * minPixelsFromNode;
    const travelDist = minPixelsFromNode + (availableDist * (distanceRel / 100.0));

    return new PIXI.Point(unitX * travelDist, unitY * travelDist);
}
