// Copyright (c) 2020-2022, The OTNS Authors.
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

package otoutfilter

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/openthread/ot-ns/logger"
)

var (
	logPattern = regexp.MustCompile(`\[(-|C|W|N|I|D|CRIT|WARN|NOTE|INFO|DEBG)].+\n`)
)

type otOutFilter struct {
	linebuf        string
	subr           io.Reader
	logPrintPrefix string
	logHandler     func(otLevel string, logMsg string)
}

func (cc *otOutFilter) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	for {
		n := cc.readFirstLine(p)
		if n > 0 {
			return n, nil
		}

		var b [4096]byte
		n, err := cc.subr.Read(b[:])
		if err != nil {
			return n, err
		}

		cc.linebuf = cc.linebuf + string(b[:n])
	}
}

func (cc *otOutFilter) readFirstLine(p []byte) int {
	for {
		newLineIdx := strings.IndexByte(cc.linebuf, '\n')
		if newLineIdx == -1 {
			return 0
		}

		// first line is received completely, now we can read something
		// remove the log in the first line
		var rn int
		var sn int

		firstline := cc.linebuf[:newLineIdx+1]

		// remove > (the input prompt) to make cli output easier to parse
		if strings.HasPrefix(firstline, "> ") {
			firstline = firstline[2:]
			sn += 2
		}

		logIdx := logPattern.FindStringSubmatchIndex(firstline)

		if logIdx == nil {
			rn += copy(p, firstline[:])
		} else {
			// filter out the log line and send to printLog()
			logger.AssertTrue(logIdx[1] == len(firstline))
			logStr := strings.TrimSpace(firstline)
			logLevelIndicatorStr := firstline[logIdx[2] : logIdx[2]+1]
			logger.AssertTrue(len(logLevelIndicatorStr) == 1)
			cc.printLog(logLevelIndicatorStr, logStr)
			sn += logIdx[1]
		}

		logger.AssertTrue(rn+sn > 0) // should always read/skip something
		cc.linebuf = cc.linebuf[sn+rn:]
		if rn > 0 {
			return rn
		}
	}
}

func (cc *otOutFilter) printLog(otLevelChar string, logStr string) {
	if cc.logHandler == nil {
		return
	}
	logStr = fmt.Sprintf("%s%s", cc.logPrintPrefix, logStr)
	cc.logHandler(otLevelChar, logStr)
}

func NewOTOutFilter(reader io.Reader, logPrintPrefix string,
	handlerLogMsg func(otLevel string, msg string)) io.Reader {
	return &otOutFilter{subr: reader, logPrintPrefix: logPrintPrefix, logHandler: handlerLogMsg}
}

func DetectLogLine(line string) (bool, string) {
	logIdx := logPattern.FindStringSubmatchIndex(line)
	if logIdx == nil {
		return false, ""
	}
	logLevelIndicatorStr := line[logIdx[2] : logIdx[2]+1]
	return true, logLevelIndicatorStr
}
