// Copyright (c) 2022-2023, The OTNS Authors.
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

import EnergyConsumptionChart from './energy/energyConsChart.js';
import NodeEnergyConsumptionBarChart from './energy/nodeEnergyConsBarChart.js';
import PowerConsumptionChart from './energy/powerConsChart.js';

const {
    VisualizeRequest,
} = require('./proto/visualize_grpc_pb.js');
const {VisualizeGrpcServiceClient} = require('./proto/visualize_grpc_grpc_web_pb.js');

//For linear charts
const decimationConfig = {
    algorithm: 'lttb',
    enabled: true,
    samples: 200,
    threshold: 200,
  };


class NodeEnergy {
    constructor(id, disabled, sleep, tx, rx) {
        this.id = id;
        this.disabled = disabled;
        this.sleep = sleep;
        this.tx = tx;
        this.rx = rx;
    }

    getTotal() {
        return this.disabled + this.sleep + this.tx + this.rx;
    }
}

let grpcServiceClient = null;
let gNodesOverTime = [];
let gTimes = [];

const energyCons = new EnergyConsumptionChart(
    document.getElementById('energyCons').getContext('2d'),
    decimationConfig
);

const energyByNode = new NodeEnergyConsumptionBarChart(
    document.getElementById('energyByNode').getContext('2d'),
    decimationConfig
);

const powerCons = new PowerConsumptionChart(
    document.getElementById('powerCons').getContext('2d'),
    decimationConfig
);

//insert functions into the respective DOM's objects
document.getElementById('nodesForEnergy').onchange = () => {
    energyCons.reset();
    energyCons.setData(gNodesOverTime, gTimes).update();
};
document.getElementById('nodesForPower').onchange = () => {
    powerCons.reset();
    powerCons.setData(gNodesOverTime,gTimes).update();
};

//Keep same format as the other js files for future use
function loadOk() {
    console.log('connecting to server ' + server);
    grpcServiceClient = new VisualizeGrpcServiceClient(server);

    
    let visualizeRequest = new VisualizeRequest();
    let metadata = {'custom-header-1': 'value1'};
    let stream = grpcServiceClient.energyReport(visualizeRequest, metadata);
    
    
    //insert nodes into the select of the html
    let selectEnergy = document.getElementById('nodesForEnergy');
    let option = document.createElement('option');
    option.value = "all";
    option.text = "Average of all nodes";
    selectEnergy.add(option);

    stream.on('data', function (resp) {
        const timestamp = resp.getTimestamp();
        const nodes = resp.getNodesenergyList();
        
        //If timestamp is maxed and nodes are null, it mean end of transmission, so plot.
        if (nodes.length == 0) {
            updateView();
        }
        else {
            let nodesMap = new Map();
            for (let i = 0; i < nodes.length; i++) {
                const node = nodes[i];
                const nodeEnergy = new NodeEnergy(node.getNodeId(), node.getDisabled(), node.getSleep(), node.getTx(), node.getRx());
                nodesMap.set(node.getNodeId(), nodeEnergy);
            }
            gTimes.push(timestamp);
            gNodesOverTime.push(nodesMap);
            
            energyCons.addNodesPoint(nodesMap, timestamp);
            powerCons.addNodesPoint(nodesMap, timestamp);
        }
    });

    stream.on('status', function (status) {
        console.log('Status code: ' + status.code);
    });
    stream.on('end', function (end) {
        // stream end signal
        console.log('Connection ended');
    });
}
loadOk();

function hasElement(selector, nodeId) {
    for (let i = 0; i < selector.options.length; i++) {
        if (parseInt(selector.options[i].value) == nodeId) {
            return true;
        }
    }
    return false;
}

function sortInsert(sel, opt) {
      let i = 0;
      for (; i < sel.options.length; i++) {
          if (sel.options[i].value == "all") {
              continue;
          }
          if (sel.options[i].value > opt.value) {
              break;
          }
      }
      sel.insertBefore(opt, sel.options[i]);
  }

function updateView() {
    //insert nodes into the select of the html
    let selectEnergy = document.getElementById('nodesForEnergy');
    let selectPower = document.getElementById('nodesForPower');

    let nodes = gNodesOverTime[gNodesOverTime.length - 1];
    for (let key of nodes.keys()) {
        let option = document.createElement('option');
        option.value = key;
        option.text = "Node " + key;
        if (!hasElement(selectEnergy,key)) {
            sortInsert(selectEnergy,option);
        }
        if (!hasElement(selectPower,key)) {
            sortInsert(selectPower,option.cloneNode(true));
        }
    }
    for (let i = selectEnergy.options.length - 1; i >= 0; i--) {
        if (!nodes.has(parseInt(selectEnergy.options[i].value)) && selectEnergy.options[i].value != "all") {
            selectEnergy.remove(i);
        }
    }
    for (let i = selectPower.options.length - 1; i >= 0; i--) {
        if (!nodes.has(parseInt(selectPower.options[i].value))) {
            selectPower.remove(i);
        }
    }

    energyCons.update();
    energyByNode.setData(gNodesOverTime[gNodesOverTime.length - 1]).update();
    powerCons.update();
}
