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

import {NodeMode} from "../proto/visualize_grpc_pb";

const {
    OtDeviceRole,
} = require('../proto/visualize_grpc_pb.js');
import * as fmt from "../vis/format_text"

export class NodeStats {

    constructor() {
        this.numNodes      =0;
        this.numLeaders    =0;
        this.numPartitions =0;
        this.numRouters    =0;
        this.numEndDevices =0;
        this.numDetached   =0;
        this.numDisabled   =0;
        this.numSleepy     =0;
        this.numFailed     =0;
    }

    static getFields() {
        return ['numNodes', 'numLeaders', 'numPartitions', 'numRouters', 'numEndDevices', 'numDetached', 'numDisabled', 'numSleepy', 'numFailed'];
    }

    equals(other) {
        return this.numNodes === other.numNodes && this.numLeaders === other.numLeaders && this.numPartitions === other.numPartitions &&
            this.numRouters === other.numRouters && this.numEndDevices === other.numEndDevices && this.numDetached === other.numDetached &&
            this.numDisabled === other.numDisabled && this.numSleepy === other.numSleepy && this.numFailed === other.numFailed
    }

    printStats() {
        return `${this.numNodes}\t${this.numLeaders}\t${this.numPartitions}\t${this.numRouters}\t${this.numEndDevices}\t${this.numDetached}\t${this.numDisabled}\t${this.numSleepy}\t${this.numFailed}`
    }
}

export class StatsVisualizer {
    constructor() {
        this.nodeRoles = {};
        this.nodeModes = {};
        this.nodePartitions = {};
        this.nodesFailed = {};
        this.stats = new NodeStats();
        this.oldStats = new NodeStats();
        this.lastPlotTimestampUs = 0;
        this.arrayTimestamps = [];
        this.arrayStats = [];
    }

    visAdvanceTime(tsUs, speed) {
        this.arrayStats = [];
        this.arrayTimestamps = [];
        if (this.checkDataPointsChange()) {
            if (tsUs > this.lastPlotTimestampUs+1e3) {
                this.addDataPoint(tsUs-1e3, this.oldStats); // extra data point to plot staircase type graphs
            }
            this.addDataPoint(tsUs, this.stats);
            this.writeLogEntry(tsUs, this.stats);
            this.lastPlotTimestampUs = tsUs;
            this.oldStats = this.stats;
        }
    }

    visHeartbeat() {
    }

    visAddNode(nodeId, x, y, radioRange) {
        this.nodeRoles[nodeId] = OtDeviceRole.OT_DEVICE_ROLE_DISABLED;
        let msg = `Added at (${x},${y})`;
        this.logNode(nodeId, msg);
    }

    visSetNodeRole(nodeId, role) {
        let oldRole = this.nodeRoles[nodeId];
        this.nodeRoles[nodeId] = role;
        if (oldRole !== role) {
            this.logNode(nodeId, `Role changed from ${fmt.roleToString(oldRole)} to ${fmt.roleToString(role)}`)
        }
    }

    visSetNodeMode(nodeId, mode) {
        let oldMode = this.nodeModes[nodeId];
        this.nodeModes[nodeId] = mode;
        let oldModeStr = fmt.modeToString(oldMode);
        let modeStr = fmt.modeToString(mode);
        if (oldModeStr !== modeStr) {
            this.logNode(nodeId, `Mode changed from "${oldModeStr}" to "${modeStr}"`);
        }
    }

    visSetNodePartitionId(nodeId, partitionId) {
        let oldPartitionId = 0;
        if (nodeId in this.nodePartitions) {
            oldPartitionId = this.nodePartitions[nodeId];
        }
        this.nodePartitions[nodeId] = partitionId;
        if (oldPartitionId !== partitionId) {
            this.logNode(nodeId, `Partition changed from ${fmt.formatPartitionId(oldPartitionId)} to ${fmt.formatPartitionId(partitionId)}`)
        }
    }

    visDeleteNode(nodeId) {
        delete this.nodeModes[nodeId];
        delete this.nodeRoles[nodeId];
        delete this.nodesFailed[nodeId];
        this.logNode(nodeId, "Deleted")
    }

    visOnNodeFail(nodeId) {
        this.nodesFailed[nodeId] = true;
        this.logNode(nodeId, "Radio is OFF")
    }

    visOnNodeRecover(nodeId) {
        delete this.nodesFailed[nodeId]
        this.logNode(nodeId, "Radio is ON")
    }

    getNodeCountByRole(role) {
        let count = 0;
        for (let nodeid in this.nodeRoles) {
            let nr = this.nodeRoles[nodeid];
            if (nr === role) {
                count += 1
            }
        }
        return count
    }

    getPartitionCount() {
        let aPts = {};
        for (let nodeid in this.nodePartitions) {
            let pts = this.nodePartitions[nodeid];
            aPts[pts] = 1;
        }
        return Object.keys(aPts).length;
    }

    // counts the number of MTD SEDs in the network
    getSleepyCount() {
        let cnt = 0;
        for (let nodeid in this.nodeModes){
            let mode = this.nodeModes[nodeid];
            if (!mode.getRxOnWhenIdle() && !mode.getFullThreadDevice()){
                cnt++;
            }
        }
        return cnt;
    }

    calcStats() {
        let s = new NodeStats();
        s.numNodes = Object.keys(this.nodeRoles).length;
        s.numLeaders = this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_LEADER);
        s.numPartitions = this.getPartitionCount();
        s.numRouters = this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_ROUTER);
        s.numEndDevices = this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_CHILD);
        s.numDetached = this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_DETACHED);
        s.numDisabled = this.getNodeCountByRole(OtDeviceRole.OT_DEVICE_ROLE_DISABLED);
        s.numSleepy = this.getSleepyCount();
        s.numFailed = Object.keys(this.nodesFailed).length;

        return s;
    }

    checkDataPointsChange() {
        this.stats = this.calcStats();
        return !this.stats.equals(this.oldStats);
    }

    addDataPoint(tsUs, stats) {
        this.arrayTimestamps.push(tsUs); // timestamp in us
        this.arrayStats.push(stats);
    }

    getNewDataPoints() {
        // these arrays get cleared upon next call to visAdvanceTime()
        return [this.arrayTimestamps, this.arrayStats]
    }

    writeLogEntry(ts, stats) {
        let entry = stats.printStats();
        console.log(`${ts}: ${entry}`);
    }

    onResize(width, height) {
        console.log("window resized to " + width + "," + height);
    }

    logNode(nodeId, msg) {
        console.log(`Node ${nodeId}: ${msg}`)
    }

}
