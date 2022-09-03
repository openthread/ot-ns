package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
	"math"
	"math/rand"
)

// IEEE 802.15.4-2015 O-QPSK PHY
const symbolTimeUs uint64 = 16
const symbolsPerOctet = 2
const aMaxSifsFrameSize = 18 // as defined in IEEE 802.15.4-2015
const phyHeaderSize = 6
const ccaTimeUs = symbolTimeUs * 8
const aifsTimeUs = symbolTimeUs * 12
const lifsTimeUs = symbolTimeUs * 20
const sifsTimeUs = symbolTimeUs * 12

// default radio parameters
const receiveSensitivityDbm = -100        // TODO for now MUST be manually kept equal to OT: SIM_RECEIVE_SENSITIVITY
const txPowerDbm = 0                      // Default, event msg Param1 will override it. OT: SIM_TX_POWER
const ccaEdThresholdDbm = -91             // Default, event msg Param2 will override it. OT: SIM_CCA_ENERGY_DETECT_THRESHOLD
const radioRangeIndoorDistInMeters = 37.0 // Handtuned - for indoor model, how many meters r is RadioRange disc until Link
// quality drops below 1.

// RSSI parameter encodings
const RssiInvalid = 127
const RssiMax = 126
const RssiMin = -126
const RssiMinusInfinity = -127

// EventQueue is the abstraction of the queue where the radio model sends its outgoing (new) events to.
type EventQueue interface {
	AddEvent(*Event)
}

// RadioModel provides access to any type of radio model.
type RadioModel interface {
	CheckRadioReachable(evt *Event, srcNode *RadioNode, dstNode *RadioNode) bool

	// GetTxRssi calculates at what RSSI level the radio frame Tx indicated by evt would be received by
	// dstNode, according to the radio model, in the ideal case of no other transmitters/interferers.
	// It returns the expected RSSI value at dstNode, or RssiMinusInfinity if the RSSI value will
	// fall below the minimum Rx sensitivity of the dstNode.
	GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode) int8

	// ApplyInterference applies any interference to a frame in transit, prior to delivery of the
	// frame at a single receiving radio dstNode.
	ApplyInterference(evt *Event, srcNode *RadioNode, dstNode *RadioNode)

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node must be the RadioNode object equivalent to the evt.NodeId node. Newly generated events may go back into
	// the EventQueue q.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)

	// GetName gets the display name of this RadioModel
	GetName() string

	// AllowsUnicastDispatch indicates whether the model allows efficient unicast dispatch of radio-frame
	// received events to only the Node(s) that are addressed by the frame. If false, the dispatcher will
	// dispatch a frame to all nodes in radio range regardless of MAC header addressing.
	AllowUnicastDispatch() bool
}

// Create creates a new RadioModel with given name, or nil if model not found.
func Create(modelName string) RadioModel {
	var model RadioModel = nil
	switch modelName {
	case "Ideal":
		model = &RadioModelIdeal{
			Name:               modelName,
			FixedFrameDuration: 1,
			FixedRssi:          -60,
		}
	case "Ideal_Rssi":
		model = &RadioModelIdeal{
			Name:               modelName,
			UseVariableRssi:    true,
			FixedFrameDuration: 1,
		}
	case "Ideal_Rssi_Dur":
		model = &RadioModelIdeal{
			Name:                 modelName,
			UseVariableRssi:      true,
			UseRealFrameDuration: true,
		}
	case "MutualInterference":
		model = &RadioModelMutualInterference{
			ActiveTransmitters: make(map[NodeId]*RadioNode),
			MinSirDb:           1, // minimum Signal-to-Interference (SIR) (dB) required to detect signal
		}
	}
	return model
}

// IsLongDataFrame checks whether the radio frame in evt is 802.15.4 "long" (true) or not.
func IsLongDataframe(evt *Event) bool {
	return (len(evt.Data) - RadioMessagePsduOffset) > aMaxSifsFrameSize
}

// getFrameDurationUs gets the duration of the PHY frame in us indicated by evt of type eventTypeRadioFrame*
func getFrameDurationUs(evt *Event) uint64 {
	var n uint64
	simplelogger.AssertTrue(len(evt.Data) >= RadioMessagePsduOffset)
	n = (uint64)(len(evt.Data) - RadioMessagePsduOffset) // PSDU size 5..127
	n += phyHeaderSize                                   // add PHY preamble, sfd, PHR bytes
	return n * symbolTimeUs * symbolsPerOctet
}

// InterferePsduData simulates the interference (garbling) of PSDU data based on a given SIR level (dB).
func InterferePsduData(data []byte, sirDb float64) []byte {
	intfData := data
	if sirDb < 0 {
		rand.Read(intfData)
	}
	intfData[0] = data[0] // keep channel info correct.
	return intfData
}

// ComputeIndoorRssi computes the RSSI for a receiver at distance dist, using a simple indoor exponent=3.xx loss model.
func ComputeIndoorRssi(srcRadioRange float64, dist float64, txPower int8, rxSensitivity int8) int8 {
	pathloss := 0.0
	distMeters := dist * radioRangeIndoorDistInMeters / srcRadioRange
	if distMeters >= 0.072 {
		pathloss = 35.0*math.Log10(distMeters) + 40.0
	}
	rssi := float64(txPower) - pathloss
	rssiInt := int(math.Round(rssi))
	// constrain RSSI value to int8 and return it. If RSSI is below the receiver's rxSensitivity,
	// then return the RssiMinusInfinity value.
	if rssiInt >= RssiInvalid {
		rssiInt = RssiMax
	} else if rssiInt < RssiMin || rssiInt < int(rxSensitivity) {
		rssiInt = RssiMinusInfinity
	}
	return int8(rssiInt)
}
