// Copyright (c) 2023-2026, The OTNS Authors.
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
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineExecutableBasedOnConfig(t *testing.T) {
	cfg := ExecutableConfig{
		Ftd:         "my-ftd-fail",
		Mtd:         "ot-cli-mtd",
		Br:          "br-script-404",
		Rcp:         "ot-rcp",
		RcpHost:     "ot-cli",
		SearchPaths: []string{".", "./otrfsim/path/not/found", "../ot-rfsim/ot-versions"},
	}

	// if file could not be located, same name is returned.
	nodeCfg := DefaultNodeConfig()
	exe := cfg.FindExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "my-ftd-fail", exe)

	// test assumes that ot-rfsim MTD has been built.
	nodeCfg.IsMtd = true
	nodeCfg.IsRouter = false
	exe = cfg.FindExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "../ot-rfsim/ot-versions/ot-cli-mtd", exe)

	// test assumes that 'br-script-404' does not exist. In that case, Find returns the plain name
	// and assumes the OS path will be used later on to locate the exe.
	nodeCfg.IsBorderRouter = true
	exe = cfg.FindExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "br-script-404", exe)

	// test assumes that ot-rfsim RCP MAY have been built.
	nodeCfg.IsRcp = true
	nodeCfg.IsBorderRouter = false
	exe = cfg.FindExecutableBasedOnConfig(&nodeCfg)
	if _, err := os.Stat("../ot-rfsim/ot-versions/ot-rcp"); err == nil {
		assert.Equal(t, "../ot-rfsim/ot-versions/ot-rcp", exe)
	} else {
		assert.Equal(t, "ot-rcp", exe)
	}

	// test assumes that ot-rfsim NCP MAY have been built.
	exe = cfg.FindHostExecutableBasedOnConfig(&nodeCfg)
	if _, err := os.Stat("../ot-rfsim/ot-versions/ot-cli"); err == nil {
		assert.Equal(t, "../ot-rfsim/ot-versions/ot-cli", exe)
	} else {
		assert.Equal(t, "ot-cli", exe)
	}

	// Also non-executable files could be supplied. The error comes only later when adding the node type.
	// This test assumes the source file below exists.
	cfg.Ftd = "../simulation/node_config.go"
	nodeCfg.IsMtd = false
	nodeCfg.IsRouter = true
	nodeCfg.IsRcp = false
	exe = cfg.FindExecutableBasedOnConfig(&nodeCfg)
	assert.Equal(t, "../simulation/node_config.go", exe)
}
