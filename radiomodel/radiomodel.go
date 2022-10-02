package radiomodel

import (
	"math"
	"math/rand"

	. "github.com/openthread/ot-ns/types"
)

// IEEE 802.15.4-2015 related parameters
type DbmValue = int8
type ChannelId = uint8

const (
	MinChannelNumber ChannelId = 0
	MaxChannelNumber ChannelId = 26
)

// default radio parameters
const (
	receiveSensitivityDbm        DbmValue = -100  // TODO for now MUST be manually kept equal to OT: SIM_RECEIVE_SENSITIVITY
	DefaultTxPowerDbm            DbmValue = 0     // Default, RadioTxEvent msg will override it. OT: SIM_TX_POWER
	DefaultCcaEdThresholdDbm     DbmValue = -91   // Default, RadioTxEvent msg will override it. OT: SIM_CCA_ENERGY_DETECT_THRESHOLD
	radioRangeIndoorDistInMeters          = 26.70 // Handtuned - for indoor model, how many meters r is RadioRange disc until Link
	// quality drops below 2 (10 dB margin).
)

// RSSI parameter encodings
const (
	RssiInvalid       DbmValue = 127
	RssiMax           DbmValue = 126
	RssiMin           DbmValue = -126
	RssiMinusInfinity DbmValue = -127
)

// EventQueue is the abstraction of the queue where the radio model sends its outgoing (new) events to.
type EventQueue interface {
	Add(*Event)
}

// RadioModel provides access to any type of radio model.
type RadioModel interface {
	// CheckRadioReachable checks if the srcNode radio can reach the dstNode radio, now.
	CheckRadioReachable(srcNode *RadioNode, dstNode *RadioNode) bool

	// GetTxRssi calculates at what RSSI level a radio frame Tx would be received by
	// dstNode, according to the radio model, in the ideal case of no other transmitters/interferers.
	// It returns the expected RSSI value at dstNode, or RssiMinusInfinity if the RSSI value will
	// fall below the minimum Rx sensitivity of the dstNode.
	GetTxRssi(srcNode *RadioNode, dstNode *RadioNode) DbmValue

	// OnEventDispatch is called when the dispatcher sends an Event to a particular dstNode. The method
	// implementation may e.g. apply interference to a frame in transit, prior to delivery of the
	// frame at a single receiving radio dstNode.
	OnEventDispatch(evt *Event, srcNode *RadioNode, dstNode *RadioNode)

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node must be the RadioNode object equivalent to the evt.NodeId node. Newly generated events may go back into
	// the EventQueue q.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)

	// GetName gets the display name of this RadioModel
	GetName() string

	// init initializes the RadioModel
	init()
}

// Create creates a new RadioModel with given name, or nil if model not found.
func Create(modelName string) RadioModel {
	var model RadioModel
	switch modelName {
	case "Ideal":
		model = &RadioModelIdeal{
			Name:      modelName,
			FixedRssi: -60,
		}
	case "Ideal_Rssi":
		model = &RadioModelIdeal{
			Name:            modelName,
			UseVariableRssi: true,
		}
	case "MutualInterference":
		model = &RadioModelMutualInterference{
			MinSirDb: 1, // minimum Signal-to-Interference (SIR) (dB) required to detect signal
		}
	default:
		model = nil
	}
	if model != nil {
		model.init()
	}
	return model
}

// interferePsduData simulates the interference (garbling) of PSDU data based on a given SIR level (dB).
func interferePsduData(data []byte, sirDb float64) []byte {
	intfData := data
	if sirDb < 0 {
		rand.Read(intfData)
	}
	intfData[0] = data[0] // keep channel info correct.
	return intfData
}

// computeIndoorRssi computes the RSSI for a receiver at distance dist, using a simple indoor exponent=3.xx loss model.
func computeIndoorRssi(srcRadioRange float64, dist float64, txPower int8, rxSensitivity int8) int8 {
	pathloss := 0.0
	distMeters := dist * radioRangeIndoorDistInMeters / srcRadioRange
	if distMeters >= 0.072 {
		pathloss = 35.0*math.Log10(distMeters) + 40.0
	}
	rssi := float64(txPower) - pathloss
	rssiInt := int(math.Round(rssi))
	// constrain RSSI value to int8 and return it. If RSSI is below the receiver's rxSensitivity,
	// then return the RssiMinusInfinity value.
	if rssiInt >= int(RssiInvalid) {
		rssiInt = int(RssiMax)
	} else if rssiInt < int(RssiMin) || rssiInt < int(rxSensitivity) {
		rssiInt = int(RssiMinusInfinity)
	}
	return int8(rssiInt)
}
