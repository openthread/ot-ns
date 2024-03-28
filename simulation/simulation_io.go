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

package simulation

import (
	"fmt"

	"github.com/openthread/ot-ns/logger"
)

// ExportNetwork exports config info of network to a YAML-friendly object.
func (s *Simulation) ExportNetwork() YamlNetworkConfig {
	var rr *int = nil

	// include radio-range if non-default
	if s.cfg.NewNodeConfig.RadioRange != DefaultNodeConfig().RadioRange {
		rr = &s.cfg.NewNodeConfig.RadioRange
	}
	res := YamlNetworkConfig{
		Position:   [3]int{0, 0, 0}, // when exporting, always a 0-offset is used.
		RadioRange: rr,
	}
	return res
}

// ExportNodes exports config/position info of all nodes to a YAML-friendly object.
func (s *Simulation) ExportNodes(nwConfig *YamlNetworkConfig) []YamlNodeConfig {
	nodes := s.GetNodes()
	res := make([]YamlNodeConfig, 0)

	for _, nodeId := range nodes {
		node := s.nodes[nodeId]
		dNode := s.Dispatcher().GetNode(nodeId)
		var rr *int = nil
		var ver *string = nil

		// include radio-range if non-default
		if (nwConfig.RadioRange != nil && node.cfg.RadioRange != *nwConfig.RadioRange) ||
			(nwConfig.RadioRange == nil && node.cfg.RadioRange != defaultRadioRange) {
			rr = &node.cfg.RadioRange
		}

		// include version if non-empty
		if len(node.cfg.Version) > 0 {
			ver = &node.cfg.Version
		}

		cfg := YamlNodeConfig{
			ID:         nodeId,
			Type:       node.cfg.Type,
			Position:   [3]int{dNode.X, dNode.Y, dNode.Z},
			RadioRange: rr,
			Version:    ver,
		}
		res = append(res, cfg)
	}
	return res
}

func (s *Simulation) ImportNodes(nwConfig YamlNetworkConfig, nodes []YamlNodeConfig) error {
	allOk := true
	rr := defaultRadioRange
	if nwConfig.RadioRange != nil {
		rr = *nwConfig.RadioRange
	}
	posOffset := nwConfig.Position
	nodeIdOffset := 0
	if nwConfig.BaseId != nil {
		nodeIdOffset = *nwConfig.BaseId
	}

	for _, node := range nodes {
		cfg := DefaultNodeConfig()

		// fill config with entries from YAML 'node'
		cfg.ID = node.ID + nodeIdOffset
		if node.RadioRange != nil {
			cfg.RadioRange = *node.RadioRange
		} else {
			cfg.RadioRange = rr
		}
		cfg.IsAutoPlaced = false
		cfg.X = node.Position[0] + posOffset[0]
		cfg.Y = node.Position[1] + posOffset[1]
		cfg.Z = node.Position[2] + posOffset[2]
		cfg.Type = node.Type
		if node.Version != nil {
			cfg.Version = *node.Version
		}

		s.NodeConfigFinalize(&cfg)
		_, err := s.AddNode(&cfg)
		if err != nil {
			logger.Warnf("Warn: %s", err)
			allOk = false // continue trying to import remaining nodes
		}
	}

	if !allOk {
		return fmt.Errorf("not all nodes could be imported - see error log above")
	}
	return nil
}
