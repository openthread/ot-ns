// Copyright (c) 2020-2024, The OTNS Authors.
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
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/prng"
)

type FailTime struct {
	FailDuration uint64 // unit: us
	FailInterval uint64 // unit: us
}

var (
	NonFailTime = FailTime{0, 0}
)

type FailureCtrl struct {
	owner           *Node
	failTime        FailTime
	recoverTs       uint64 // unit: us; timestamp when recovery from failure starts (valid if currently failed)
	failTs          uint64 // unit: us; timestamp when failure starts (valid if currently not failed)
	remainTm        uint64 // unit: us; time that remains in this fail cycle after failure has ended.
	prevOpTimestamp uint64 // unit: us; time of previous reported next-operation timestamp.
}

func newFailureCtrl(owner *Node, failTime FailTime) *FailureCtrl {
	fc := &FailureCtrl{
		owner:    owner,
		failTime: failTime,
	}
	return fc
}

func (ft FailTime) CanFail() bool {
	return ft.FailDuration > 0
}

func (fc *FailureCtrl) SetFailTime(failTime FailTime) {
	fc.failTime = failTime

	fc.recoverTs = 0
	fc.failTs = 0
	fc.remainTm = 0
	if !failTime.CanFail() && fc.owner.IsFailed() {
		fc.owner.Recover()
	}
	fc.calcNextFailTimestamp()
}

// OnTimeAdvanced must be called when the node's time advances. It performs fail/recover operations as
// needed, and returns the timestamp of a next expected fail/recover operation as well as a flag
// that is set to 'true' if the next-operation timestamp moved further into the future.
func (fc *FailureCtrl) OnTimeAdvanced(oldTime uint64) (uint64, bool) {
	isUpdated := false
	if !fc.failTime.CanFail() {
		return Ever, isUpdated
	}

	logger.AssertTrue(fc.owner.CurTime > oldTime)

	// if node is failed currently
	if fc.owner.IsFailed() {
		logger.AssertTrue(fc.failTs == 0)
		if fc.owner.CurTime >= fc.recoverTs {
			fc.recoverTs = 0
			fc.calcNextFailTimestamp()
			fc.owner.Recover()
			fc.prevOpTimestamp = fc.failTs
			return fc.failTs, true
		}
		isUpdated = fc.recoverTs > fc.prevOpTimestamp
		fc.prevOpTimestamp = fc.recoverTs
		return fc.recoverTs, isUpdated
	}

	// if node is not failed currently
	logger.AssertTrue(fc.recoverTs == 0)
	if fc.owner.CurTime >= fc.failTs {
		fc.recoverTs = fc.owner.CurTime + fc.failTime.FailDuration
		fc.failTs = 0
		fc.owner.Fail()
		fc.prevOpTimestamp = fc.recoverTs
		return fc.recoverTs, true
	}
	isUpdated = fc.failTs > fc.prevOpTimestamp
	fc.prevOpTimestamp = fc.failTs
	return fc.failTs, isUpdated
}

func (fc *FailureCtrl) calcNextFailTimestamp() {
	if !fc.failTime.CanFail() {
		return
	}
	logger.AssertTrue(fc.failTime.FailDuration > 0 && fc.failTime.FailInterval > fc.failTime.FailDuration)
	failStartTimeMax := int(fc.failTime.FailInterval - fc.failTime.FailDuration)
	failTsRel := prng.NewFailTime(failStartTimeMax)
	fc.failTs = failTsRel + fc.owner.CurTime + fc.remainTm
	fc.remainTm = fc.failTime.FailInterval - fc.failTime.FailDuration - failTsRel
	logger.AssertTrue(fc.remainTm < fc.failTime.FailInterval)
}
