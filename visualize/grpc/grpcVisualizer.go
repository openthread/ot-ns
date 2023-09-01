// Copyright (c) 2020-2023, The OTNS Authors.
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
	"sync"
	"time"

	"github.com/openthread/ot-ns/energy"
	. "github.com/openthread/ot-ns/types"
	"github.com/openthread/ot-ns/visualize"
	pb "github.com/openthread/ot-ns/visualize/grpc/pb"
	"github.com/openthread/ot-ns/visualize/grpc/replay"
	"github.com/simonlingoogle/go-simplelogger"
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

func (gv *grpcVisualizer) SetNetworkInfo(networkInfo visualize.NetworkInfo) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.networkInfo = networkInfo
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNetworkInfo{SetNetworkInfo: &pb.SetNetworkInfoEvent{
		Real:    networkInfo.Real,
		Version: networkInfo.Version,
		Commit:  networkInfo.Commit,
	}}}, false)
}

func (gv *grpcVisualizer) Run() {
	err := gv.server.Run()
	if err != nil {
		simplelogger.Warnf("gRPC server quit: %v", err)
	}
}

func (gv *grpcVisualizer) Stop() {
	gv.server.stop()
	if gv.replay != nil {
		gv.replay.Close()
	}
}

func (gv *grpcVisualizer) AddNode(nodeid NodeId, x int, y int, radioRange int) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addNode(nodeid, x, y, radioRange)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddNode{AddNode: &pb.AddNodeEvent{
		NodeId:     int32(nodeid),
		X:          int32(x),
		Y:          int32(y),
		RadioRange: int32(radioRange),
	}}}, false)
}

func (gv *grpcVisualizer) OnExtAddrChange(nodeid NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	simplelogger.Debugf("extaddr changed: node=%d, extaddr=%016x, old extaddr=%016x", nodeid, extaddr, gv.f.nodes[nodeid].extaddr)
	gv.f.onExtAddrChange(nodeid, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnExtAddrChange{OnExtAddrChange: &pb.OnExtAddrChangeEvent{
		NodeId:  int32(nodeid),
		ExtAddr: extaddr,
	}}}, false)
}

func (gv *grpcVisualizer) SetNodeRloc16(nodeid NodeId, rloc16 uint16) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeRloc16(nodeid, rloc16)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeRloc16{SetNodeRloc16: &pb.SetNodeRloc16Event{
		NodeId: int32(nodeid),
		Rloc16: uint32(rloc16),
	}}}, false)
}

func (gv *grpcVisualizer) SetNodeRole(nodeid NodeId, role OtDeviceRole) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeRole(nodeid, role)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeRole{SetNodeRole: &pb.SetNodeRoleEvent{
		NodeId: int32(nodeid),
		Role:   pb.OtDeviceRole(role),
	}}}, false)
}

func (gv *grpcVisualizer) SetNodeMode(nodeid NodeId, mode NodeMode) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodeMode(nodeid, mode)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodeMode{SetNodeMode: &pb.SetNodeModeEvent{
		NodeId: int32(nodeid),
		NodeMode: &pb.NodeMode{
			RxOnWhenIdle:     mode.RxOnWhenIdle,
			FullThreadDevice: mode.FullThreadDevice,
			FullNetworkData:  mode.FullNetworkData,
		},
	}}}, false)
}

func (gv *grpcVisualizer) Send(srcid NodeId, dstid NodeId, mvinfo *visualize.MsgVisualizeInfo) {
	gv.Lock()
	defer gv.Unlock()

	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_Send{Send: &pb.SendEvent{
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
		},
	}}}, false)
}

func (gv *grpcVisualizer) SetNodePartitionId(nodeid NodeId, parid uint32) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodePartitionId(nodeid, parid)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodePartitionId{SetNodePartitionId: &pb.SetNodePartitionIdEvent{
		NodeId:      int32(nodeid),
		PartitionId: parid,
	}}}, false)
}

func (gv *grpcVisualizer) SetSpeed(speed float64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setSpeed(speed)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetSpeed{SetSpeed: &pb.SetSpeedEvent{
		Speed: speed,
	}}}, false)
}

func (gv *grpcVisualizer) AdvanceTime(ts uint64, speed float64) {
	gv.Lock()
	defer gv.Unlock()

	if gv.f.advanceTime(ts, speed) {
		gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AdvanceTime{AdvanceTime: &pb.AdvanceTimeEvent{
			Ts:    ts,
			Speed: speed,
		}}}, true)
	}
}

func (gv *grpcVisualizer) OnNodeFail(nodeid NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.onNodeFail(nodeid)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnNodeFail{OnNodeFail: &pb.OnNodeFailEvent{
		NodeId: int32(nodeid),
	}}}, false)
}

func (gv *grpcVisualizer) OnNodeRecover(nodeid NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.onNodeRecover(nodeid)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_OnNodeRecover{OnNodeRecover: &pb.OnNodeRecoverEvent{
		NodeId: int32(nodeid),
	}}}, false)
}

