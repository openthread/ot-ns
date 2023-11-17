// Copyright (c) 2023, The OTNS Authors.
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

import {StatsVisualizer} from "./stats/StatsVisualizer";
import NodeNumbersChart from "./stats/nodeNumbersChart";

const {
    VisualizeRequest, VisualizeEvent
} = require('./proto/visualize_grpc_pb.js');
const {VisualizeGrpcServiceClient} = require('./proto/visualize_grpc_grpc_web_pb.js');

let vis = null;
let grpcServiceClient = null;
let lastTimestampUs = 0;

const nodeNumbersChart = new NodeNumbersChart(
    document.getElementById('statsViewer').getContext('2d'),
);

function loadOk() {
    console.log('connecting to server ' + server);
    grpcServiceClient = new VisualizeGrpcServiceClient(server);

    vis = new StatsVisualizer();
    
    let visualizeRequest = new VisualizeRequest();
    let metadata = {'custom-header-1': 'value1'};
    let stream = grpcServiceClient.visualize(visualizeRequest, metadata);
    
    stream.on('data', function (resp) {
        let e = null;
        switch (resp.getTypeCase()) {
            case VisualizeEvent.TypeCase.ADD_NODE:
                e = resp.getAddNode();
                vis.visAddNode(e.getNodeId(), e.getX(), e.getY(), e.getRadioRange());
                break;
            case VisualizeEvent.TypeCase.DELETE_NODE:
                e = resp.getDeleteNode();
                vis.visDeleteNode(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.ON_NODE_FAIL:
                e = resp.getOnNodeFail();
                vis.visOnNodeFail(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.ON_NODE_RECOVER:
                e = resp.getOnNodeRecover();
                vis.visOnNodeRecover(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_PARTITION_ID:
                e = resp.getSetNodePartitionId();
                vis.visSetNodePartitionId(e.getNodeId(), e.getPartitionId());
                break;
            case VisualizeEvent.TypeCase.ADVANCE_TIME:
                e = resp.getAdvanceTime();
                lastTimestampUs = e.getTs();
                vis.visAdvanceTime(lastTimestampUs, e.getSpeed());
                const [aTs,aStat] = vis.getNewDataPoints();
                for( let i in aTs) {
                    nodeNumbersChart.addData(aTs[i], aStat[i]);
                }
                if (aTs.length > 0) {
                    nodeNumbersChart.update(lastTimestampUs);
                }
                break;
            case VisualizeEvent.TypeCase.HEARTBEAT:
                e = resp.getHeartbeat();
                vis.visHeartbeat();
                nodeNumbersChart.update(lastTimestampUs);
                break;
            case VisualizeEvent.TypeCase.SET_NODE_ROLE:
                e = resp.getSetNodeRole();
                vis.visSetNodeRole(e.getNodeId(), e.getRole());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_MODE:
                e = resp.getSetNodeMode();
                vis.visSetNodeMode(e.getNodeId(), e.getNodeMode());
                break;
            default:
                break;
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
