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

    initFrom(grpcStats) {
        this.numNodes = grpcStats.getNumnodes()
        this.numLeaders = grpcStats.getNumleaders()
        this.numPartitions = grpcStats.getNumpartitions()
        this.numRouters = grpcStats.getNumrouters()
        this.numEndDevices = grpcStats.getNumenddevices()
        this.numDetached = grpcStats.getNumdetached()
        this.numDisabled = grpcStats.getNumdisabled()
        this.numFailed = grpcStats.getNumfailed()
    }

    toLogString() {
        return `${this.numNodes}\t${this.numLeaders}\t${this.numPartitions}\t${this.numRouters}\t${this.numEndDevices}\t${this.numDetached}\t${this.numDisabled}\t${this.numSleepy}\t${this.numFailed}`
    }
}

export class TimeWindowStats {

    constructor() {
        this.ts = 0;
        this.windowDuration = 0;
        this.numBytesSent = {}; // per-node counter
        this.numBytesSentTotal = 0;
    }

}

export class StatsVisualizer {
    constructor() {
        this.stats = new NodeStats();
        this.ts = 0;
        this.arrayTimestamps = [];
        this.arrayStats = [];
        this.timeWindowStats = new TimeWindowStats();
    }

    visNodeStatsInfo(tsUs, grpcStats) {
        this.arrayStats = [];
        this.arrayTimestamps = [];
        if (tsUs > this.ts+1e3) {
            this.addDataPoint(tsUs-1e3, this.stats); // extra data point to plot staircase type graphs
        }

        this.stats = new NodeStats()
        this.stats.initFrom(grpcStats)
        this.addDataPoint(tsUs, this.stats);
        this.writeLogEntry(tsUs, this.stats);
        this.ts = tsUs;
    }

    visAdvanceTime(tsUs) {
    }

    visHeartbeat() {
    }

    addDataPoint(tsUs, stats) {
        this.arrayTimestamps.push(tsUs); // timestamp in us
        this.arrayStats.push(stats);
    }

    getNewDataPoints() {
        // these arrays get cleared upon next call to visNodeStatsInfo()
        return [this.arrayTimestamps, this.arrayStats]
    }

    writeLogEntry(ts, stats) {
        let entry = stats.toLogString();
        console.log(`${ts}: ${entry}`);
    }

    onResize(width, height) {
        console.log("window resized to " + width + "," + height);
    }

}
