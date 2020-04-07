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

package progctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

func TestNew(t *testing.T) {
	ctx := New(context.Background())
	_ = context.Context(ctx)  // ProgCtx should implement context.Context
	ctx2 := New(nil)          // nolint
	_ = context.Context(ctx2) // ProgCtx should implement context.Context
}

func TestProgCtx_Cancel(t *testing.T) {
	ctx := New(context.Background())
	err := errors.Errorf("test error")
	ctx.Cancel(err)
	go func() {
		err := errors.Errorf("test error")
		ctx.Cancel(err)
		assert.True(t, ctx.Err() == context.Canceled)
		<-ctx.Done()
	}()
	<-ctx.Done()
}

func TestProgCtx_CancelNilError(t *testing.T) {
	ctx := New(context.Background())
	ctx.Cancel(nil)
	go func() {
		err := errors.Errorf("test error")
		ctx.Cancel(err)
		assert.True(t, ctx.Err() == context.Canceled)
		<-ctx.Done()
	}()
	<-ctx.Done()
}

func TestProgCtx_Wait(t *testing.T) {
	ctx := New(context.Background())
	ctx.WaitAdd("test1", 1)
	go func() {
		ctx.WaitDone("test1")
	}()

	ctx.WaitAdd("test2", 2)
	for i := 0; i < 2; i++ {
		go func() { defer ctx.WaitDone("test2") }()
	}

	ctx.WaitAdd("test3", 3)
	for i := 0; i < 3; i++ {
		go func() { defer ctx.WaitDone("test3") }()
	}

	ctx.Wait()
}
