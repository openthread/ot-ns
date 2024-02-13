// Copyright (c) 2024, The OTNS Authors.
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

package cli

import (
	"testing"

	"github.com/openthread/ot-ns/simulation"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

var testYamlArray = `
[4,5,6]
`

var testYamlFile = `
network:
    pos-shift: [0, 0, 0]
nodes:
    - id: 1
      type: router
      pos: [100, 100, 0]
    - id: 2
      type: fed
      pos: [60, 150, 0]
    - id: 3
      type: med
      version: v11
      pos: [100, 150, 0]
    - id: 4
      type: router
      version: v12
      pos: [200, 100, 0]
`

func TestYamlArrayUnmarshall(t *testing.T) {
	myArray := [3]int{0, 0, 0}
	err := yaml.Unmarshal([]byte(testYamlArray), &myArray)
	assert.Nil(t, err)
	assert.Equal(t, 4, myArray[0])
	assert.Equal(t, 5, myArray[1])
	assert.Equal(t, 6, myArray[2])
}

func TestYamlConfigUnmarshall(t *testing.T) {
	cfgFile := simulation.YamlConfigFile{}
	err := yaml.Unmarshal([]byte(testYamlFile), &cfgFile)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(cfgFile.NetworkConfig.Position))
	assert.Equal(t, 4, len(cfgFile.NodesList))
	assert.Equal(t, "v11", *cfgFile.NodesList[2].Version)
}
