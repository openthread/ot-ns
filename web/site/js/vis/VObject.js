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
import {Visualizer} from "./PixiVisualizer";


export default class VObject {

    constructor() {
        this._root = null;
        this._timerMgr = new TimerMgr(this);
        this.vis = Visualizer();
    }

    get root() {
        return this._root
    }

    set root(v) {
        this._root = v
    }

    get position() {
        return this._root.position
    }

    set position(pos) {
        this._root.position = pos
    }

    get visible() {
        return this._root.visible
    }

    set visible(v) {
        this._root.visible = v
    }

    get width() {
        return this._root.width
    }

    get height() {
        return this._root.height
    }

    get interactive() {
        return this._root.interactive
    }

    set interactive(v) {
        this._root.interactive = v
    }

    get text() {
        return this._root.text
    }

    set text(s) {
        this._root.text = s
    }

    update(dt) {
        this._timerMgr.update(dt);
    }

    addCallback(duration, callback) {
        return this._timerMgr.addCallback(duration, callback)
    }

    cancelCallback(cid) {
        return this._timerMgr.cancelCallback(cid)
    }

    addChild(child) {
        this._root.addChild(child._root || child)
    }

    addChildAt(child, index) {
        this._root.addChildAt(child._root || child, index)
    }

    removeChild(child) {
        this._root.removeChild(child._root || child)
    }

    destroy() {
        this._root.destroy()
    }

    setOnTouchStart(func) {
        this._root.on("mousedown", func);
        // this._root.on("pointerdown", func)
        this._root.on("touchstart", func)
    }

    setOnTouchEnd(func) {
        this._root.on("mouseup", func);
        this._root.on("mouseupoutside", func);
        // this._root.on("pointerup", func)
        // this._root.on("pointerupoutside", func)
        this._root.on("touchend", func);
        this._root.on("touchendoutside", func)
    }

    setOnTap(func) {
        this._root.on("click", func);
        // this._root.on("pointertap", func)
        this._root.on("tap", func)
    }

    setOnTouchMove(func) {
        this._root.on("mousemove", func);
        this._root.on("touchmove", func)
        // this._root.on("pointermove", func)
    }

    setDraggable() {
        this.setOnTouchStart((e) => {
            e.stopPropagation();
            this._draggingMouseDownPos = e.data.getLocalPosition(this.vis._root)
        });

        this.setOnTouchMove((e) => {
            let pos = e.data.getLocalPosition(this.vis._root);
            if (this._draggingMouseDownPos) {
                if (Math.abs(pos.x - this._draggingMouseDownPos.x) >= 5 || Math.abs(pos.y - this._draggingMouseDownPos.y) >= 5) {
                    this.startDragging(this._draggingMouseDownPos)
                }
            }

            if (this.isDragging()) {
                e.stopPropagation();
                this.onDraggingMove(pos)
            }
        });

        this.setOnTouchEnd((e) => {
            let pos = e.data.getLocalPosition(this.vis._root);
            if (this.isDragging()) {
                e.stopPropagation();
                this.stopDragging(pos)
            }
            delete this._draggingMouseDownPos
        })
    }


    isDragging() {
        return typeof this._draggingOffset !== 'undefined'
    }

    startDragging(pos) {
        delete this._draggingMouseDownPos;
        this._draggingOffset = new PIXI.Point(this.position.x - pos.x, this.position.y - pos.y);

        let timerFunc = () => {
            this.onDraggingTimer();
            this._draggingTimer = this.addCallback(0.2, timerFunc)
        };
        this._draggingTimer = this.addCallback(0.2, timerFunc)
    }

    onDraggingMove(pos) {
        this._doDraggingMove(pos)
    }

    clearDraggingStates(e) {
        delete this._draggingMouseDownPos;

        if (this.isDragging()) {
            let pos = e.data.getLocalPosition(this.vis._root);
            this.stopDragging(pos)
        }
    }

    stopDragging(pos) {
        this.cancelCallback(this._draggingTimer);
        delete this._draggingTimer;
        this._doDraggingMove(pos);
        delete this._draggingOffset;
        // apply the drag position at last
        this.onDraggingDone();
    }

    _doDraggingMove(pos) {
        this._root.position = new PIXI.Point(Math.round(this._draggingOffset.x + pos.x), Math.round(this._draggingOffset.y + pos.y))
    }

    onDraggingTimer() {
    }

    onDraggingDone() {
    }
}


class TimerMgr {
    constructor(owner) {
        this.owner = owner;
        this._callbacks = {};
        this._nextCid = 1
    }

    update(dt) {
        for (let cid in this._callbacks) {
            let cb = this._callbacks[cid];
            cb[0] -= dt;
            if (cb[0] <= 0) {
                delete this._callbacks[cid];
                cb[1]()
            }
        }
    }

    addCallback(duration, callback) {
        let cid = this._nextCid;
        this._nextCid += 1;
        this._callbacks[cid] = [duration, callback];
        return cid
    }

    cancelCallback(cid) {
        delete this._callbacks[cid]
    }
}
