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

// Package progctx implements utilities for managing the context of a program.
package progctx

import (
	"context"
	"github.com/pkg/errors"
	"sync"

	"github.com/simonlingoogle/go-simplelogger"
)

// ProgCtx represent the context of a program during it's lifetime.
type ProgCtx struct {
	context.Context // the inner context of the program
	wg              sync.WaitGroup
	cancel          context.CancelFunc
	routinesLock    sync.Mutex
	routines        map[string]int
	deferred        []func()
}

// WaitCount returns the number of goroutines to wait for.
func (ctx *ProgCtx) WaitCount() int {
	ctx.routinesLock.Lock()
	defer ctx.routinesLock.Unlock()

	total := 0
	for _, c := range ctx.routines {
		total += c
	}
	return total
}

// Cancel cancel the program context with a given error.
// It is only effective the first time it's called.
func (ctx *ProgCtx) Cancel(err interface{}) {
	if ctx.Err() != nil {
		return
	}

	defer func() {
		ctx.deferred = nil
	}()

	ctx.cancel()

	if e, ok := err.(error); ok {
		simplelogger.TraceError("program exit: %v", e)
	} else {
		simplelogger.Infof("program exit: %v", err)
	}

	for _, f := range ctx.deferred {
		f()
	}
}

// WaitAdd adds a new goroutine to wait for.
func (ctx *ProgCtx) WaitAdd(name string, delta int) {
	ctx.routinesLock.Lock()
	ctx.routines[name] += delta
	ctx.routinesLock.Unlock()

	ctx.wg.Add(delta)
}

// WaitDone notifies that a goroutine has finished.
func (ctx *ProgCtx) WaitDone(name string) {
	ctx.routinesLock.Lock()
	defer ctx.routinesLock.Unlock()

	count := ctx.routines[name]
	if count <= 0 {
		simplelogger.Panicf("routine %s is not running, should not call WaitDone", name)
	}

	ctx.routines[name] -= 1
	ctx.wg.Done()
}

// Wait waits for all goroutines to finish.
func (ctx *ProgCtx) Wait() {
	ctx.routinesLock.Lock()
	simplelogger.Infof("program context waiting routines: %v", ctx.routines)
	ctx.routinesLock.Unlock()

	ctx.wg.Wait()
}

// Defer registers a function to be called when the program context is cancelled.
// The function will be called when `Cancel` is first called.
func (ctx *ProgCtx) Defer(f func()) {
	if ctx.Err() != nil {
		panic(errors.Errorf("Can not `Defer` after context is done"))
	}

	ctx.deferred = append(ctx.deferred, f)
}

// New creates a new ProgCtx from the parent context.
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
