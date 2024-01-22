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

import PixiVisualizer from "./vis/PixiVisualizer";
import * as PIXI from 'pixi.js'
import {SetResources} from "./vis/resources";
import {StatusCode} from "grpc-web";

const {
    VisualizeRequest, VisualizeEvent,
} = require('./proto/visualize_grpc_pb.js');
const {VisualizeGrpcServiceClient} = require('./proto/visualize_grpc_grpc_web_pb.js');


let type = "WebGL";
if (!PIXI.utils.isWebGLSupported()) {
    type = "canvas"
}

PIXI.utils.sayHello(type);
//Create a Pixi Application
let resolution = window.devicePixelRatio;

let [w, h] = getDesiredFieldSize();
let app = new PIXI.Application({
    width: w,
    height: h,
    // backgroundColor: 0xdddddd,
    transparent: true,
    autoDensity: true,
    antialias: true,
    resolution: resolution,
    sharedTicker: true,
});

document.body.appendChild(app.view);

// ensure that double-click outside the nodeWindow does not select text in there.
document.getElementById('nodeWindow').addEventListener("dblclick", function () {
    return false;
});

let vis = null;
let grpcServiceClient = null;
let ticker = PIXI.Ticker.shared;

function getDesiredFieldSize() {
    return [window.innerWidth - 20, window.innerHeight - 20]
}

window.addEventListener("resize", function () {
    let [w, h] = getDesiredFieldSize();
    app.renderer.resize(w, h);
    if (vis !== null) {
        vis.onResize(w, h)
    }
});

