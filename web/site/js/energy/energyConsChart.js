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


export default class EnergyConsumptionChart {
    constructor(ctx, deciConfig) {
        this.chart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [],
                datasets: [{
                    label: 'Value',
                    data: [],
                    backgroundColor: [
                        'rgba(255, 99, 132, 0.2)',
                    ],
                    borderColor: [
                        'rgba(255, 99, 132, 1)',
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
                        text: 'Network average energy consumption'
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
                          text: 'Energy (mJ)',
                        },
                        beginAtZero: true
                    }
                }
            }
        });
    }

    update() {
        if (document.getElementById('nodesForEnergy').value === 'all') {
            this.chart.options.plugins.title.text = "Network average energy consumption";
        } else {
            this.chart.options.plugins.title.text = "Node "+ document.getElementById('nodesForEnergy').value +" energy consumption";
        }
        this.chart.update();
    }

    reset() {
        this.chart.data.labels = [];
        this.chart.data.datasets[0].data = [];
        this.chart.update();
    }

    addNodesPoint(nodes, timestamp) {
        let total = 0;
        if (document.getElementById('nodesForEnergy').value === 'all') {
            for (const node of nodes.values()) {
                total += node.getTotal();
            }
            total = total/nodes.size;
        } else {
            if (nodes.has(parseInt(document.getElementById('nodesForEnergy').value))) {
                total = nodes.get(parseInt(document.getElementById('nodesForEnergy').value)).getTotal();
            } else {
                //This node did not exist in this time interval
                return this;
            }
        }
        this.chart.data.datasets[0].data.push({
            x: timestamp,
            y: total
        });
        this.chart.data.labels.push(timestamp);
        return this;
    }

    setData(nodesOverTime, timestamps) {
        let idx = 0;
        for (let nodes of nodesOverTime) {
            this.addNodesPoint(nodes, timestamps[idx++]);
        }
        return this;
    }
}