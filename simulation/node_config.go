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
	"os"
	"path/filepath"
	"strings"

	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

type ExecutableConfig struct {
	Ftd         string
	Mtd         string
	Br          string
	SearchPaths []string
}

type NodeAutoPlacer struct {
	X, Y            int
	Xref, Yref      int
	Xmax            int
	NodeDeltaCoarse int
	NodeDeltaFine   int
	fineCount       int
	isReset         bool
}

var DefaultExecutableConfig ExecutableConfig = ExecutableConfig{
	Ftd:         "ot-cli-ftd",
	Mtd:         "ot-cli-ftd",
	Br:          "ot-br.sh",
	SearchPaths: []string{".", "./ot-rfsim/ot-versions"},
}

func (cfg *ExecutableConfig) SearchPathsString() string {
	s := "["
	simplelogger.AssertTrue(len(cfg.SearchPaths) >= 1)
	for _, sp := range cfg.SearchPaths {
		s += "\"" + sp + "\", "
	}
	return s[0:len(s)-2] + "]"
}

// GetExecutableForThreadVersion gets the prebuilt executable for given Thread version string as in cli.ThreadVersion
func GetExecutableForThreadVersion(version string) string {
	simplelogger.AssertTrue(strings.HasPrefix(version, "v1") && len(version) == 3)
	return "ot-rfsim/ot-versions/ot-cli-ftd_" + version
}

func isFile(exePath string) bool {
	if _, err := os.Stat(exePath); err == nil {
		return true
	}
	return false
}

func (cfg *ExecutableConfig) DetermineExecutableBasedOnConfig(nodeCfg *NodeConfig) string {
	exeName := cfg.Ftd
	if nodeCfg.IsMtd {
		exeName = cfg.Mtd
	}
	if nodeCfg.IsBorderRouter {
		exeName = cfg.Br
	}

	if filepath.IsAbs(exeName) {
		return exeName
	}

	// if not found directly, it means it's just a name that needs to be located in our search paths.
	for _, sp := range cfg.SearchPaths {
		exePath := filepath.Join(sp, exeName)
		if isFile(exePath) {
			if filepath.IsAbs(exePath) || exePath[0] == '.' {
				return exePath
			}
			return "./" + exePath
		}
	}
	return "./EXECUTABLE-NOT-FOUND"
}

func NewNodeAutoPlacer() *NodeAutoPlacer {
	return &NodeAutoPlacer{
		Xref:            100,
		Yref:            100,
		Xmax:            1500,
		X:               100,
		Y:               100,
		NodeDeltaCoarse: 100,
		NodeDeltaFine:   40,
		fineCount:       0,
		isReset:         true,
	}
}

// UpdateXReference updates the reference X position of the NodeAutoPlacer to 'x'. It starts placing from there.
func (nap *NodeAutoPlacer) UpdateXReference(x int) {
	nap.Xref = x
	nap.X = x
}

// UpdateYReference updates the reference Y position of the NodeAutoPlacer to 'y'. It starts placing from there.
func (nap *NodeAutoPlacer) UpdateYReference(y int) {
	nap.Yref = y
	nap.Y = y
}

// UpdateReference updates the reference position of the NodeAutoPlacer to 'x', 'y'. It starts placing from there.
func (nap *NodeAutoPlacer) UpdateReference(x, y int) {
	nap.Xref = x
	nap.X = x
	nap.Yref = y
	nap.Y = y
}

// NextNodePosition lets the autoplacer pick the next position for a new node to be placed.
func (nap *NodeAutoPlacer) NextNodePosition(isBelowParent bool) (int, int) {
	var x, y int
	if isBelowParent {
		y = nap.Y + nap.NodeDeltaCoarse/2
		x = nap.X + nap.fineCount*nap.NodeDeltaFine - nap.NodeDeltaFine
		nap.fineCount++
	} else {
		if !nap.isReset {
			nap.X += nap.NodeDeltaCoarse
			if nap.X > nap.Xmax {
				nap.X = nap.Xref
				nap.Y += nap.NodeDeltaCoarse
			}
		}
		nap.isReset = false
		nap.fineCount = 0
		x = nap.X
		y = nap.Y
	}
	return x, y
}

// ReuseNextNodePosition instructs the autoplacer to re-use the NextNodePosition() that was given out in the
// last call to this method.
func (nap *NodeAutoPlacer) ReuseNextNodePosition() {
	nap.isReset = true
}
