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

package types

// NodeConfig is a generic config for a new simulated node (used in dispatcher, simulation, radiomodel,
// ... packages).
type NodeConfig struct {
	ID             int
	Type           string // Type as requested on creation (router, sed, fed, br, etc.)
	Version        string // Thread version string or "" for default
	X, Y, Z        int
	IsAutoPlaced   bool
	IsRaw          bool // A raw node skips all initialization CLI commands, including any init-script.
	IsMtd          bool
	IsRouter       bool
	IsBorderRouter bool
	IsOTBR		   bool
	RxOffWhenIdle  bool
	NodeLogFile    bool
	RadioRange     int
	ExecutablePath string // executable full path or "" for auto-determined
	Restore        bool
	InitScript     []string // a sequence of CLI commands executed at first startup of node
	RandomSeed     int32
}

// UpdateNodeConfigFromType sets NodeConfig flags correctly, based on chosen node type cfg.Type
func (cfg *NodeConfig) UpdateNodeConfigFromType() {
	switch cfg.Type {
	case ROUTER, REED, FTD:
		cfg.IsRouter = true
		cfg.IsMtd = false
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = false
	case FED:
		cfg.IsRouter = false
		cfg.IsMtd = false
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = false
	case MED, MTD:
		cfg.IsRouter = false
		cfg.IsMtd = true
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = false
	case SED, SSED:
		cfg.IsRouter = false
		cfg.IsMtd = true
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = true
		cfg.IsOTBR = false
	case BR:
		cfg.IsRouter = true
		cfg.IsMtd = false
		cfg.IsBorderRouter = true
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = false
	case OTBR:
		cfg.IsRouter = true
		cfg.IsMtd = false
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = true
	case WIFI:
		cfg.IsRouter = true
		cfg.IsMtd = false
		cfg.IsBorderRouter = false
		cfg.RxOffWhenIdle = false
		cfg.IsOTBR = false

	default:
		panic("unknown node type cfg.Type")
	}
}
