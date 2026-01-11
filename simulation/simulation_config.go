// Copyright (c) 2020-2026, The OTNS Authors.
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
	"time"

	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/prng"
	. "github.com/openthread/ot-ns/types"
)

// Thread Network defaults
const (
	DefaultChannel         = 11
	DefaultExtPanid        = "dead00beef00cafe"
	DefaultMeshLocalPrefix = "fdde:ad00:beef:0::"
	DefaultNetworkKey      = "00112233445566778899aabbccddeeff"
	DefaultNetworkName     = "otns"
	DefaultPanid           = 0xface
	DefaultPskc            = "3aa55f91ca47d1e4e71a08cb35e91591"
)

type Config struct {
	ExeConfig        ExecutableConfig
	ExeConfigDefault ExecutableConfig
	NewNodeConfig    NodeConfig
	NewNodeScripts   *YamlScriptConfig
	Speed            float64
	SimStepDuration  time.Duration
	ReadOnly         bool
	Realtime         bool
	RtSpeedSmoothing float64 // Used if Realtime=true: 0.0 < value <= 1.0 - determines PID filter tau_f
	RtDriftCorrGain  float64 // Used if Realtime=true: 0.0 < value < ~8.0 - equals PID filter K_p
	AutoGo           bool
	DumpPackets      bool
	DispatcherHost   string
	DispatcherPort   int
	RadioModel       string
	Id               int
	Channel          ChannelId
	LogLevel         logger.Level
	LogFileLevel     logger.Level
	RandomSeed       prng.RandomSeed
	OutputDir        string
}

func DefaultConfig() *Config {
	return &Config{
		ExeConfig:        DefaultExecutableConfig,
		ExeConfigDefault: DefaultExecutableConfig,
		NewNodeConfig:    DefaultNodeConfig(),
		NewNodeScripts:   DefaultNodeScripts(),
		Speed:            1.0,
		SimStepDuration:  time.Second,
		ReadOnly:         false,
		Realtime:         false,
		RtSpeedSmoothing: 0.4,
		RtDriftCorrGain:  3.0,
		AutoGo:           true,
		DumpPackets:      false,
		DispatcherHost:   "localhost",
		DispatcherPort:   InitialDispatcherPort,
		RadioModel:       "MutualInterference",
		Id:               0,
		Channel:          DefaultChannel,
		LogLevel:         logger.WarnLevel,
		LogFileLevel:     logger.DebugLevel,
		RandomSeed:       0,
		OutputDir:        "tmp",
	}
}
