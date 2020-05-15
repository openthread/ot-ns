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

package dispatcher

import (
	"math/rand"

	"github.com/simonlingoogle/go-simplelogger"
)

// FailTime represents a node fail time configuration.
type FailTime struct {
	FailDuration uint64 // Expected fail duration (us)
	FailInterval uint64 // Expected fail interval (us)
}

// CanFail returns if the node can ever fail using this configuration.
func (ft FailTime) CanFail() bool {
	return ft.FailDuration > 0
}

var (
	// NonFailTime is a fail time configuration that never fail.
	NonFailTime = FailTime{0, 0}
)

type failureCtrl struct {
	owner            *Node
	failTime         FailTime
	recoverTs        uint64
	elapsedTimeAccum uint64
}

func newFailureCtrl(owner *Node, failTime FailTime) *failureCtrl {
	return &failureCtrl{
		owner:    owner,
		failTime: failTime,
	}
}

func (fc *failureCtrl) SetFailTime(failTime FailTime) {
	fc.failTime = failTime
	if !failTime.CanFail() && fc.owner.IsFailed() {
		fc.recoverTs = 0
		fc.owner.Recover()
	}
}

func (fc *failureCtrl) OnTimeAdvanced(oldTime uint64) {
	if !fc.failTime.CanFail() {
		return
	}

	if fc.owner.IsFailed() {
		simplelogger.AssertTrue(fc.recoverTs > 0 && fc.elapsedTimeAccum == 0)
		fc.tryRecoverNode()
		return
	}

	simplelogger.AssertTrue(fc.recoverTs == 0)

	periodTime := fc.failTime.FailInterval
	fc.elapsedTimeAccum += fc.owner.CurTime - oldTime
	for !fc.owner.IsFailed() && fc.elapsedTimeAccum >= periodTime/100 {
		fc.elapsedTimeAccum -= periodTime / 100
		if rand.Float32() < 0.01 {
			// make the node fail
			fc.failNode()
		}
	}
}

func (fc *failureCtrl) tryRecoverNode() {
	simplelogger.AssertTrue(fc.owner.IsFailed())
	if fc.owner.CurTime >= fc.recoverTs {
		fc.recoverTs = 0
		fc.owner.Recover()
	}
}

func (fc *failureCtrl) failNode() {
	simplelogger.AssertTrue(!fc.owner.IsFailed())

	fc.recoverTs = fc.owner.CurTime + fc.failTime.FailDuration
	fc.elapsedTimeAccum = 0

	fc.owner.Fail()
}
