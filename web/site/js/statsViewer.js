// Copyright (c) 2023-2024, The OTNS Authors.
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
    NodeStatsRequest, VisualizeEvent
} = require('./proto/visualize_grpc_pb.js');
const {VisualizeGrpcServiceClient} = require('./proto/visualize_grpc_grpc_web_pb.js');

let vis = null;
let grpcServiceClient = null;
let ts = 0;

const nodeNumbersChart = new NodeNumbersChart(
    document.getElementById('statsViewer').getContext('2d'),
);

function loadOk() {
    console.log('connecting to server ' + server);
    grpcServiceClient = new VisualizeGrpcServiceClient(server);

    vis = new StatsVisualizer();
    
    let nodeStatsRequest = new NodeStatsRequest();
    let metadata = {} // {'custom-header-1': 'value1'};
    let stream = grpcServiceClient.nodeStats(nodeStatsRequest, metadata);
    
    stream.on('data', function (resp) {
        let e = null;
        switch (resp.getTypeCase()) {
            case VisualizeEvent.TypeCase.NODE_STATS_INFO:
                e = resp.getNodeStatsInfo()
                ts = e.getTimestamp();
                vis.visAdvanceTime(ts);
                vis.visNodeStatsInfo(ts, e.getNodeStats());
                const [aTs,aStat] = vis.getNewDataPoints();
                for( let i in aTs) {
                    nodeNumbersChart.addData(aTs[i], aStat[i]);
                }
                nodeNumbersChart.update(ts);
                break;
            case VisualizeEvent.TypeCase.ADVANCE_TIME:
                e = resp.getAdvanceTime();
                ts = e.getTimestamp();
                vis.visAdvanceTime(ts);
                break;
            case VisualizeEvent.TypeCase.HEARTBEAT:
                vis.visHeartbeat();
                nodeNumbersChart.update(ts);
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
