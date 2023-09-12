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

	"github.com/simonlingoogle/go-simplelogger"
)

const (
	maxFadeMapSize = 5000000
)

type shadowFading struct {
	rndSeed int64
	fadeMap map[int64]DbValue
}

func newShadowFading() *shadowFading {
	sf := &shadowFading{
		rndSeed: rand.Int63(),
		fadeMap: make(map[int64]DbValue, 1000),
	}
	return sf
}

// computeShadowFading calculates shadow fading (SF) for a radio link based on a simple random process.
// It models a fixed, position-dependent radio signal power attenuation (SF>0) or increase (SF<0) due to multipath effects
// and static obstacles. In the dB domain it is modeled as a normal distribution (mu=0, sigma).
// See https://en.wikipedia.org/wiki/Fading and 3GPP TR 38.901 V17.0.0, section 7.4.1 and 7.4.4, and
// Table 7.5-6 Part-2.
// TODO: better implement the autocorrelation of SF over a correlation length d_cor = 6 m (NLOS case)
func (sf *shadowFading) computeShadowFading(src *RadioNode, dst *RadioNode, params *RadioModelParams) DbValue {
	if params.ShadowFadingSigmaDb <= 0 {
		return 0.0
	}

	// calc node positions in grid units of 5 m, using only positive values (uint16 range)
	x1 := uint16(math.Round(src.X*params.MeterPerUnit*0.2) + 32768)
	y1 := uint16(math.Round(src.Y*params.MeterPerUnit*0.2) + 32768)
	x2 := uint16(math.Round(dst.X*params.MeterPerUnit*0.2) + 32768)
	y2 := uint16(math.Round(dst.Y*params.MeterPerUnit*0.2) + 32768)
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

	// give each (xL,yL) & (xR,yR) coordinate combination on each RF channel its own fixed int64 seed-value.
	seed := sf.rndSeed + 424242*int64(src.RadioChannel) + int64(xL) + int64(yL)<<16 + int64(xR)<<32 + int64(yR)<<48

	// look up if that value was already precomputed.
	if v, ok := sf.fadeMap[seed]; ok {
		return v
	}

	// if not, compute the value
	rndSource := rand.NewSource(seed)
	rnd := rand.New(rndSource)
	// draw a single (reproducible) random number based on the position coordinates.
	v := rnd.NormFloat64() * params.ShadowFadingSigmaDb

	// and store it
	sf.fadeMap[seed] = v

	// if storage gets too big, purge it - will be recomputed (and thus slow down the simulation a bit)
	// this normally would only happen with long simulations with moving nodes.
	if len(sf.fadeMap) > maxFadeMapSize {
		simplelogger.Debugf("shadowFading model: purging fadeMap cache")
		sf.fadeMap = make(map[int64]DbValue, 10000)
		sf.fadeMap[seed] = v
	}
	return v
}
