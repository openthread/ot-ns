package radiomodel

import (
	. "github.com/openthread/ot-ns/types"
	"math"
)

// IEEE 802.15.4-2015 O-QPSK PHY
const symbolTimeUs uint64 = 16
const symbolsPerOctet = 2
const freqHz = 2400000000

// aMaxSifsFrameSize as defined in IEEE 802.15.4-2015
const aMaxSifsFrameSize = 18
const phyHeaderSize = 6
const ccaTimeUs = symbolTimeUs * 8
const aifsTimeUs = symbolTimeUs * 12
const lifsTimeUs = symbolTimeUs * 20
const sifsTimeUs = symbolTimeUs * 12

// default radio parameters
const receiveSensitivityDbm = -100 // TODO for now MUST be equal to const in OpenThread node. Would like to adapt.
const RssiInvalid = 127
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
}

// IsLongDataFrame checks whether the radio frame in evt is 802.15.4 "long" (true) or not.
func IsLongDataframe(evt *Event) bool {
	return (len(evt.Data) - RadioMessagePsduOffset) > aMaxSifsFrameSize
}

// InterferePsduData simulates the interference (garbling) of PSDU data
func InterferePsduData(d []byte) []byte {
	ret := make([]byte, len(d))
	ret[0] = d[0] // keep channel info
	return ret
}

// ComputeIndoorRssi computes the RSSI for a receiver at distance dist, using a simple indoor exponent=3.xx loss model.
func ComputeIndoorRssi(dist float64, txPower int8, rxSensitivity int8) int8 {
	fspl := 0.0
	if dist > 0.01 {
		fspl = 30.0*math.Log10(dist) + 40.0
	}
	rssi := float64(txPower) - fspl
	rssiInt := int(math.Round(rssi))
	if rssiInt >= RssiInvalid {
		rssiInt = RssiInvalid // if too high, our int8 cannot hold the value anymore.
	} else if rssi <= RssiMinusInfinity || rssiInt < int(rxSensitivity) {
		rssiInt = RssiMinusInfinity
	}
	return int8(rssiInt)
	//return -76 // FIXME debugnod
}
