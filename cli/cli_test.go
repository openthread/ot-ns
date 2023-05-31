// Copyright (c) 2020-2023, The OTNS Authors.
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
	"testing"
	"time"

	. "github.com/openthread/ot-ns/types"
	"github.com/stretchr/testify/assert"
)

func TestParseBytes(t *testing.T) {
	var cmd Command
	err := ParseBytes([]byte("wrongcmd"), &cmd)
	assert.NotNil(t, err)

	assert.Nil(t, ParseBytes([]byte("add router"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == ROUTER)
	assert.Nil(t, ParseBytes([]byte("add med"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == MED)
	assert.Nil(t, ParseBytes([]byte("add sed"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == SED)
	assert.Nil(t, ParseBytes([]byte("add fed"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == FED)
	assert.Nil(t, ParseBytes([]byte("add router x 100 y 200"), &cmd))
	assert.True(t, *cmd.Add.X == 100 && *cmd.Add.Y == 200)
	assert.Nil(t, ParseBytes([]byte("add router id 100"), &cmd))
	assert.True(t, cmd.Add.Id.Val == 100)
	assert.Nil(t, ParseBytes([]byte("add router rr 1234"), &cmd))
	assert.True(t, cmd.Add.RadioRange.Val == 1234)
	assert.Nil(t, ParseBytes([]byte("add router x 1 y 2 id 3 rr 1234"), &cmd))
	assert.Nil(t, ParseBytes([]byte("add router rr 1234 id 3 y 2 x 1"), &cmd))

	assert.True(t, ParseBytes([]byte("countdown 3"), &cmd) == nil && cmd.CountDown != nil)
	assert.True(t, ParseBytes([]byte("countdown 3 \"abc\""), &cmd) == nil && cmd.CountDown != nil)

	assert.True(t, ParseBytes([]byte("counters"), &cmd) == nil && cmd.Counters != nil)

	assert.True(t, ParseBytes([]byte("del 1"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, ParseBytes([]byte("del 1 2"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, ParseBytes([]byte("del 1 2 3"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, ParseBytes([]byte("del"), &cmd) != nil)

	assert.True(t, ParseBytes([]byte("demo_legend \"title\" 100 200"), &cmd) == nil && cmd.DemoLegend != nil)

	assert.True(t, ParseBytes([]byte("exe mtd \"MyExecutable_thingy\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, ParseBytes([]byte("exe ftd \"./path/to/my/ot-cli-ftd\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, ParseBytes([]byte("exe br \"./path/to/my/br-script.sh\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, ParseBytes([]byte("exe"), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, ParseBytes([]byte("exe default"), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, ParseBytes([]byte("exe v12"), &cmd) == nil && cmd.Exe != nil)

	assert.True(t, ParseBytes([]byte("exit"), &cmd) == nil && cmd.Exit != nil)

	assert.Nil(t, ParseBytes([]byte("go 1"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, ParseBytes([]byte("go 1.1"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, ParseBytes([]byte("go 64us"), &cmd))
	assert.NotNil(t, cmd.Go)
	parsedDuration, _ := time.ParseDuration("64us")
	assert.Equal(t, 64*time.Microsecond, parsedDuration)
	assert.Nil(t, ParseBytes([]byte("go 5h"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, ParseBytes([]byte("go ever"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, ParseBytes([]byte("go 100 speed 0.5"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, ParseBytes([]byte("go 100 speed 2"), &cmd))
	assert.NotNil(t, cmd.Go)

	assert.True(t, ParseBytes([]byte("joins"), &cmd) == nil && cmd.Joins != nil)

	assert.True(t, ParseBytes([]byte("log"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, ParseBytes([]byte("log debug"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, ParseBytes([]byte("log info"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, ParseBytes([]byte("log warn"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, ParseBytes([]byte("log error"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, ParseBytes([]byte("log fatal"), &cmd) != nil && cmd.LogLevel != nil) // not supported.

	assert.True(t, ParseBytes([]byte("move 1 200 300"), &cmd) == nil && cmd.Move != nil)

	assert.True(t, ParseBytes([]byte("node 1 \"cmd\""), &cmd) == nil && cmd.Node != nil, cmd.Node.Command != nil)
	assert.True(t, ParseBytes([]byte("node 1"), &cmd) == nil && cmd.Node != nil && cmd.Node.Command == nil)

	assert.True(t, ParseBytes([]byte("nodes"), &cmd) == nil && cmd.Nodes != nil)

	assert.True(t, ParseBytes([]byte("partitions"), &cmd) == nil && cmd.Partitions != nil)
	assert.True(t, ParseBytes([]byte("pts"), &cmd) == nil && cmd.Partitions != nil)

	assert.True(t, ParseBytes([]byte("ping 1 2"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 any"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 mleid"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 aloc"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 rloc"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 linklocal"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 \"2001::1\""), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 datasize 100"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.DataSize.Val == 100)
	assert.True(t, ParseBytes([]byte("ping 1 2 interval 6"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.Interval.Val == 6)
	assert.True(t, ParseBytes([]byte("ping 1 2 hoplimit 3"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.HopLimit.Val == 3)
	assert.True(t, ParseBytes([]byte("ping 1 2 datasize 20 interval 3 hoplimit 60"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("ping 1 2 datasize 20 hoplimit 60 interval 3"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, ParseBytes([]byte("pings"), &cmd) == nil && cmd.Pings != nil)

	assert.True(t, ParseBytes([]byte("plr"), &cmd) == nil && cmd.Plr != nil && cmd.Plr.Val == nil)
	assert.True(t, ParseBytes([]byte("plr 1"), &cmd) == nil && cmd.Plr != nil && *cmd.Plr.Val == 1)

	assert.True(t, ParseBytes([]byte("radio 1 on"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, ParseBytes([]byte("radio 1 off"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, ParseBytes([]byte("radio 1 2 3 on"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, ParseBytes([]byte("radio 4 5 6 off"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, ParseBytes([]byte("radio 4 5 6 ft 10 60"), &cmd) == nil && cmd.Radio != nil)

	assert.True(t, ParseBytes([]byte("scan 1"), &cmd) == nil && cmd.Scan != nil)

	assert.True(t, ParseBytes([]byte("speed"), &cmd) == nil && cmd.Speed != nil && cmd.Speed.Speed == nil)
	assert.True(t, ParseBytes([]byte("speed 1.5"), &cmd) == nil && cmd.Speed != nil && *cmd.Speed.Speed == 1.5)

	assert.True(t, ParseBytes([]byte("time"), &cmd) == nil && cmd.Time != nil)

	assert.True(t, ParseBytes([]byte("watch"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes == nil)
	assert.True(t, ParseBytes([]byte("watch all"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes == nil && cmd.Watch.All == "all")
	assert.True(t, ParseBytes([]byte("watch 2 5 6"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil)
	assert.True(t, ParseBytes([]byte("watch 1 2 5 6 debug"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil &&
		len(cmd.Watch.Level) == 5)
	assert.True(t, ParseBytes([]byte("watch default T"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes == nil &&
		len(cmd.Watch.Level) == 1)
	assert.True(t, ParseBytes([]byte("watch 1 2 5 6 I"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil &&
		len(cmd.Watch.Level) == 1)
	assert.True(t, ParseBytes([]byte("unwatch 2 5 6"), &cmd) == nil && cmd.Unwatch != nil)
	assert.True(t, ParseBytes([]byte("unwatch all"), &cmd) == nil && cmd.Unwatch != nil)

	assert.True(t, ParseBytes([]byte("web"), &cmd) == nil && cmd.Web != nil)
}

func TestContextlessCommandPat(t *testing.T) {
	assert.True(t, contextLessCommandsPat.MatchString("exit"))
	assert.True(t, contextLessCommandsPat.MatchString("node 1"))
	assert.True(t, contextLessCommandsPat.MatchString("!nodes"))
	assert.True(t, contextLessCommandsPat.MatchString("!ping 23 24"))
}
