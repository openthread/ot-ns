// Copyright (c) 2022, The OTNS Authors.
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

package energy

import (
	"fmt"
	"os"
	"sort"

	"github.com/openthread/ot-ns/logger"
	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
)

type EnergyAnalyser struct {
	nodes                map[int]*NodeEnergy
	networkHistory       []NetworkConsumption
	energyHistoryByNodes [][]*pb.NodeEnergy
	title                string
}

func (e *EnergyAnalyser) AddNode(nodeID int, timestamp uint64) {
	if _, ok := e.nodes[nodeID]; ok {
		return
	}
	e.nodes[nodeID] = newNode(nodeID, timestamp)
}

func (e *EnergyAnalyser) DeleteNode(nodeID int) {
	delete(e.nodes, nodeID)

	if len(e.nodes) == 0 {
		e.ClearEnergyData()
	}
}

func (e *EnergyAnalyser) GetNode(nodeID int) *NodeEnergy {
	return e.nodes[nodeID]
}

func (e *EnergyAnalyser) GetNetworkEnergyHistory() []NetworkConsumption {
	return e.networkHistory
}

func (e *EnergyAnalyser) GetEnergyHistoryByNodes() [][]*pb.NodeEnergy {
	return e.energyHistoryByNodes
}

func (e *EnergyAnalyser) GetLatestEnergyOfNodes() []*pb.NodeEnergy {
	return e.energyHistoryByNodes[len(e.energyHistoryByNodes)-1]
}

func (e *EnergyAnalyser) StoreNetworkEnergy(timestamp uint64) {
	nodesEnergySnapshot := make([]*pb.NodeEnergy, 0, len(e.nodes))
	networkSnapshot := NetworkConsumption{
		Timestamp: timestamp,
	}

	netSize := float64(len(e.nodes))
	for _, node := range e.nodes {
		node.ComputeRadioState(timestamp)

		e := &pb.NodeEnergy{
			NodeId:   int32(node.nodeId),
			Disabled: float64(node.radio.SpentDisabled) * RadioDisabledConsumption,
			Sleep:    float64(node.radio.SpentSleep) * RadioSleepConsumption,
			Tx:       float64(node.radio.SpentTx) * RadioTxConsumption,
			Rx:       float64(node.radio.SpentRx) * RadioRxConsumption,
		}

		networkSnapshot.EnergyConsDisabled += e.Disabled / netSize
		networkSnapshot.EnergyConsSleep += e.Sleep / netSize
		networkSnapshot.EnergyConsTx += e.Tx / netSize
		networkSnapshot.EnergyConsRx += e.Rx / netSize
		nodesEnergySnapshot = append(nodesEnergySnapshot, e)
	}

	e.networkHistory = append(e.networkHistory, networkSnapshot)
	e.energyHistoryByNodes = append(e.energyHistoryByNodes, nodesEnergySnapshot)
}

func (e *EnergyAnalyser) SaveEnergyDataToFile(name string, timestamp uint64) {
	if name == "" {
		if e.title == "" {
			name = "energy"
		} else {
			name = e.title
		}
	}

	//Get current directory and add name to the path
	dir, _ := os.Getwd()

	//create "energy_results" directory if it does not exist
	if _, err := os.Stat(dir + "/energy_results"); os.IsNotExist(err) {
		err := os.Mkdir(dir+"/energy_results", 0777)
		if err != nil {
			logger.Error("Failed to create energy_results directory")
			return
		}
	}

	path := fmt.Sprintf("%s/energy_results/%s", dir, name)
	fileNodes, err := os.Create(path + "_nodes.txt")
	if err != nil {
		logger.Errorf("Error creating file: %v", err)
		return
	}
	defer fileNodes.Close()

	fileNetwork, err := os.Create(path + ".txt")
	if err != nil {
		logger.Errorf("Error creating file: %v", err)
		return
	}
	defer fileNetwork.Close()

	//Save all nodes' energy data to file
	e.writeEnergyByNodes(fileNodes, timestamp)

	//Save network energy data to file (timestamp converted to milliseconds)
	e.writeNetworkEnergy(fileNetwork, timestamp)
}

func (e *EnergyAnalyser) writeEnergyByNodes(fileNodes *os.File, timestamp uint64) {
	fmt.Fprintf(fileNodes, "Duration of the simulated network (in milliseconds): %d\n", timestamp/1000)
	fmt.Fprintf(fileNodes, "ID\tDisabled (mJ)\tIdle (mJ)\tTransmiting (mJ)\tReceiving (mJ)\n")

	sortedNodes := make([]int, 0, len(e.nodes))
	for id := range e.nodes {
		sortedNodes = append(sortedNodes, id)
	}
	sort.Ints(sortedNodes)

	for _, id := range sortedNodes {
		node := e.nodes[id]
		fmt.Fprintf(fileNodes, "%d\t%f\t%f\t%f\t%f\n",
			id,
			float64(node.radio.SpentDisabled)*RadioDisabledConsumption,
			float64(node.radio.SpentSleep)*RadioSleepConsumption,
			float64(node.radio.SpentTx)*RadioTxConsumption,
			float64(node.radio.SpentRx)*RadioRxConsumption,
		)
	}
}

func (e *EnergyAnalyser) writeNetworkEnergy(fileNetwork *os.File, timestamp uint64) {
	fmt.Fprintf(fileNetwork, "Duration of the simulated network (in milliseconds): %d\n", timestamp/1000)
	fmt.Fprintf(fileNetwork, "Time (ms)\tDisabled (mJ)\tIdle (mJ)\tTransmiting (mJ)\tReceiving (mJ)\n")
	for _, snapshot := range e.networkHistory {
		fmt.Fprintf(fileNetwork, "%d\t%f\t%f\t%f\t%f\n",
			snapshot.Timestamp/1000,
			snapshot.EnergyConsDisabled,
			snapshot.EnergyConsSleep,
			snapshot.EnergyConsTx,
			snapshot.EnergyConsRx,
		)
	}
}

func (e *EnergyAnalyser) ClearEnergyData() {
	logger.Debugf("Node's energy data cleared")
	e.networkHistory = make([]NetworkConsumption, 0, 3600)
	e.energyHistoryByNodes = make([][]*pb.NodeEnergy, 0, 3600)
}

func (e *EnergyAnalyser) SetTitle(title string) {
	e.title = title
}

func NewEnergyAnalyser() *EnergyAnalyser {
	ea := &EnergyAnalyser{
		nodes:                make(map[int]*NodeEnergy),
		networkHistory:       make([]NetworkConsumption, 0, 3600), //Start with space for 1 sample every 30s for 1 hour = 1*60*60/30 = 3600 samples
		energyHistoryByNodes: make([][]*pb.NodeEnergy, 0, 3600),
	}
	return ea
}
