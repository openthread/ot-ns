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
import {Scrollbox} from 'pixi-scrollbox'
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

export default class LogWindow extends VObject {
    constructor() {
        super();

        let height = window.innerHeight - LOG_WINDOW_BOTTOM_PADDING;
        this.logIndex = 0;

        this._root = new Scrollbox({
            boxWidth: LOG_WINDOW_WIDTH,
            boxHeight: height,
            fade: true,
            overflowX: "auto",
            overflowY: "auto"
        });
        this.logContainer = new PIXI.Container();
        this.lastline = new PIXI.Graphics();
        this.lastline.clear();
        this.lastline.beginFill(0xFFFFFF);
        // this.lastline.lineStyle(0);
        this.lastline.drawRect(0, 0, LOG_WINDOW_WIDTH, LOG_TEXT_LINE_HEIGHT);
        this.lastline.endFill();
        this.logContainer.addChild(this.lastline);

        this._root.content.addChild(this.logContainer);
        this._root.update();
        this.loglist = [];
    }

    addLog(text, color = LOG_WINDOW_FONT_COLOR) {
        if (this.loglist.length === LOG_WINDOW_MAX_SIZE) {
            let rm = this.loglist.shift();
            this.logContainer.removeChild(rm)
        }

        LOG_TEXT_STYLE.fill = color;
        let log = new PIXI.Text(text, LOG_TEXT_STYLE);
        log.position.set(3, 3 + this.logIndex * LOG_TEXT_LINE_HEIGHT);
        this.logIndex++;
        this.logContainer.addChild(log);
        this.loglist.push(log);

        this.logContainer.position.set(0, -this.loglist[0].position.y);
        this.lastline.position.set(0, log.y + LOG_TEXT_LINE_HEIGHT);

        this._root.resize({boxWidth: LOG_WINDOW_WIDTH, boxHeight: this._root.boxHeight});
        this._root.ensureVisible(0, log.y + this.logContainer.y, LOG_WINDOW_WIDTH, log.height);
    }

    clear() {
        this.loglist = [];
        this.logContainer.removeChildren();
        this.logContainer.position.set(0, 0);
        this.logIndex = 0;
        this._root.resize({boxWidth: LOG_WINDOW_WIDTH, boxHeight: this._root.boxHeight});
    }

    resetLayout(width, height) {
        this._root.resize({boxWidth: LOG_WINDOW_WIDTH, boxHeight: height - LOG_WINDOW_BOTTOM_PADDING})
    }
}
