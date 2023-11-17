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

import Chart from 'chart.js/auto';
import {NodeStats} from "./StatsVisualizer";

export default class NodeNumbersChart {
    constructor(ctx) {
        this.fields = NodeStats.getFields();
        this.lastStats = new NodeStats();
        this.lastTimestampUs = 0;

        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: this.createDataSetsConfig(),
            },
            options: {
                animation: false,
                elements: {
                    point: {
                        radius: 1,
                        hoverRadius: 2,
                    },
                },
                parsing: false,
                plugins: {
                    title: {
                        display: true,
                        text: 'Node statistics viewer - click legend to toggle graphs; hover data points for details.'
                    },
                },
                responsive: true,
                maintainAspectRatio: false,
                scales: {
                    x: {
                        type: 'linear',
                        title: {
                            display: true,
                            text: 'Time (seconds)'
                        },
                        beginAtZero: true
                    },
                    y: {
                        title: {
                          display: true,
                          text: 'Number of nodes',
                        },
                        ticks: {
                            stepSize: 1,
                        },
                        beginAtZero: true
                    }
                },
            }
        });
    }

    createDataSetsConfig() {
        let aCfg = [];
        let aColors = [ "#0bb4ff", "#e60049", "#e6d800",
                                "#9b19f5", "#50e991", "#b3d4ff",
                                "#dc0ab4", "#00bfa0", "#ffa300"];
        for (let n in this.fields) {
            let colDark = aColors[n];
            let colLight = colDark + '99'; // add alpha channel to make it look lighter.
            let cfgItem = {
                label: this.fields[n],
                data: [],
                backgroundColor: [colLight],
                borderColor: [colDark],
                borderWidth: 1,
                tension: 0.0,
            };
            aCfg.push(cfgItem);
        }
        return aCfg;
    }

    update(timestampUs) {
        if (timestampUs > this.lastTimestampUs) {
            // move 'final graph point' to current timestamp.
            this._datasetsPop(this.chart.data);
            this._datasetsPush(timestampUs, this.lastStats, this.chart.data);
            this.chart.update();
        }
        this.lastTimestampUs = timestampUs;
    }

    addData(timestampUs, stats) {
        this.lastStats = stats;
        this._datasetsPop(this.chart.data); // remove 'final point'
        this._datasetsPushAndAdjust(timestampUs, stats, this.chart.data);
        this._datasetsPush(timestampUs + 0.1, stats, this.chart.data); // restore 'final point'
        //this.chart.data.labels.push(timestamp); // to check what effect this has. Currently, not missed.
        if (timestampUs > this.lastTimestampUs) {
            this.lastTimestampUs = timestampUs;
        }
    }

    _datasetsPush(ts, stats, data) {
        for (let i in this.fields) {
            let field = this.fields[i];
            let y = stats[field];
            data.datasets[i].data.push({x: ts/1e6, y: y}); // convert x: us to sec
        }
    }

    _datasetsPushAndAdjust(ts, stats, data) {
        for (let i in this.fields) {
            let dlen = data.datasets[i].data.length;
            let fieldName = this.fields[i];
            let y = stats[fieldName];
            if (dlen >= 2) {
                let y_old = data.datasets[i].data[dlen-1].y; // get last element
                let y_old2 = data.datasets[i].data[dlen-2].y; // get 2nd last element
                if (y == y_old && y == y_old2) { // to avoid too many points at same y value, remove superfluous y points
                    data.datasets[i].data.pop();
                }
            }
            data.datasets[i].data.push({x: ts / 1e6, y: y}); // convert x: us to sec
        }
    }

    _datasetsPop(data) {
        for (let i in this.fields) {
            data.datasets[i].data.pop();
        }
    }
}