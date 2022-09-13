package radiomodel

import (
	"math"

	. "github.com/openthread/ot-ns/types"
)

// RadioNode is the status of a single radio node of the radio model, used by all radio models.
type RadioNode struct {
	// IsCcaFailed tracks whether the last CCA process failed (true), or not (false).
	IsCcaFailed bool

	// IsTxFailed tracks whether the current/last Tx attempt failed (true), or not (false).
	IsTxFailed bool

	// TxPhase tracks the current Tx phase. 0 = Not started. >0 is started (exact value depends on radio model)
	TxPhase int

	// TxPower contains the last Tx power used by the OpenThread node.
	TxPower int8

	// CcaEdThresh contains the last used/set CCA ED threshold of the OpenThread node.
	CcaEdThresh int8

	// RxSensitivity contains the Rx sensitivity in dBm (not influenced by OpenThread node.)
	RxSensitivity int8

	// TimeLastTxEnded is the timestamp (us) when the last Tx or Tx-attempt by this RadioNode ended.
	TimeLastTxEnded uint64

	// IsLastTxLong indicates whether the RadioNode's last Tx was a long frame (LIFS applies) or short (SIFS applies)
	IsLastTxLong bool

	// InterferedBy indicates by which other node this RadioNode was interfered during current transmission.
	InterferedBy map[NodeId]*RadioNode

	// IsDeleted tracks whether this node has been deleted in the simulation.
	IsDeleted bool

	// Node position expressed in dimensionless units.
	X, Y float64

	// RadioRange is the max allowed radio range as configured by the simulation for this node.
	RadioRange float64
}

func NewRadioNode(cfg *NodeConfig) *RadioNode {
	rn := &RadioNode{
		TxPower:       DefaultTxPowerDbm,
		CcaEdThresh:   DefaultCcaEdThresholdDbm,
		RxSensitivity: receiveSensitivityDbm,
		X:             float64(cfg.X),
		Y:             float64(cfg.Y),
		RadioRange:    float64(cfg.RadioRange),
		InterferedBy:  make(map[NodeId]*RadioNode),
	}
	return rn
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
