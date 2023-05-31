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

package simulation

import (
	"strings"

	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

type ExecutableConfig struct {
	Ftd string
	Mtd string
	Br  string
}

var DefaultExecutableConfig ExecutableConfig = ExecutableConfig{
	Ftd: "./ot-cli-ftd",
	Mtd: "./ot-cli-ftd",
	Br:  "./otbr-sim.sh",
}

// GetExecutableForThreadVersion gets the prebuilt executable for given Thread version string as in cli.ThreadVersion
func GetExecutableForThreadVersion(version string) string {
	simplelogger.AssertTrue(strings.HasPrefix(version, "v1") && len(version) == 3)
	return "./ot-rfsim/ot-versions/ot-cli-ftd_" + version
}

func DetermineExecutableBasedOnConfig(nodeCfg *NodeConfig, executableCfg *ExecutableConfig) string {
	if nodeCfg.IsRouter {
		return executableCfg.Ftd
	}
	if nodeCfg.IsMtd {
		return executableCfg.Mtd
	}
	if nodeCfg.IsBorderRouter {
		return executableCfg.Br
	}
	// FED or other type.
	return executableCfg.Ftd
}
