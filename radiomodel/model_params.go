// Copyright (c) 2022-2023, The OTNS Authors.
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

package radiomodel

import "math"

// default radio & simulation parameters
const (
	defaultNoiseFloorIndoorDbm DbValue = -95.0 // Indoor model ambient noise floor (dBm)
	defaultMeterPerUnit        float64 = 0.10  // Default distance equivalent in meters of one grid/pixel distance unit.
)

// RadioModelParams stores model parameters for the radio model.
type RadioModelParams struct {
	MeterPerUnit         float64 // the distance in meters, equivalent to a single distance unit(pixel)
	IsDiscLimit          bool    // If true, RF signal Tx range is limited to the RadioRange set for each node
	RssiMinDbm           DbValue // Lowest RSSI value (dBm) that can be returned, overriding other calculations
	RssiMaxDbm           DbValue // Highest RSSI value (dBm) that can be returned, overriding other calculations
	ExponentDb           DbValue // the exponent (dB) in the regular/LOS model
	FixedLossDb          DbValue // the fixed loss (dB) term in the regular/LOS model
	NlosExponentDb       DbValue // the exponent (dB) in the NLOS model
	NlosFixedLossDb      DbValue // the fixed loss (dB) term in the NLOS model
	NoiseFloorDbm        DbValue // the noise floor (ambient noise, in dBm)
	SnrMinThresholdDb    DbValue // the minimal value an SNR/SINR should be, to have a non-zero frame success probability.
	ShadowFadingSigmaDb  DbValue // sigma (stddev) parameter for Shadow Fading (SF), in dB
	TimeFadingSigmaMaxDb DbValue // max sigma (stddev) parameter for time-variant fading, in dB
	MeanTimeFadingChange float64 // mean time in sec, when TV fading value changes (mean of exponential distrib times).
}

// newRadioModelParams gets a new set of parameters with default values, as a basis to configure further.
func newRadioModelParams() *RadioModelParams {
	return &RadioModelParams{
		MeterPerUnit:         defaultMeterPerUnit,
		IsDiscLimit:          false,
		RssiMinDbm:           RssiMin,
		RssiMaxDbm:           RssiMax,
		ExponentDb:           UndefinedDbValue,
		FixedLossDb:          UndefinedDbValue,
		NlosExponentDb:       UndefinedDbValue,
		NlosFixedLossDb:      UndefinedDbValue,
		NoiseFloorDbm:        UndefinedDbValue,
		SnrMinThresholdDb:    UndefinedDbValue,
		ShadowFadingSigmaDb:  UndefinedDbValue,
		TimeFadingSigmaMaxDb: UndefinedDbValue,
		MeanTimeFadingChange: 0.0,
	}
}

// ITU-T model
func setIndoorModelParamsItu(params *RadioModelParams) {
	params.ExponentDb = 30.0
	params.FixedLossDb = paround(20.0*math.Log10(2400) - 28.0)
}

// see 3GPP TR 38.901 V17.0.0, Table 7.4.1-1: Pathloss models.
func setIndoorModelParams3gpp(params *RadioModelParams) {
	params.ExponentDb = 17.3
	params.FixedLossDb = paround(32.4 + 20*math.Log10(2.4))
	params.NlosExponentDb = 38.3
	params.NlosFixedLossDb = paround(17.3 + 24.9*math.Log10(2.4))
	params.NoiseFloorDbm = defaultNoiseFloorIndoorDbm
	params.SnrMinThresholdDb = -4.0 // see calcber.m Octave file
	params.ShadowFadingSigmaDb = 8.03
	params.TimeFadingSigmaMaxDb = 4.0 // TODO: this is unverified for 3GPP indoor model
	params.MeanTimeFadingChange = 150 // in seconds, TODO: this is unverified
}

// experimental outdoor model with LoS
func setOutdoorModelParams(params *RadioModelParams) {
	params.MeterPerUnit = 0.5
	params.ExponentDb = 17.3
	params.FixedLossDb = paround(32.4 + 20*math.Log10(2.4))
	params.NoiseFloorDbm = defaultNoiseFloorIndoorDbm
	params.SnrMinThresholdDb = -4.0 // see calcber.m Octave file
	params.ShadowFadingSigmaDb = 3.0
	params.TimeFadingSigmaMaxDb = 1.0 // TODO: this is unverified for outdoor
	params.MeanTimeFadingChange = 150 // in seconds, TODO: this is unverified
}
