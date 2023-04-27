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

import VObject from "./VObject";
import {PAUSE_SPEED} from "./consts";

// A VObject that can be assigned a particular Lifetime, defined as either a virtual time (simulated time)
// or real clock time. This is used for animations that last a particular duration in virtual time and can
// adapt to pausing and speed changes.
export default class LVObject extends VObject {

    constructor() {
        super();
    }

    update(dt) {
        super.update(dt);
        if (this._lifetimeRemaining > 0){
            if (this._lifetimeRealMode) {
                this._lifetimeRemaining -= dt;
            }else{
                if (this.vis.speed == PAUSE_SPEED || this.vis.curTime > this._lastSimTime ) {
                    this._lifetimeRemaining = this._lifetime - (this.vis.curTime - this._startSimTime) / 1000000;
                    this._lastSimTime = this.vis.curTime;
                }else{
                    this._lifetimeRemaining -= dt * this.vis.speed;
                }
            }
            if (this._lifetimeRemaining < 0) {
                this._lifetimeRemaining = 0;
            }
        }
    }

    toString() {
        return "[" + this.id.toString() + "]";
    }

    // set LVObject lifetime to a virtual (simtime) lifetime in us.
    setLifetimeVirtual(dtSimUs) {
        this._startSimTime = this.vis.curTime;
        this._lastSimTime = this._startSimTime;
        this._lifetime = dtSimUs / 1000000;
        this._lifetimeRemaining = this._lifetime;
        this._lifetimeRealMode = false; // virtual-time mode
    }

    // set LVObject lifetime to a real (clock) lifetime in us.
    setLifetimeReal(dtRealUs) {
        this._lifetime = dtRealUs / 1000000;
        this._lifetimeRemaining = this._lifetime;
        this._lifetimeRealMode = true;
    }

    configureLifetime(mvInfo, defaultRealDuration = 700000) {
        if (mvInfo.getVisTrueDuration()) {
            this.setLifetimeVirtual(mvInfo.getSendDurationUs());
        }else{
            this.setLifetimeReal(defaultRealDuration);
        }
    }

    // return real-time equivalent of remaining LVObject lifetime in seconds, or 0 if lifetime is over.
    getRealLifetimeRemaining() {
        if (this._lifetimeRemaining <= 0) {
            return 0;
        }
        if(this._lifetimeRealMode){
            return this._lifetimeRemaining;
        }else{
            return this._lifetimeRemaining / this.vis.speed; // may become Infinity when speed = 0
        }
    }

    // return real-time equivalent of initial (i.e. as set during construction) LVObject lifetime in seconds
    getRealLifetimeInitial() {
        if(this._lifetimeRealMode){
            return this._lifetime;
        }else{
            return this._lifetime / this.vis.speed ; // may become Infinity when speed = 0
        }
    }

    getLifetimeProgress() {
        if (this._lifetimeRemaining <= 0) {
            return 1.0;
        }
        return 1.0 - this._lifetimeRemaining / this._lifetime;
    }

}
