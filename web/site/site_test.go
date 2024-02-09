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

package web_site

import (
	"mime"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMimeTypes(t *testing.T) {
	jsMimeType := mime.TypeByExtension(".js")
	if jsMimeType == "application/javascript" {
		assert.Equal(t, "application/javascript", jsMimeType)
	} else {
		assert.Equal(t, "text/javascript; charset=utf-8", jsMimeType)
	}
}

func TestServe(t *testing.T) {
	go func() {
		_ = Serve("localhost:8997")
	}()
	deadline := time.Now().Add(time.Second * 5)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://localhost:8997/static/")
		if err != nil || resp.StatusCode != 200 || resp.ContentLength <= 0 {
			time.Sleep(time.Millisecond * 100)
		} else {
			break // once server works, quickly proceed to tests below.
		}
	}

	resp, err := http.Get("http://localhost:8997/static/")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.True(t, resp.ContentLength > 0)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("content-type"))

	resp, err = http.Get("http://localhost:8997/visualize?addr=")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.True(t, resp.ContentLength > 0)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("content-type"))

	resp, err = http.Get("http://localhost:8997/energyViewer?addr=")
	assert.Nil(t, err)
	assert.Equal(t, resp.StatusCode, 200)
	assert.True(t, resp.ContentLength > 0)
	assert.Equal(t, "text/html; charset=utf-8", resp.Header.Get("content-type"))
}