func (gv *grpcVisualizer) SetController(ctrl visualize.SimulationController) {
	gv.simctrl = ctrl
}

func (gv *grpcVisualizer) SetNodePos(nodeid NodeId, x, y int) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setNodePos(nodeid, x, y)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNodePos{SetNodePos: &pb.SetNodePosEvent{
		NodeId: int32(nodeid),
		X:      int32(x),
		Y:      int32(y),
	}}}, false)
}

func (gv *grpcVisualizer) DeleteNode(id NodeId) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.deleteNode(id)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_DeleteNode{DeleteNode: &pb.DeleteNodeEvent{
		NodeId: int32(id),
	}}}, false)
}

func (gv *grpcVisualizer) AddRouterTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addRouterTable(id, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddRouterTable{AddRouterTable: &pb.AddRouterTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}}, false)
}

func (gv *grpcVisualizer) RemoveRouterTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.removeRouterTable(id, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_RemoveRouterTable{RemoveRouterTable: &pb.RemoveRouterTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}}, false)
}

func (gv *grpcVisualizer) AddChildTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.addChildTable(id, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddChildTable{AddChildTable: &pb.AddChildTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}}, false)
}

func (gv *grpcVisualizer) RemoveChildTable(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.removeChildTable(id, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_RemoveChildTable{RemoveChildTable: &pb.RemoveChildTableEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}}, false)
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
	gv.AddVisualizationEvent(e, false)
}

func (gv *grpcVisualizer) CountDown(duration time.Duration, text string) {
	gv.Lock()
	defer gv.Unlock()

	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_CountDown{CountDown: &pb.CountDownEvent{
		DurationMs: int64(duration / time.Millisecond),
		Text:       text,
	}}}, false)
}

func (gv *grpcVisualizer) SetParent(id NodeId, extaddr uint64) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setParent(id, extaddr)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetParent{SetParent: &pb.SetParentEvent{
		NodeId:  int32(id),
		ExtAddr: extaddr,
	}}}, false)
}

func (gv *grpcVisualizer) SetTitle(titleInfo visualize.TitleInfo) {
	gv.Lock()
	defer gv.Unlock()

	gv.f.setTitleInfo(titleInfo)
	gv.AddVisualizationEvent(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetTitle{SetTitle: &pb.SetTitleEvent{
		Title:    titleInfo.Title,
		X:        int32(titleInfo.X),
		Y:        int32(titleInfo.Y),
		FontSize: int32(titleInfo.FontSize),
	}}}, false)
}

func (gv *grpcVisualizer) prepareStream(stream *grpcStream) error {
	// set network info
	if err := stream.Send(&pb.VisualizeEvent{Type: &pb.VisualizeEvent_SetNetworkInfo{SetNetworkInfo: &pb.SetNetworkInfoEvent{
		Real:    gv.f.networkInfo.Real,
		Version: gv.f.networkInfo.Version,
		Commit:  gv.f.networkInfo.Commit,
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
			Ts:    gv.f.curTime,
			Speed: gv.f.curSpeed,
		}},
	}); err != nil {
		return err
	}

	// draw all nodes
	for nodeid, node := range gv.f.nodes {
		addNodeEvent := &pb.VisualizeEvent{Type: &pb.VisualizeEvent_AddNode{AddNode: &pb.AddNodeEvent{
			NodeId:     int32(nodeid),
			X:          int32(node.x),
			Y:          int32(node.y),
			RadioRange: int32(node.radioRange),
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
	}

	return nil
}

func (gv *grpcVisualizer) AddVisualizationEvent(event *pb.VisualizeEvent, trivial bool) {
	if gv.replay != nil {
		gv.replay.Append(event, trivial)
	}
	gv.server.SendEvent(event, trivial)
}

func (gv *grpcVisualizer) UpdateNodesEnergy(nodes []*pb.NodeEnergy, timestamp uint64, updateView bool) {
	gv.Lock()
	defer gv.Unlock()

	//simplelogger.Debugf("Updating Nodes Energy to the charts")
	gv.server.SendEnergyEvent(&pb.NetworkEnergyEvent{
		Timestamp:   timestamp / 1000000, // convert to s
		NodesEnergy: nodes,
	})
	if updateView {
		gv.server.SendEnergyEvent(&pb.NetworkEnergyEvent{
			Timestamp:   math.MaxUint64, // convert to s
			NodesEnergy: make([]*pb.NodeEnergy, 0),
		},
		)
	}
}

func (gc *grpcVisualizer) SetEnergyAnalyser(ea *energy.EnergyAnalyser) {
	gc.Lock()
	defer gc.Unlock()

	gc.energyAnalyser = ea
}

func NewGrpcVisualizer(address string, replayFn string) visualize.Visualizer {
	gsv := &grpcVisualizer{
		simctrl: nil,
		f:       newGrpcField(),
	}

	if replayFn != "" {
		gsv.replay = replay.NewReplay(replayFn)
	}

	gsv.server = newGrpcServer(gsv, address)
	return gsv
}
