package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"github.com/simonlingoogle/go-simplelogger"
	"math"
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
const receiveSensitivityDbm = -100 // TODO for now MUST be manually kept equal to OT: SIM_RECEIVE_SENSITIVITY
const txPowerDbm = 0               // Default, event msg Param1 will override it. OT: SIM_TX_POWER
const ccaEdThresholdDbm = -75      // Default, event msg Param2 will override it. OT: SIM_CCA_ENERGY_DETECT_THRESHOLD

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
	// GetTxRssi checks whether the radio frame Tx indicated by evt is successful or not, according to the radio model.
	// It returns the RSSI value at dstNode if successful, or RssiMinusInfinity if not successful.
	GetTxRssi(evt *Event, srcNode *RadioNode, dstNode *RadioNode, distMeters float64) int8

	// HandleEvent handles all radio-model events coming out of the simulator event queue.
	// node must be the RadioNode object equivalent to the evt.NodeId node. Newly generated events may go back into
	// the EventQueue q.
	HandleEvent(node *RadioNode, q EventQueue, evt *Event)

	// GetName gets the display name of this RadioModel
	GetName() string
}

// Create creates a new RadioModel with given name, or nil if model not found.
func Create(modelName string) RadioModel {
	var model RadioModel = nil
	switch modelName {
	case "Ideal":
		model = &RadioModelIdeal{
			Name:               modelName,
			FixedFrameDuration: 1,
			FixedRssi:          -20,
		}
	case "Ideal_Rssi":
		model = &RadioModelIdeal{
			UseVariableRssi:    true,
			FixedFrameDuration: 1,
			Name:               modelName,
		}
	case "Ideal_Rssi_Dur":
		model = &RadioModelIdeal{
			UseVariableRssi:      true,
			UseRealFrameDuration: true,
			Name:                 modelName,
		}
	case "MutualInterference":
		model = &RadioModelMutualInterference{}
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

// InterferePsduData simulates the interference (garbling) of PSDU data
func InterferePsduData(d []byte) []byte {
	ret := make([]byte, len(d)) // make all 0 bytes
	ret[0] = d[0]               // keep channel info
	return ret
}

// ComputeIndoorRssi computes the RSSI for a receiver at distance dist, using a simple indoor exponent=3.xx loss model.
func ComputeIndoorRssi(dist float64, txPower int8, rxSensitivity int8) int8 {
	pathloss := 0.0
	if dist > 0.01 {
		pathloss = 35.0*math.Log10(dist) + 40.0
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
