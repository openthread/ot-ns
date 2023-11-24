// Copyright (c) 2023, The OTNS Authors.
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

import (
	"math"
	"math/rand"

	"github.com/openthread/ot-ns/logger"
)

const (
	initialCacheSize = 10000
	maxCacheSize     = 5000000
)

type fadingModel struct {
	rndSeed          int64
	ts               uint64
	shFadeMap        map[int64]DbValue
	tvFadeMap        map[int64]DbValue
	tvFadeSigmaMap   map[int64]DbValue
	changeTvfTimeMap map[int64]uint64
}

func newFadingModel() *fadingModel {
	sf := &fadingModel{
		rndSeed:          rand.Int63(),
		ts:               0,
		shFadeMap:        make(map[int64]DbValue, initialCacheSize),
		tvFadeSigmaMap:   make(map[int64]DbValue, initialCacheSize),
		tvFadeMap:        make(map[int64]DbValue, initialCacheSize),
		changeTvfTimeMap: make(map[int64]uint64, initialCacheSize),
	}
	return sf
}

// computeFading calculates shadow fading (SF) and time-variant fading for a radio link based on a simple random process.
//
// SF: models a fixed, position-dependent radio signal power attenuation (SF>0) or increase (SF<0) due to multipath effects
// and static obstacles. In the dB domain it is modeled as a normal distribution (mu=0, sigma).
// A symmetric link is assumed between transmitter/receiver. (I.e. reversing roles gives same SF value.)
// See https://en.wikipedia.org/wiki/Fading and 3GPP TR 38.901 V17.0.0, section 7.4.1 and 7.4.4, and
// Table 7.5-6 Part-2.
//
// TVF: time-variant fading or 'RSSI variation over time' is described in research, however a comprehensive model wasn't
// found yet at the time of writing this. Some pointers can be found in:
//   - https://www.researchgate.net/publication/258222418
//   - https://www.researchgate.net/publication/224156949
//   - https://www.researchgate.net/publication/336096154
//   - https://www.fed4fire.eu/wp-content/uploads/sites/10/2019/04/10.wns3-2018-paper_camera-ready-version.pdf
//   - https://uwspace.uwaterloo.ca/bitstream/handle/10012/16230/Jacob_Midul.pdf?sequence=3&isAllowed=y
//
// TODO: better implement the autocorrelation of SF over a correlation length d_cor = 6 m (NLOS case)
func (sf *fadingModel) computeFading(src *RadioNode, dst *RadioNode, params *RadioModelParams) DbValue {
	// each unique (src,dst) link gets a unique random seed
	seed := sf.rndSeed + calcLinkUID(src, dst, params.MeterPerUnit)

	var vSF, vTVF float64
	if v, ok := sf.shFadeMap[seed]; ok { // look up if that seed (radio link) was already precomputed.
		vSF = v
		vTVF = sf.tvFadeMap[seed]
		if sf.ts > sf.changeTvfTimeMap[seed] { // TVF dB may require occassional regeneration (randomly)
			sigmaTVF := sf.tvFadeSigmaMap[seed]
			vTVF = rand.NormFloat64() * sigmaTVF // new random TV fade-value, based on existing sigma
			sf.tvFadeMap[seed] = vTVF
			nextChangeDeltaSec := rand.ExpFloat64() * params.MeanTimeFadingChange
			sf.changeTvfTimeMap[seed] = sf.ts + uint64(nextChangeDeltaSec*1e6) // pick next change time.
		}
	} else { // if not, compute the values
		rndSource := rand.NewSource(seed)
		rnd := rand.New(rndSource)

		// draw a single (reproducible) random number based on the link's unique seed, and store it.
		vSF = rnd.NormFloat64() * params.ShadowFadingSigmaDb
		sf.shFadeMap[seed] = vSF

		// draw a second number (reproducible) for sigma and store it.
		sigmaTVF := rnd.Float64() * params.TimeFadingSigmaMaxDb
		sf.tvFadeSigmaMap[seed] = sigmaTVF

		vTVF = rand.NormFloat64() * sigmaTVF // compute TVF amount (random over time);
		sf.tvFadeMap[seed] = vTVF            // and store it.
		nextChangeDeltaSec := rand.ExpFloat64() * params.MeanTimeFadingChange
		sf.changeTvfTimeMap[seed] = sf.ts + uint64(nextChangeDeltaSec*1e6) // pick next change time.
	}

	return vSF + vTVF
}

func (sf *fadingModel) onAdvanceTime(ts uint64) {
	// if storage gets too big, purge it - items will be recomputed (and thus slow down the simulation a bit)
	// this normally would only happen with long simulations with moving nodes.
	if len(sf.shFadeMap) > maxCacheSize {
		sf.clearCaches()
	}
	sf.ts = ts
}

func (sf *fadingModel) clearCaches() {
	logger.Debugf("Radio fading model: purging fadeMap caches")
	sf.shFadeMap = make(map[int64]DbValue, initialCacheSize)
	sf.tvFadeSigmaMap = make(map[int64]DbValue, initialCacheSize)
	sf.tvFadeMap = make(map[int64]DbValue, initialCacheSize)
	sf.changeTvfTimeMap = make(map[int64]uint64, initialCacheSize)
}

func calcLinkUID(src *RadioNode, dst *RadioNode, meterPerUnit float64) int64 {
	// calc node positions in grid units of 5 m, using only positive values (uint16 range)
	x1 := uint16(math.Round(src.X*meterPerUnit*0.2) + 32768)
	y1 := uint16(math.Round(src.Y*meterPerUnit*0.2) + 32768)
	x2 := uint16(math.Round(dst.X*meterPerUnit*0.2) + 32768)
	y2 := uint16(math.Round(dst.Y*meterPerUnit*0.2) + 32768)
	xL := x2
	yL := y2
	xR := x1
	yR := y1

	// use left-most node (and in case of doubt, top-most) - screen coordinates
	if x1 < x2 || (x1 == x2 && y1 < y2) {
		xL = x1
		yL = y1
		xR = x2
		yR = y2
	}

	// Give each (xL,yL) & (xR,yR) coordinate combination its own fixed int64 seed-value.
	// Also each radio-channel could have its own seed, by adding:  424242*int64(src.RadioChannel)
	// However, radio-channels are relatively near in spectrum terms, so we may assume little influence on SF.
	// TODO: in case sub-GHz channels are allowed, this argument doesn't hold anymore.
	uid := int64(xL) + int64(yL)<<16 + int64(xR)<<32 + int64(yR)<<48
	return uid
}
