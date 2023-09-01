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

package energy

import (
	. "github.com/openthread/ot-ns/types"
)

/*
 * Default consumption values by state of STM32WB55rg at 3.3V.
 * Consumption in kilowatts, time in microseconds, resulting energy in mJ.
 */
const (
	RadioDisabledConsumption float64 = 0.00000011 //kilowatts, to be confirmed
	RadioTxConsumption       float64 = 0.00001716 //kilowatts @ i = 5.2 mA
	RadioRxConsumption       float64 = 0.00001485 //kilowatts @ i = 4.5 mA
	RadioSleepConsumption    float64 = 0.00001485 //kilowatts @ i = 4.5 mA
)

const (
	ComputePeriod uint64 = 30000000 // in microseconds
)

type RadioStatus struct {
	State         RadioStates
	SpentDisabled uint64
	SpentSleep    uint64
	SpentTx       uint64
	SpentRx       uint64
	Timestamp     uint64
}

type NetworkConsumption struct {
	Timestamp          uint64
	EnergyConsDisabled float64
	EnergyConsSleep    float64
	EnergyConsTx       float64
	EnergyConsRx       float64
}
