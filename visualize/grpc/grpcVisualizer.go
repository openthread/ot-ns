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

package visualize_grpc

import (
	"math"
	"net"
	"sync"
	"time"

	"github.com/openthread/ot-ns/energy"
	"github.com/openthread/ot-ns/logger"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	"github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/openthread/ot-ns/visualize/grpc/replay"
)

type grpcVisualizer struct {
	simctrl             visualize.SimulationController
	server              *grpcServer
	f                   *grpcField
	showDemoLegendEvent *pb.VisualizeEvent
	replay              *replay.Replay
	energyAnalyser      *energy.EnergyAnalyser

	sync.Mutex
}

// NewGrpcVisualizer creates a new Visualizer that uses Google RPC communication with a web page.
func NewGrpcVisualizer(address string, replayFn string, chanNewClientNotifier chan string) visualize.Visualizer {
	gsv := &grpcVisualizer{
		simctrl: nil,
		f:       newGrpcField(),
	}

	if replayFn != "" {
		gsv.replay = replay.NewReplay(replayFn)
	}

	gsv.server = newGrpcServer(gsv, address, chanNewClientNotifier)
	return gsv
}

func (gv *grpcVisualizer) SetNetworkInfo(networkInfo visualize.NetworkInfo) {
	gv.Lock()
	defer gv.Unlock()

	if networkInfo.NodeId == InvalidNodeId {
		gv.f.networkInfo = networkInfo
	} else {
		gv.f.setNodeVersion(networkInfo.NodeId, networkInfo.Version)
		gv.f.setNodeCommit(networkInfo.NodeId, networkInfo.Commit)
		gv.f.setNodeThreadVersion(networkInfo.NodeId, networkInfo.ThreadVersion)
	}
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNetworkInfo{SetNetworkInfo: &pb.SetNetworkInfoEvent{
		Real:          networkInfo.Real,
		Version:       networkInfo.Version,
		Commit:        networkInfo.Commit,
		NodeId:        int32(networkInfo.NodeId),
		ThreadVersion: int32(networkInfo.ThreadVersion),
	}}})
}

func (gv *grpcVisualizer) Init() {
	//
}

func (gv *grpcVisualizer) Run() {
	defer logger.Debugf("gRPC server exit.")
	err := gv.server.Run()

	gv.Lock()
	defer gv.Unlock()
	gv.server.stop()

	if err != nil {
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Op == "listen" {
				logger.Errorf("gRPC server could not listen on %s - port may be in use?", gv.server.address)
				return
			}
		}
		logger.Errorf("gRPC server quit with error: %v", err)
	}
}

func (gv *grpcVisualizer) Stop() {
	gv.Lock()
	defer gv.Unlock()

	gv.server.stop()
	if gv.replay != nil {
		gv.replay.Close()
	}
}

func (gv *grpcVisualizer) AddNode(nodeid NodeId, cfg *NodeConfig) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addNode(nodeid, cfg)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddNode{AddNode: &pb.AddNodeEvent{
		NodeId:     int32(nodeid),
		X:          int32(cfg.X),
		Y:          int32(cfg.Y),
		Z:          int32(cfg.Z),
		RadioRange: int32(cfg.RadioRange),
		NodeType:   cfg.Type,
	}}})
}

func (gv *grpcVisualizer) OnExtAddrChange(nodeid NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	logger.Debugf("extaddr changed: node=%d, extaddr=%016x, old extaddr=%016x", nodeid, extaddr, gv.f.nodes[nodeid].extaddr)
	gv.f.onExtAddrChange(nodeid, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnExtAddrChange{OnExtAddrChange: &pb.OnExtAddrChangeEvent{
		NodeId:  int32(nodeid),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) SetNodeRloc16(nodeid NodeId, rloc16 uint16) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeRloc16(nodeid, rloc16)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeRloc16{SetNodeRloc16: &pb.SetNodeRloc16Event{
		NodeId: int32(nodeid),
		Rloc16: uint32(rloc16),
	}}})
}

func (gv *grpcVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeRole(nodeid, role)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeRole{SetNodeRole: &pb.SetNodeRoleEvent{
		NodeId: int32(nodeid),
		Role:   pb.OtDeviceRole(role),
	}}})
}

func (gv *grpcVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeMode(nodeid, mode)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeMode{SetNodeMode: &pb.SetNodeModeEvent{
		NodeId: int32(nodeid),
		NodeMode: &pb.NodeMode{
			RxOnWhenIdle:     mode.RxOnWhenIdle,
			FullThreadDevice: mode.FullThreadDevice,
			FullNetworkData:  mode.FullNetworkData,
		},
	}}})
}

