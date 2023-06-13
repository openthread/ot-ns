// Copyright (c) 2023, The OTNS Authors.
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

package simulation

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openthread/ot-ns/types"
)

func TestDetermineExecutableBasedOnConfig(t *testing.T) {
	cfg := ExecutableConfig{
		Ftd:         "my-ftd-fail",
		Mtd:         "ot-cli-mtd",
		Br:          "br-script",
		SearchPaths: []string{".", "./otrfsim/path/not/found", "../ot-rfsim/build/bin"},
	}

	// if file could not be located, special name is returned.
	nodeCfg := types.DefaultNodeConfig()
	exe := cfg.DetermineExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "./EXECUTABLE-NOT-FOUND", exe)

	// test assumes that ot-rfsim has been built.
	nodeCfg.IsMtd = true
	nodeCfg.IsRouter = false
	exe = cfg.DetermineExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "../ot-rfsim/build/bin/ot-cli-mtd", exe)

	// test assumes that ot-rfsim has been built.
	cfg.Mtd = "./ot-cli-mtd"
	exe = cfg.DetermineExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "../ot-rfsim/build/bin/ot-cli-mtd", exe)

	// Also non-executable files could be supplied. The error comes only later when adding the node type.
	cfg.Ftd = "../simulation/node_config.go"
	nodeCfg.IsMtd = false
	nodeCfg.IsRouter = true
	exe = cfg.DetermineExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "../simulation/node_config.go", exe)
}
