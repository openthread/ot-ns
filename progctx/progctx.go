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

package progctx

import (
	"context"
	"sync"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/logger"
)

type ProgCtx struct {
	context.Context
	wg           sync.WaitGroup
	cancel       context.CancelFunc
	routinesLock sync.Mutex
	routines     map[string]int
	deferred     []func()
}

func (ctx *ProgCtx) WaitCount() int {
	ctx.routinesLock.Lock()
	defer ctx.routinesLock.Unlock()

	total := 0
	for _, c := range ctx.routines {
		total += c
	}
	return total
}

func (ctx *ProgCtx) Cancel(err interface{}) {
	if ctx.Err() != nil {
		return
	}

	defer func() {
		ctx.deferred = nil
	}()

	ctx.cancel()

	if e, ok := err.(error); ok {
		logger.TraceError("program exit requested with ctx error: %v", e)
	} else {
		logger.Debugf("program exit requested without ctx error: %v", err)
	}

	for _, f := range ctx.deferred {
		f()
	}
}

func (ctx *ProgCtx) WaitAdd(name string, delta int) {
	ctx.routinesLock.Lock()
	ctx.routines[name] += delta
	ctx.routinesLock.Unlock()

	ctx.wg.Add(delta)
}

func (ctx *ProgCtx) WaitDone(name string) {
	ctx.routinesLock.Lock()
	defer ctx.routinesLock.Unlock()

	count := ctx.routines[name]
	if count <= 0 {
		logger.Panicf("routine %s is not running, should not call WaitDone", name)
	}

	ctx.routines[name] -= 1
	ctx.wg.Done()
}

func (ctx *ProgCtx) Wait() {
	ctx.routinesLock.Lock()
	logger.Debugf("program context waiting routines: %v", ctx.routines)
	ctx.routinesLock.Unlock()

	ctx.wg.Wait()
}

func (ctx *ProgCtx) Defer(f func()) {
	if ctx.Err() != nil {
		panic(errors.Errorf("Can not `Defer` after context is done"))
	}

	ctx.deferred = append(ctx.deferred, f)
}

func New(parent context.Context) *ProgCtx {
	if parent == nil {
		parent = context.Background()
	}

	ctx, cancel := context.WithCancel(parent)

	return &ProgCtx{
		Context:  ctx,
		wg:       sync.WaitGroup{},
		cancel:   cancel,
		routines: map[string]int{},
	}
}
