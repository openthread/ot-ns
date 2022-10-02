package radiomodel

import (
	"math"

	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
)

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	// TxPower contains the last Tx power used by the node.
	TxPower int8

	// CcaEdThresh contains the last used/set CCA ED threshold of the node.
	CcaEdThresh int8

	// RxSensitivity contains the Rx sensitivity in dBm of the node.
	RxSensitivity int8

	// TimeLastTxEnded is the timestamp (us) when the last Tx (attempt) by this RadioNode ended.
	TimeLastTxEnded uint64

	// InterferedBy indicates by which other node this RadioNode was interfered during current transmission.
	InterferedBy map[NodeId]*RadioNode

	// IsDeleted tracks whether this node has been deleted in the simulation.
	IsDeleted bool

	// Node position expressed in dimensionless units.
	X, Y float64

	// RadioRange is the max allowed radio range as configured by the simulation for this node.
	RadioRange float64

	RadioState     RadioStates
	RadioChannel   uint8
	RadioLockState bool
	rssiSampleMax  int8
}

func NewRadioNode(cfg *NodeConfig) *RadioNode {
	rn := &RadioNode{
		TxPower:        DefaultTxPowerDbm,
		CcaEdThresh:    DefaultCcaEdThresholdDbm,
		RxSensitivity:  receiveSensitivityDbm,
		X:              float64(cfg.X),
		Y:              float64(cfg.Y),
		RadioRange:     float64(cfg.RadioRange),
		InterferedBy:   make(map[NodeId]*RadioNode),
		RadioChannel:   11,
		RadioLockState: false,
	}
	return rn
}

func (rn *RadioNode) SetChannel(ch ChannelId) {
	simplelogger.AssertTrue(ch >= MinChannelNumber && ch <= MaxChannelNumber)
	simplelogger.AssertFalse(rn.RadioLockState, "SetChannel(): radio state was locked")
	// FIXME: if changing channel during rx, fail it.
	if ch != rn.RadioChannel {
		// TODO
	}
	rn.RadioChannel = ch
}

func (rn *RadioNode) SetRadioState(state RadioStates) {
	simplelogger.AssertFalse(rn.RadioLockState, "SetRadioState(): radio state was locked")
	rn.RadioState = state
}

func (rn *RadioNode) LockRadioState(lock bool) {
	rn.RadioLockState = lock
}

func (rn *RadioNode) SetNodePos(x int, y int) {
	rn.X, rn.Y = float64(x), float64(y)
}

func (rn *RadioNode) Delete() {
	rn.IsDeleted = true
}

// GetDistanceInMeters gets the distance to another RadioNode (in dimensionless units).
func (rn *RadioNode) GetDistanceTo(other *RadioNode) (dist float64) {
	dx := other.X - rn.X
	dy := other.Y - rn.Y
	dist = math.Sqrt(dx*dx + dy*dy)
	return
}