function loadOk() {
    console.log('connecting to gRPC server ' + server);
    grpcServiceClient = new VisualizeGrpcServiceClient(server);

    vis = new PixiVisualizer(app, grpcServiceClient);

    let [w, h] = getDesiredFieldSize();
    vis.onResize(w, h);
    ticker.add(function (dt) {
        vis.update(ticker.deltaMS / 1000)
    });

    let visualizeRequest = new VisualizeRequest();
    let metadata = {'tab': 'visualize'};
    let stream = grpcServiceClient.visualize(visualizeRequest, metadata);
    stream.on('data', function (resp) {
        let e = null;
        switch (resp.getTypeCase()) {
            case VisualizeEvent.TypeCase.SEND:
                e = resp.getSend();
                vis.visSend(e.getSrcId(), e.getDstId(), e.getMvInfo());
                break;
            case VisualizeEvent.TypeCase.ADD_NODE:
                e = resp.getAddNode();
                vis.visAddNode(e.getNodeId(), e.getX(), e.getY(), e.getRadioRange(), e.getNodeType());
                break;
            case VisualizeEvent.TypeCase.DELETE_NODE:
                e = resp.getDeleteNode();
                vis.visDeleteNode(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_POS:
                e = resp.getSetNodePos();
                vis.visSetNodePos(e.getNodeId(), e.getX(), e.getY());
                break;
            case VisualizeEvent.TypeCase.ON_NODE_FAIL:
                e = resp.getOnNodeFail();
                vis.visOnNodeFail(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.ON_NODE_RECOVER:
                e = resp.getOnNodeRecover();
                vis.visOnNodeRecover(e.getNodeId());
                break;
            case VisualizeEvent.TypeCase.ADD_ROUTER_TABLE:
                e = resp.getAddRouterTable();
                vis.visAddRouterTable(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.REMOVE_ROUTER_TABLE:
                e = resp.getRemoveRouterTable();
                vis.visRemoveRouterTable(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.ADD_CHILD_TABLE:
                e = resp.getAddChildTable();
                vis.visAddChildTable(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.REMOVE_CHILD_TABLE:
                e = resp.getRemoveChildTable();
                vis.visRemoveChildTable(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.SET_PARENT:
                e = resp.getSetParent();
                // TODO - currently OT does not emit this event. Workaround is used to call visSetParent().
                vis.visSetParent(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_PARTITION_ID:
                e = resp.getSetNodePartitionId();
                vis.visSetNodePartitionId(e.getNodeId(), e.getPartitionId());
                break;
            case VisualizeEvent.TypeCase.ADVANCE_TIME:
                e = resp.getAdvanceTime();
                vis.visAdvanceTime(e.getTs(), e.getSpeed());
                break;
            case VisualizeEvent.TypeCase.HEARTBEAT:
                e = resp.getHeartbeat();
                vis.visHeartbeat();
                break;
            case VisualizeEvent.TypeCase.SET_NODE_RLOC16:
                e = resp.getSetNodeRloc16();
                vis.visSetNodeRloc16(e.getNodeId(), e.getRloc16());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_ROLE:
                e = resp.getSetNodeRole();
                vis.visSetNodeRole(e.getNodeId(), e.getRole());
                break;
            case VisualizeEvent.TypeCase.COUNT_DOWN:
                e = resp.getCountDown();
                vis.visCountDown(e.getDurationMs(), e.getText());
                break;
            case VisualizeEvent.TypeCase.SHOW_DEMO_LEGEND:
                e = resp.getShowDemoLegend();
                vis.visShowDemoLegend(e.getX(), e.getY(), e.getTitle());
                break;
            case VisualizeEvent.TypeCase.SET_SPEED:
                e = resp.getSetSpeed();
                vis.visSetSpeed(e.getSpeed());
                break;
            case VisualizeEvent.TypeCase.ON_EXT_ADDR_CHANGE:
                e = resp.getOnExtAddrChange();
                vis.visOnExtAddrChange(e.getNodeId(), e.getExtAddr());
                break;
            case VisualizeEvent.TypeCase.SET_TITLE:
                e = resp.getSetTitle();
                vis.visSetTitle(e.getTitle(), e.getX(), e.getY(), e.getFontSize());
                break;
            case VisualizeEvent.TypeCase.SET_NODE_MODE:
                e = resp.getSetNodeMode();
                vis.visSetNodeMode(e.getNodeId(), e.getNodeMode());
                break;
            case VisualizeEvent.TypeCase.SET_NETWORK_INFO:
                e = resp.getSetNetworkInfo();
                vis.visSetNetworkInfo(e.getVersion(), e.getCommit(), e.getReal(), e.getNodeId(), e.getThreadVersion());
                break;
            default:
                console.error('unknown event!!! ' + resp.getTypeCase());
                break
        }

    });

    stream.on('status', function (status) {
        if (status != null) {
            if (status.code != StatusCode.OK) {
                console.error('visualize gRPC stream status: code = ' + status.code + ' details = ' + status.details);
                vis.stopIdleCheckTimer(); // stop expecting the HeartBeat events
            }else{
                console.log('visualize gRPC stream status: code = ' + status.code + ' details = ' + status.details);
            }
        }
    });

    stream.on('end', function (end) {
        // stream end signal
        console.log('visualize gRPC stream end');
        vis.stopIdleCheckTimer();
    });
}

app.loader
    .add('WhiteSolidCircle64', '/static/image/white-shapes/circle-64.png')
    .add('WhiteSolidTriangle64', '/static/image/white-shapes/triangle-64.png')
    .add('WhiteSolidHexagon64', '/static/image/white-shapes/hexagon-64.png')
    .add('WhiteSolidSquare64', '/static/image/white-shapes/square-64.png')
    .add('WhiteDashed4Circle64', '/static/image/white-shapes/circle-dashed-4-64.png')
    .add('WhiteDashed8Circle64', '/static/image/white-shapes/circle-dashed-8-64.png')
    .add('WhiteDashed8Circle128', '/static/image/white-shapes/circle-dashed-8-128.png')
    .add('FailedNodeMark', '/static/image/gua.png')
    .add('CheckedCheckbox32', '/static/image/checked-checkbox-32.png')
    .add('UncheckedCheckbox32', '/static/image/unchecked-checkbox-32.png')
    .add('WhiteRoundedDashedSquare128', '/static/image/white-shapes/square-dashed-rounded-128.png')
    .add('Play32', '/static/image/play-32.png')
    .add('Pause32', '/static/image/pause-32.png')
    .load((loader, res) => {
        SetResources(res);
        loadOk()
    });