func (gv *grpcVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *visualize.MsgVisualizeInfo) {
	gv.Lock()
	defer gv.Unlock()

	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_Send{Send: &pb.SendEvent{
		SrcId: int32(srcid),
		DstId: int32(dstid),
		MvInfo: &pb.MsgVisualizeInfo{
			Channel:         uint32(mvinfo.Channel),
			FrameControl:    uint32(mvinfo.FrameControl),
			Seq:             uint32(mvinfo.Seq),
			DstAddrShort:    uint32(mvinfo.DstAddrShort),
			DstAddrExtended: mvinfo.DstAddrExtended,
			SendDurationUs:  mvinfo.SendDurationUs,
			VisTrueDuration: gv.f.speed <= 0.01,
			PowerDbm:        int32(mvinfo.PowerDbm),
			FrameSizeBytes:  uint32(mvinfo.FrameSizeBytes),
		},
	}}})
}

func (gv *grpcVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodePartitionId(nodeid, parid)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodePartitionId{SetNodePartitionId: &pb.SetNodePartitionIdEvent{
		NodeId:      int32(nodeid),
		PartitionId: parid,
	}}})
}

func (gv *grpcVisualizer) SetSpeed(speed float64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setSpeed(speed)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetSpeed{SetSpeed: &pb.SetSpeedEvent{
		Speed: speed,
	}}})
}

func (gv *grpcVisualizer) AdvanceTime(ts uint64, speed float64) {
	gv.Lock()
	defer gv.Unlock()

	if gv.f.advanceTime(ts, speed) {
		gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AdvanceTime{AdvanceTime: &pb.AdvanceTimeEvent{
			Timestamp: ts,
			Speed:     speed,
		}}})
	}
}

func (gv *grpcVisualizer) OnNodeFail(nodeid NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.onNodeFail(nodeid)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnNodeFail{OnNodeFail: &pb.OnNodeFailEvent{
		NodeId: int32(nodeid),
	}}})
}

func (gv *grpcVisualizer) OnNodeRecover(nodeid NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.onNodeRecover(nodeid)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnNodeRecover{OnNodeRecover: &pb.OnNodeRecoverEvent{
		NodeId: int32(nodeid),
	}}})
}

func (gv *grpcVisualizer) SetController(ctrl visualize.SimulationController) {
	gv.Lock()
	defer gv.Unlock()

	gv.simctrl = ctrl
}

func (gv *grpcVisualizer) SetNodePos(nodeid NodeId, x, y, z int) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodePos(nodeid, x, y, z)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodePos{SetNodePos: &pb.SetNodePosEvent{
		NodeId: int32(nodeid),
		X:      int32(x),
		Y:      int32(y),
		Z:      int32(z),
	}}})
}

func (gv *grpcVisualizer) DeleteNode(id NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.deleteNode(id)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_DeleteNode{DeleteNode: &pb.DeleteNodeEvent{
		NodeId: int32(id),
	}}})
}

func (gv *grpcVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addRouterTable(id, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddRouterTable{AddRouterTable: &pb.AddRouterTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.removeRouterTable(id, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_RemoveRouterTable{RemoveRouterTable: &pb.RemoveRouterTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) AddChildTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addChildTable(id, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddChildTable{AddChildTable: &pb.AddChildTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.removeChildTable(id, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_RemoveChildTable{RemoveChildTable: &pb.RemoveChildTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) ShowDemoLegend(x int, y int, title string) {
	gv.Lock()
	defer gv.Unlock()

	e := &pb.VisualizeEvent{Type: &pb.VisualizeEvent_ShowDemoLegend{ShowDemoLegend: &pb.ShowDemoLegendEvent{
		X:     int32(x),
		Y:     int32(y),
		Title: title,
	}}}
	gv.showDemoLegendEvent = e
	gv.addVisualizeEvent(e)
}

func (gv *grpcVisualizer) CountDown(duration time.Duration, text string) {
	gv.Lock()
	defer gv.Unlock()

	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_CountDown{CountDown: &pb.CountDownEvent{
		DurationMs: int64(duration / time.Millisecond),
		Text:       text,
	}}})
}

func (gv *grpcVisualizer) SetParent(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setParent(id, extaddr)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetParent{SetParent: &pb.SetParentEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}})
}

func (gv *grpcVisualizer) SetTitle(titleInfo visualize.TitleInfo) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setTitleInfo(titleInfo)
	gv.addVisualizeEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetTitle{SetTitle: &pb.SetTitleEvent{
		Title:    titleInfo.Title,
		X:        int32(titleInfo.X),
		Y:        int32(titleInfo.Y),
		FontSize: int32(titleInfo.FontSize),
	}}})
}

