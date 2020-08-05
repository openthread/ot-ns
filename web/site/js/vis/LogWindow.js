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

import * as PIXI from "pixi.js-legacy";
import VObject from "./VObject";
import {Scrollbox} from 'pixi-scrollbox'


const LOG_TEXT_STYLE = {
    fill: "#0052ff",
    fontFamily: "Verdana",
    fontSize: 11,
    fontStyle: "italic",
    fontWeight: "normal"
};

const LOG_TEXT_LINE_HEIGHT = 15;
export const LOG_WINDOW_WIDTH = 400;

export default class LogWindow extends VObject {
    constructor() {
        super();

        let height = window.innerHeight - 100;

        this._root = new Scrollbox({
            boxWidth: LOG_WINDOW_WIDTH,
            boxHeight: height,
            fade: true,
            overflowX: "none",
            overflowY: "scroll"
        });
        this._root.update();
        this.loglist = [];
    }

    addLog(text) {
        let textCtrl = new PIXI.Text(text, LOG_TEXT_STYLE);
        textCtrl.position.set(3, 3 + this.loglist.length * LOG_TEXT_LINE_HEIGHT);
        this._root.content.addChild(textCtrl);
        this._root.resize({boxWidth: LOG_WINDOW_WIDTH, boxHeight: this._root.boxHeight});
        this._root.ensureVisible(textCtrl.x, textCtrl.y, textCtrl.width, textCtrl.height);
        this.loglist.push(textCtrl);
    }

    resetLayout(width, height) {
        this._root.resize({boxWidth: LOG_WINDOW_WIDTH, boxHeight: height - 100})
    }
}
