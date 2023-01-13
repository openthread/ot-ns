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

const CHART_COLORS = {
    red: 'rgb(255, 99, 132)',
    black: 'rgb(0, 0, 0)',
    orange: 'rgb(255, 159, 64)',
    yellow: 'rgb(255, 205, 86)',
    green: 'rgb(75, 192, 192)',
    blue: 'rgb(54, 162, 235)',
    purple: 'rgb(153, 102, 255)',
    grey: 'rgb(201, 203, 207)'
  };

export default class NodeEnergyConsumptionBarChart {
    constructor(ctx, deciConfig) {
        this.chart = new Chart(ctx, {
            type: 'bar',
            data: {
                labels: [],
                datasets: [
                    {
                        label: 'Disabled',
                        backgroundColor: CHART_COLORS.grey,
                    },
                    {
                        label: 'Sleep',
                        backgroundColor: CHART_COLORS.yellow,
                    },
                    {
                        label: 'Rx',
                        backgroundColor: CHART_COLORS.blue,
                    },
                    {
                        label: 'Tx',
                        backgroundColor: CHART_COLORS.red,
                    }
                  ]
            },
            options: {
                plugins: {
                    title: {
                        display: true,
                        text: 'Energy consumption per radio state'
                    },
                    decimation: deciConfig,
                },
                responsive: true,
                scales: {
                    x: {
                        title: {
                            display: true,
                            text: 'Node ID'
                        },
                        stacked: true,
                    },
                    y: {
                        title: {
                          display: true,
                          text: 'Energy (mJ)',
                        },
                        stacked: true,
                    }
                }
            }
        });
    }

    update() {
        this.chart.update();
    }

    updateNode(node) {
        /*
         * this.chart.data.labels, this.chart.data.datasets[0], this.chart.data.datasets[1], this.chart.data.datasets[2], this.chart.data.datasets[3] are arrays of the same length,
         * containing the node's ID, the node's energy consumption in disabled, sleep, rx, tx states, respectively.
        */

        //if the node is not in the chart, add it sorted by node ID, increasing order
        let nodeIdx = this.chart.data.labels.indexOf(node.id);
        if (nodeIdx === -1) {
            let index = 0;
            for (; index < this.chart.data.labels.length; index++) {
                if (parseInt(this.chart.data.labels[index]) > node.id) {
                    break;
                }
            }
            this.chart.data.labels.splice(index, 0, node.id);
            this.chart.data.datasets[0].data.splice(index, 0, node.disabled);
            this.chart.data.datasets[1].data.splice(index, 0, node.sleep);
            this.chart.data.datasets[2].data.splice(index, 0, node.rx);
            this.chart.data.datasets[3].data.splice(index, 0, node.tx);
        } else {
            this.chart.data.datasets[0].data[nodeIdx] = node.disabled;
            this.chart.data.datasets[1].data[nodeIdx] = node.sleep;
            this.chart.data.datasets[2].data[nodeIdx] = node.rx;
            this.chart.data.datasets[3].data[nodeIdx] = node.tx;
        }
    }


    setData(nodesOverTime) {
        //Remove nodes of the chart that are not in the Map nodes
        for (let i = 0; i < this.chart.data.labels.length; i++) {
            if (!nodesOverTime.has(parseInt(this.chart.data.labels[i]))) {
                this.chart.data.labels.splice(i, 1);
                this.chart.data.datasets[0].data.splice(i, 1);
                this.chart.data.datasets[1].data.splice(i, 1);
                this.chart.data.datasets[2].data.splice(i, 1);
                this.chart.data.datasets[3].data.splice(i, 1);
                i--;
            }
        }
        
        for (let node of nodesOverTime.values()) {
            this.updateNode(node);
        }
        return this;
    }
}