func (gv *grpcVisualizer) UpdateNodesEnergy(nodes []*energy.NodeEnergy, timestamp uint64, updateView bool) {
	gv.Lock()
	defer gv.Unlock()

	// convert to protobuf data structure
	nodesPb := make([]*pb.NodeEnergy, len(nodes))
	for n, nodeEnergy := range nodes {
		nodesPb[n] = &pb.NodeEnergy{
			NodeId:   int32(nodeEnergy.NodeId),
			Disabled: nodeEnergy.Disabled,
			Sleep:    nodeEnergy.Sleep,
			Tx:       nodeEnergy.Tx,
			Rx:       nodeEnergy.Rx,
		}
	}

	//logger.Debugf("Updating Nodes Energy to the charts")
	gv.server.SendEnergyEvent(&pb.EnergyEvent{
		Timestamp:   timestamp / 1000000, // convert to s
		NodesEnergy: nodesPb,
	})
	if updateView {
		gv.server.SendEnergyEvent(&pb.EnergyEvent{
			Timestamp:   math.MaxUint64, // convert to s
			NodesEnergy: make([]*pb.NodeEnergy, 0),
		},
		)
	}
}

func (gv *grpcVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {
	gv.Lock()
	defer gv.Unlock()

	gv.energyAnalyser = ea
}

func (gv *grpcVisualizer) UpdateNodeStats(nsi *visualize.NodeStatsInfo) {
	gv.Lock()
	defer gv.Unlock()

	nodeStatsPb := &pb.NodeStats{
		NumNodes:      uint32(nsi.Stats.NumNodes),
		NumLeaders:    uint32(nsi.Stats.NumLeaders),
		NumPartitions: uint32(nsi.Stats.NumPartitions),
		NumRouters:    uint32(nsi.Stats.NumRouters),
		NumEndDevices: uint32(nsi.Stats.NumEndDevices),
		NumDetached:   uint32(nsi.Stats.NumDetached),
		NumDisabled:   uint32(nsi.Stats.NumDisabled),
		NumSleepy:     uint32(nsi.Stats.NumSleepy),
		NumFailed:     uint32(nsi.Stats.NumFailed),
	}
	gv.f.setNodeStatsInfo(*nsi)
	e := &pb.VisualizeEvent{Type: &pb.VisualizeEvent_NodeStatsInfo{NodeStatsInfo: &pb.NodeStatsInfoEvent{
		Timestamp: nsi.TimeUs,
		NodeStats: nodeStatsPb,
	}}}
	gv.addVisualizeEvent(e)
}

func (gv *grpcVisualizer) UpdateTimeWindowStats(timeWinStatsInfo *visualize.TimeWindowStatsInfo) {
	// not used for now
}

func (gv *grpcVisualizer) prepareStream(stream *grpcStream) error {
	// set global network info (not-node-specific)
	if err := stream.Send(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNetworkInfo{SetNetworkInfo: &pb.SetNetworkInfoEvent{
		Real:    gv.f.networkInfo.Real,
		Version: gv.f.networkInfo.Version,
		Commit:  gv.f.networkInfo.Commit,
		NodeId:  int32(InvalidNodeId),
	}}}); err != nil {
		return err
	}

	// show demo legend if necessary
	if gv.showDemoLegendEvent != nil {
		if err := stream.Send(gv.showDemoLegendEvent); err != nil {
			return err
		}
	}

	// set speed
	if err := stream.Send(&pb.VisualizeEvent{
		Type: &pb.VisualizeEvent_SetSpeed{SetSpeed: &pb.SetSpeedEvent{
			Speed: gv.f.speed,
		}},
	}); err != nil {
		return err
	}

	// set title
	if gv.f.titleInfo.Title != "" {
		if err := stream.Send(&pb.VisualizeEvent{
			Type: &pb.VisualizeEvent_SetTitle{SetTitle: &pb.SetTitleEvent{
				Title:    gv.f.titleInfo.Title,
				X:        int32(gv.f.titleInfo.X),
				Y:        int32(gv.f.titleInfo.Y),
				FontSize: int32(gv.f.titleInfo.FontSize),
			}},
		}); err != nil {
			return err
		}
	}

	// advance time
	if err := stream.Send(&pb.VisualizeEvent{
		Type: &pb.VisualizeEvent_AdvanceTime{AdvanceTime: &pb.AdvanceTimeEvent{
			Timestamp: gv.f.curTime,
			Speed:     gv.f.curSpeed,
		}},
	}); err != nil {
		return err
	}

	if stream.vizType == meshTopologyVizType {
		// draw all nodes
		for nodeid, node := range gv.f.nodes {
			addNodeEvent := &pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddNode{AddNode: &pb.AddNodeEvent{
				NodeId:     int32(nodeid),
				X:          int32(node.x),
				Y:          int32(node.y),
				RadioRange: int32(node.radioRange),
				NodeType:   node.nodeType,
			}}}

			if err := stream.Send(addNodeEvent); err != nil {
				return err
			}
		}

		// draw node attributes
		for nodeid, node := range gv.f.nodes {
			// extaddr
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_OnExtAddrChange{OnExtAddrChange: &pb.OnExtAddrChangeEvent{
					NodeId:  int32(nodeid),
					ExtAddr: node.extaddr,
				}},
			}); err != nil {
				return err
			}
			// rloc16
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetNodeRloc16{SetNodeRloc16: &pb.SetNodeRloc16Event{
					NodeId: int32(nodeid),
					Rloc16: uint32(node.rloc16),
				}},
			}); err != nil {
				return err
			}
			// role
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetNodeRole{SetNodeRole: &pb.SetNodeRoleEvent{
					NodeId: int32(nodeid),
					Role:   pb.OtDeviceRole(node.role),
				}},
			}); err != nil {
				return err
			}
			// mode
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetNodeMode{SetNodeMode: &pb.SetNodeModeEvent{
					NodeId: int32(nodeid),
					NodeMode: &pb.NodeMode{
						RxOnWhenIdle:     node.mode.RxOnWhenIdle,
						FullThreadDevice: node.mode.FullThreadDevice,
						FullNetworkData:  node.mode.FullNetworkData,
					},
				}},
			}); err != nil {
				return err
			}
			// partition id
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetNodePartitionId{SetNodePartitionId: &pb.SetNodePartitionIdEvent{
					NodeId:      int32(nodeid),
					PartitionId: node.partitionId,
				}},
			}); err != nil {
				return err
			}
			// parent
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetParent{SetParent: &pb.SetParentEvent{
					NodeId:  int32(nodeid),
					ExtAddr: node.parent,
				}},
			}); err != nil {
				return err
			}

			// child table
			for extaddr := range node.childTable {
				if err := stream.Send(&pb.VisualizeEvent{
					Type: &pb.VisualizeEvent_AddChildTable{AddChildTable: &pb.AddChildTableEvent{
						NodeId:  int32(nodeid),
						ExtAddr: extaddr,
					}},
				}); err != nil {
					return err
				}
			}
			// router table
			for extaddr := range node.routerTable {
				if err := stream.Send(&pb.VisualizeEvent{
					Type: &pb.VisualizeEvent_AddRouterTable{AddRouterTable: &pb.AddRouterTableEvent{
						NodeId:  int32(nodeid),
						ExtAddr: extaddr,
					}},
				}); err != nil {
					return err
				}
			}
			// node fail
			if node.failed {
				if err := stream.Send(&pb.VisualizeEvent{
					Type: &pb.VisualizeEvent_OnNodeFail{OnNodeFail: &pb.OnNodeFailEvent{
						NodeId: int32(nodeid),
					}},
				}); err != nil {
					return err
				}
			}
			// node type and thread version
			if err := stream.Send(&pb.VisualizeEvent{
				Type: &pb.VisualizeEvent_SetNetworkInfo{SetNetworkInfo: &pb.SetNetworkInfoEvent{
					Real:          false,
					Version:       node.version,
					Commit:        node.commit,
					NodeId:        int32(nodeid),
					ThreadVersion: int32(node.threadVersion),
				}},
			}); err != nil {
				return err
			}
		}
	}

	if stream.vizType == nodeStatsVizType {
		ns := gv.f.nodeStatsInfo.Stats
		pbNodeStats := &pb.NodeStats{
			NumNodes:      uint32(ns.NumNodes),
			NumLeaders:    uint32(ns.NumLeaders),
			NumPartitions: uint32(ns.NumPartitions),
			NumRouters:    uint32(ns.NumRouters),
			NumEndDevices: uint32(ns.NumEndDevices),
			NumDetached:   uint32(ns.NumDetached),
			NumDisabled:   uint32(ns.NumDisabled),
			NumSleepy:     uint32(ns.NumSleepy),
			NumFailed:     uint32(ns.NumFailed),
		}
		pbNodeStatsInfo := &pb.NodeStatsInfoEvent{
			Timestamp: gv.f.curTime,
			NodeStats: pbNodeStats,
		}
		if err := stream.Send(&pb.VisualizeEvent{
			Type: &pb.VisualizeEvent_NodeStatsInfo{NodeStatsInfo: pbNodeStatsInfo},
		}); err != nil {
			return err
		}
	}

	return nil
}

func (gv *grpcVisualizer) addVisualizeEvent(event *pb.VisualizeEvent) {
	if gv.replay != nil {
		gv.replay.Append(event)
	}
	gv.server.SendEvent(event)
}
