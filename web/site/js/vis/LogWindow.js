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

import * as PIXI from "pixi.js";
import VObject from "./VObject";
import {
    LOG_WINDOW_FONT_FAMILY, LOG_WINDOW_FONT_SIZE, LOG_WINDOW_FONT_COLOR
} from "./consts";


const LOG_TEXT_STYLE = {
    fill: LOG_WINDOW_FONT_COLOR,
    fontFamily: LOG_WINDOW_FONT_FAMILY,
    fontSize: LOG_WINDOW_FONT_SIZE,
    fontStyle: "normal",
    fontWeight: "normal"
};

const LOG_TEXT_LINE_HEIGHT = 15;
const LOG_WINDOW_MAX_SIZE = 100;
export const LOG_WINDOW_WIDTH = 400;
const LOG_WINDOW_BOTTOM_PADDING = 100;
const WHEEL_LINES_PER_NOTCH = 3;

// A scrollable, bottom-tailing log view. Replaces the former pixi-scrollbox
// dependency with a masked container scrolled via the mouse wheel.
export default class LogWindow extends VObject {
    constructor() {
        super();

        this.boxHeight = window.innerHeight - LOG_WINDOW_BOTTOM_PADDING;
        this.loglist = [];
        // While true, new lines keep the view pinned to the bottom (tail mode).
        // The user scrolling up with the wheel temporarily disables it.
        this._stickToBottom = true;

        this._root = new PIXI.Container();
        this._root.eventMode = 'static';

        // White highlight bar drawn just below the most recent line.
        this.lastline = new PIXI.Graphics();
        this.lastline.beginFill(0xFFFFFF);
        this.lastline.drawRect(0, 0, LOG_WINDOW_WIDTH, LOG_TEXT_LINE_HEIGHT);
        this.lastline.endFill();

        this.logContainer = new PIXI.Container();
        this.logContainer.addChild(this.lastline);
        this._root.addChild(this.logContainer);

        // Clip the log lines to the box bounds.
        this._mask = new PIXI.Graphics();
        this._root.addChild(this._mask);
        this.logContainer.mask = this._mask;

        this._root.on('wheel', (e) => this._onWheel(e));

        this._applyBoxSize();
    }

    addLog(text, color = LOG_WINDOW_FONT_COLOR) {
        if (this.loglist.length === LOG_WINDOW_MAX_SIZE) {
            let rm = this.loglist.shift();
            this.logContainer.removeChild(rm)
        }

        let log = new PIXI.Text(text, Object.assign({}, LOG_TEXT_STYLE, {fill: color}));
        this.logContainer.addChild(log);
        this.loglist.push(log);
        this._relayout();
    }

    clear() {
        for (const log of this.loglist) {
            this.logContainer.removeChild(log)
        }
        this.loglist = [];
        this._stickToBottom = true;
        this._relayout();
    }

    resetLayout(width, height) {
        this.boxHeight = height - LOG_WINDOW_BOTTOM_PADDING;
        this._applyBoxSize();
    }

    _applyBoxSize() {
        this._mask.clear();
        this._mask.beginFill(0xFFFFFF);
        this._mask.drawRect(0, 0, LOG_WINDOW_WIDTH, this.boxHeight);
        this._mask.endFill();
        this._root.hitArea = new PIXI.Rectangle(0, 0, LOG_WINDOW_WIDTH, this.boxHeight);
        this._relayout();
    }

    _relayout() {
        for (let i = 0; i < this.loglist.length; i++) {
            this.loglist[i].position.set(3, 3 + i * LOG_TEXT_LINE_HEIGHT)
        }
        this.lastline.position.set(0, 3 + this.loglist.length * LOG_TEXT_LINE_HEIGHT);

        if (this._stickToBottom) {
            this.logContainer.y = this._minScrollY();
        } else {
            this._clampScroll();
        }
    }

    _contentHeight() {
        // log lines + the trailing highlight bar + top padding
        return (this.loglist.length + 1) * LOG_TEXT_LINE_HEIGHT + 3;
    }

    // Most-negative allowed logContainer.y (i.e. scrolled fully to the bottom).
    _minScrollY() {
        return Math.min(0, this.boxHeight - this._contentHeight());
    }

    _clampScroll() {
        this.logContainer.y = Math.max(this._minScrollY(), Math.min(0, this.logContainer.y))
    }

    _onWheel(e) {
        let dir = e.deltaY > 0 ? 1 : -1;
        this.logContainer.y -= dir * WHEEL_LINES_PER_NOTCH * LOG_TEXT_LINE_HEIGHT;
        this._clampScroll();
        // Re-enable tailing once the user scrolls back to the bottom.
        this._stickToBottom = this.logContainer.y <= this._minScrollY() + 1;
    }
}
