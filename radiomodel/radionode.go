package radiomodel

import (
	"math"

	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	Id NodeId

	// TxPower contains the last Tx power used by the node.
	TxPower int8

	// RxSensitivity contains the Rx sensitivity in dBm of the node.
	RxSensitivity int8

	// TimeLastTxEnded is the timestamp (us) when the last Tx (attempt) by this RadioNode ended.
	TimeLastTxEnded uint64

	// RadioRange is the radio range as configured by the simulation for this node.
	RadioRange float64

	// RadioState is the current radio's state.
	RadioState RadioStates

	// RadioChannel is the current radio's channel (For Rx, Tx, or sampling).
	RadioChannel uint8

	// Node position expressed in dimensionless units.
	X, Y float64

	// interferedBy indicates by which other node this RadioNode was interfered during current transmission.
	interferedBy map[NodeId]*RadioNode

	// receivingFrom indicates from which other node this RadioNode is correctly receiving (from the start).
	receivingFrom NodeId

	// rssiSampleMax tracks the max RSSI detected during a channel sampling operation.
	rssiSampleMax int8
}

func NewRadioNode(nodeid NodeId, cfg *NodeConfig) *RadioNode {
	rn := &RadioNode{
		Id:            nodeid,
		TxPower:       defaultTxPowerDbm,
		RxSensitivity: receiveSensitivityDbm,
		X:             float64(cfg.X),
		Y:             float64(cfg.Y),
		RadioRange:    float64(cfg.RadioRange),
		RadioChannel:  DefaultChannelNumber,
		interferedBy:  make(map[NodeId]*RadioNode),
		receivingFrom: 0,
		rssiSampleMax: RssiMinusInfinity,
	}
	return rn
}

func (rn *RadioNode) SetChannel(ch ChannelId) {
	simplelogger.AssertTrue(ch >= MinChannelNumber && ch <= MaxChannelNumber)
	// if changing channel during rx, fail the rx.
	if ch != rn.RadioChannel {
		rn.receivingFrom = 0
	}
	rn.RadioChannel = ch
}

func (rn *RadioNode) SetRadioState(state RadioStates) {
	// if changing state during rx, fail the rx.
	if state != rn.RadioState {
		rn.receivingFrom = 0
	}
	rn.RadioState = state
}

func (rn *RadioNode) SetNodePos(x int, y int) {
	// simplified model: ignore pos changes during Rx.
	rn.X, rn.Y = float64(x), float64(y)
}

// GetDistanceInMeters gets the distance to another RadioNode (in dimensionless units).
func (rn *RadioNode) GetDistanceTo(other *RadioNode) (dist float64) {
	dx := other.X - rn.X
	dy := other.Y - rn.Y
	dist = math.Sqrt(dx*dx + dy*dy)
	return
}
