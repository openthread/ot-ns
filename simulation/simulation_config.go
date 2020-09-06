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

package simulation

import "github.com/openthread/ot-ns/threadconst"

const (
	DefaultNetworkName = "OTSIM"
	DefaultMasterKey   = "00112233445566778899aabbccddeeff"
	DefaultPanid       = 0xface
	DefaultChannel     = 11
)

type Config struct {
	NetworkName    string
	MasterKey      string
	Panid          uint16
	Channel        int
	OtCliPath      string
	Speed          float64
	ReadOnly       bool
	RawMode        bool
	Real           bool
	DispatcherHost string
	DispatcherPort int
	DumpPackets    bool
}

func DefaultConfig() *Config {
	return &Config{
		NetworkName:    DefaultNetworkName,
		MasterKey:      DefaultMasterKey,
		Panid:          DefaultPanid,
		Channel:        DefaultChannel,
		Speed:          1,
		ReadOnly:       false,
		RawMode:        false,
		OtCliPath:      "./ot-cli-ftd",
		Real:           false,
		DispatcherHost: "localhost",
		DispatcherPort: threadconst.InitialDispatcherPort,
	}
}
