// Copyright (c) 2020-2024, The OTNS Authors.
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

package cli

import (
	"regexp"
	"sort"
)

var (
	contextLessCommandsPat = regexp.MustCompile(`(exit|node|!.+)\b`)
	backgroundCommandsPat  = regexp.MustCompile(`(discover|dns resolve|dns resolve4|dns browse|dns service|dns servicehost|scan|networkdiagnostic get|networkdiagnostic reset|meshdiag)\b`)
)

func isContextlessCommand(line string) bool {
	return contextLessCommandsPat.MatchString(line)
}

func isBackgroundCommand(cmd *Command) bool {
	if cmd.Node != nil && cmd.Node.Command != nil {
		return backgroundCommandsPat.MatchString(*cmd.Node.Command)
	}
	if cmd.Scan != nil {
		return true
	}
	return false
}

// getUniqueAndSorted returns a unique-ID'd and sorted version of []NodeSelector.
func getUniqueAndSorted(input []NodeSelector) []NodeSelector {
	u := make([]int, 0, len(input))
	m := make(map[int]NodeSelector, len(input))

	// find unique integers
	for _, ns := range input {
		if ns.All != nil { // if 'all' nodes are selected, return only the 'all' selector.
			return []NodeSelector{ns}
		}
		if nsExisting, ok := m[ns.Id]; ok {
			if nsExisting.IdRange > 0 || ns.IdRange == 0 {
				continue // only consider 1st occurrence of a given NodeId. Ranges have priority.
			}
		}
		m[ns.Id] = ns
	}

	// sort
	for id := range m {
		u = append(u, id)
	}
	sort.Ints(u)

	// output as []NodeSelector
	n := make([]NodeSelector, 0, len(u))
	for _, id := range u {
		n = append(n, m[id])
	}

	return n
}
