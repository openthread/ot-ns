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
import {BUTTON_LABEL_FONT_FAMILY} from "./consts";

export default class Button extends VObject {
    constructor(owner, text, callback, onRefresh) {
        super();
        this.owner = owner;
        this._callback = callback;
        this._onRefresh = onRefresh;
        this._minWidth = 16;

        let label = new PIXI.Text(text, {fontFamily: BUTTON_LABEL_FONT_FAMILY, fontSize: 16});
        this._label = label;
        label.anchor.set(0.5, 0.5);

        let graphics = new PIXI.Graphics();
        this.root = graphics;
        this.root.addChild(label);
        this.root.interactive = true;

        let button = this;
        this.setOnTouchStart((e) => {
            this.onTouchStart()
        });
        this.setOnTouchEnd((e) => {
            this.onTouchEnd()
        });
        this.setOnTap((e) => {
            if (button.owner.isDragging()) {
                return
            }

            button.owner.clearDraggingStates(e);
            e.stopPropagation();
            button.onTap(e)
        });

        this._resetlayout()
    }

    get minWidth() {
        return this._minWidth
    }

    set minWidth(w) {
        this._minWidth = w;
        this._resetlayout();
    }

    get sprite() {
        return this._sprite
    }

    set sprite(sp) {
        if (this._sprite) {
            this._sprite.destroy()
        }
        this._sprite = sp;
        if (this._sprite) {
            this.addChild(this._sprite)
        }
        let targetWidth = 16;
        let targetHeight = 16;
        sp.scale.set(targetWidth / sp.width, targetHeight / sp.height);
        sp.anchor.set(0.5, 0.5);
        this._resetlayout()
    }

    get text() {
        return this._label.text
    }

    set text(v) {
        this._label.text = v;
        this._resetlayout()
    }

    refresh() {
        if (this._onRefresh) {
            this._onRefresh(this)
        }
    }

    _resetlayout() {
        let graphics = this.root;
        let label = this._label;
        let sprite = this._sprite;

        let labelWidth = label.width;
        let labelHeight = label.height;
        let spriteWidth = sprite ? sprite.width : 0;
        let spriteHeight = sprite ? sprite.height : 0;
        let spriteLabelGap = sprite ? 3 : 0;

        let width = spriteWidth + labelWidth + spriteLabelGap + 16;
        let height = Math.max(labelHeight, spriteHeight) + 16;

        if (width < this.minWidth)
            width = this.minWidth

        graphics.clear();
        graphics.beginFill(0xeeeeee);
        graphics.lineStyle(2, 0x424242);
        graphics.drawRoundedRect(-width / 2, -height / 2, width, height, 7);
        graphics.endFill();

        if (sprite) {
            sprite.position.set(-width / 2 + spriteWidth / 2 + 8, 0)
        }

        label.position.set(-width / 2 + 8 + spriteWidth + spriteLabelGap + labelWidth / 2, 0)
    }

    onTouchStart() {
        this.root.tint = 0x484848
    }

    onTouchEnd() {
        this.root.tint = 0xffffff
    }

    onTap(e) {
        this._callback(e)
    }
}
