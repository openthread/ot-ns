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

import Chart from 'chart.js/auto';

export default class PowerConsumptionChart {
    constructor(ctx, deciConfig) {
        this.lastValue = 0;
        this.lastTimestamp = 0;
        this.accumulated = 0;
        this.hasFirstValue = false;

        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Value',
                    data: [],
                    backgroundColor: [
                        'rgba(100, 99, 132, 0.2)',
                    ],
                    borderColor: [
                        'rgba(100, 99, 132, 1)',
                    ],
                    borderWidth: 1,
                    tension: 0.3
                }]
            },
            options: {
                // animation: false,
                parsing: false,
                plugins: {
                    decimation: deciConfig,
                    title: {
                        display: true,
                        text: 'Node # power consumption'
                    },
                },
                responsive: true,
                scales: {
                    x: {
                        type: 'linear',
                        title: {
                            display: true,
                            text: 'Time (seconds)'
                        }
                    },
                    y: {
                        title: {
                          display: true,
                          text: 'Power (mW)',
                        },
                        beginAtZero: true
                    }
                }
            }
        });
    }

    update() {
        let title = "Node "+ document.getElementById('nodesForPower').value +" power consumption.";
        title += " Mean = " + (this.accumulated/this.chart.data.labels.length).toFixed(3) + " mW";
        this.chart.options.plugins.title.text = title;
        this.chart.update();
    }

    reset() {
        this.lastValue = 0;
        this.lastTimestamp = 0;
        this.chart.data.labels = [];
        this.chart.data.datasets[0].data = [];
        this.accumulated = 0;
        this.hasFirstValue = false;
        this.chart.update();
    }

    addNodesPoint(nodes,timestamp) {
        let selectedNodeId = document.getElementById('nodesForPower').value;
        if (selectedNodeId == "") {
            //Get any node from the Map nodes
            for (let nodeId in nodes.values()) {
                selectedNodeId = nodeId;
                break;
            }
        }
        
        let total = 0;
        if (nodes.has(parseInt(selectedNodeId))) {
            total = nodes.get(parseInt(selectedNodeId)).getTotal();
        } else {
            //This node did not exist in this time interval
            return this;
        }
        
        if (this.hasFirstValue) {
            let curPowerCons = (total - this.lastValue)/(timestamp - this.lastTimestamp);
            this.accumulated += curPowerCons;
            this.chart.data.datasets[0].data.push({
                x: timestamp,
                y: curPowerCons
            });
            this.chart.data.labels.push(timestamp);
        } else {
            this.hasFirstValue = true;
        }
        this.lastValue = total;
        this.lastTimestamp = timestamp;
        return this;
    }

    setData(nodesOverTime, timestamps) {
        for (let idx = 0; idx < nodesOverTime.length; idx++) {    
            this.addNodesPoint(nodesOverTime[idx], timestamps[idx]);
        }
        return this;
    }
}
