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
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	. "github.com/openthread/ot-ns/types"
)

func TestParseBytes(t *testing.T) {
	var cmd Command
	err := parseBytes([]byte("wrongcmd"), &cmd)
	assert.NotNil(t, err)

	assert.Nil(t, parseBytes([]byte("add router"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == ROUTER)
	assert.Nil(t, parseBytes([]byte("add med"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == MED)
	assert.Nil(t, parseBytes([]byte("add sed"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == SED)
	assert.Nil(t, parseBytes([]byte("add fed"), &cmd))
	assert.True(t, cmd.Add != nil && cmd.Add.Type.Val == FED)
	assert.Nil(t, parseBytes([]byte("add router x 100 y 200"), &cmd))
	assert.True(t, *cmd.Add.X == 100 && *cmd.Add.Y == 200)
	assert.Nil(t, parseBytes([]byte("add router id 100"), &cmd))
	assert.True(t, cmd.Add.Id.Val == 100)
	assert.Nil(t, parseBytes([]byte("add router rr 1234"), &cmd))
	assert.True(t, cmd.Add.RadioRange.Val == 1234)
	assert.Nil(t, parseBytes([]byte("add router x 1 y 2 id 3 rr 1234"), &cmd))
	assert.Nil(t, parseBytes([]byte("add router rr 1234 id 3 y 2 x 1"), &cmd))

	assert.Nil(t, parseBytes([]byte("autogo"), &cmd))
	assert.NotNil(t, cmd.AutoGo)
	assert.Nil(t, parseBytes([]byte("autogo 1"), &cmd))
	assert.NotNil(t, cmd.AutoGo)
	assert.Nil(t, parseBytes([]byte("autogo 0"), &cmd))
	assert.NotNil(t, cmd.AutoGo)
	assert.Nil(t, parseBytes([]byte("autogo y"), &cmd))
	assert.NotNil(t, cmd.AutoGo)
	assert.Nil(t, parseBytes([]byte("autogo n"), &cmd))
	assert.NotNil(t, cmd.AutoGo)

	assert.True(t, parseBytes([]byte("countdown 3"), &cmd) == nil && cmd.CountDown != nil)
	assert.True(t, parseBytes([]byte("countdown 3 \"abc\""), &cmd) == nil && cmd.CountDown != nil)

	assert.True(t, parseBytes([]byte("counters"), &cmd) == nil && cmd.Counters != nil)

	assert.True(t, parseBytes([]byte("del 1"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, parseBytes([]byte("del 1 2"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, parseBytes([]byte("del 1 2 3"), &cmd) == nil && cmd.Del != nil)
	assert.True(t, parseBytes([]byte("del"), &cmd) != nil)

	assert.True(t, parseBytes([]byte("demo_legend \"title\" 100 200"), &cmd) == nil && cmd.DemoLegend != nil)

	assert.True(t, parseBytes([]byte("exe mtd \"MyExecutable_thingy\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, parseBytes([]byte("exe ftd \"./path/to/my/ot-cli-ftd\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, parseBytes([]byte("exe br \"./path/to/my/br-script.sh\""), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, parseBytes([]byte("exe"), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, parseBytes([]byte("exe default"), &cmd) == nil && cmd.Exe != nil)
	assert.True(t, parseBytes([]byte("exe v12"), &cmd) == nil && cmd.Exe != nil)

	assert.True(t, parseBytes([]byte("exit"), &cmd) == nil && cmd.Exit != nil)

	assert.Nil(t, parseBytes([]byte("go 1"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, parseBytes([]byte("go 1.1"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, parseBytes([]byte("go 64us"), &cmd))
	assert.NotNil(t, cmd.Go)
	parsedDuration, _ := time.ParseDuration("64us")
	assert.Equal(t, 64*time.Microsecond, parsedDuration)
	assert.Nil(t, parseBytes([]byte("go 5h"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, parseBytes([]byte("go ever"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, parseBytes([]byte("go 100 speed 0.5"), &cmd))
	assert.NotNil(t, cmd.Go)
	assert.Nil(t, parseBytes([]byte("go 100 speed 2"), &cmd))
	assert.NotNil(t, cmd.Go)

	assert.True(t, parseBytes([]byte("help"), &cmd) == nil && cmd.Help != nil)
	assert.True(t, parseBytes([]byte("host add \"myhost.example.com\" \"fc00::1234\" 35683 1717"), &cmd) == nil && cmd.Host != nil)

	assert.True(t, parseBytes([]byte("joins"), &cmd) == nil && cmd.Joins != nil)

	assert.True(t, parseBytes([]byte("log"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, parseBytes([]byte("log debug"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, parseBytes([]byte("log info"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, parseBytes([]byte("log warn"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, parseBytes([]byte("log error"), &cmd) == nil && cmd.LogLevel != nil)
	assert.True(t, parseBytes([]byte("log fatal"), &cmd) != nil && cmd.LogLevel != nil) // not supported.

	assert.True(t, parseBytes([]byte("move 1 200 300"), &cmd) == nil && cmd.Move != nil)

	assert.True(t, parseBytes([]byte("node 1 \"cmd\""), &cmd) == nil && cmd.Node != nil, cmd.Node.Command != nil)
	assert.True(t, parseBytes([]byte("node 1"), &cmd) == nil && cmd.Node != nil && cmd.Node.Command == nil)

	assert.True(t, parseBytes([]byte("nodes"), &cmd) == nil && cmd.Nodes != nil)

	assert.True(t, parseBytes([]byte("partitions"), &cmd) == nil && cmd.Partitions != nil)
	assert.True(t, parseBytes([]byte("pts"), &cmd) == nil && cmd.Partitions != nil)

	assert.True(t, parseBytes([]byte("ping 1 2"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 any"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 mleid"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 13 223 slaac"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 rloc"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 linklocal"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 \"2001::1\""), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 datasize 100"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.DataSize.Val == 100)
	assert.True(t, parseBytes([]byte("ping 1 2 interval 6"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.Interval.Val == 6)
	assert.True(t, parseBytes([]byte("ping 1 2 hoplimit 3"), &cmd) == nil && cmd.Ping != nil && cmd.Ping.HopLimit.Val == 3)
	assert.True(t, parseBytes([]byte("ping 1 2 datasize 20 interval 3 hoplimit 60"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("ping 1 2 datasize 20 hoplimit 60 interval 3"), &cmd) == nil && cmd.Ping != nil)
	assert.True(t, parseBytes([]byte("pings"), &cmd) == nil && cmd.Pings != nil)

	assert.True(t, parseBytes([]byte("plr"), &cmd) == nil && cmd.Plr != nil && cmd.Plr.Val == nil)
	assert.True(t, parseBytes([]byte("plr 1"), &cmd) == nil && cmd.Plr != nil && *cmd.Plr.Val == 1)
	assert.True(t, parseBytes([]byte("plr 0.78910"), &cmd) == nil && cmd.Plr != nil && *cmd.Plr.Val == 0.78910)

	assert.True(t, parseBytes([]byte("radio 1 on"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, parseBytes([]byte("radio 1 off"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, parseBytes([]byte("radio 1 2 3 on"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, parseBytes([]byte("radio 4 5 6 off"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, parseBytes([]byte("radio 4 5 6 ft 10 60"), &cmd) == nil && cmd.Radio != nil)
	assert.True(t, parseBytes([]byte("radiomodel AnyName"), &cmd) == nil && cmd.RadioModel != nil && cmd.RadioModel.Model == "AnyName")
	assert.True(t, parseBytes([]byte("radiomodel 42"), &cmd) == nil && cmd.RadioModel != nil && cmd.RadioModel.Model == "42")
	assert.True(t, parseBytes([]byte("radiomodel"), &cmd) == nil && cmd.RadioModel != nil && cmd.RadioModel.Model == "")
	assert.True(t, parseBytes([]byte("radioparam"), &cmd) == nil && cmd.RadioParam != nil && cmd.RadioParam.Param == "" && cmd.RadioParam.Val == nil)
	assert.True(t, parseBytes([]byte("radioparam name1"), &cmd) == nil && cmd.RadioParam != nil && cmd.RadioParam.Param == "name1" && cmd.RadioParam.Val == nil)
	assert.True(t, parseBytes([]byte("radioparam name1 0.43"), &cmd) == nil && cmd.RadioParam != nil && cmd.RadioParam.Param == "name1" && *cmd.RadioParam.Val == 0.43)
	assert.True(t, parseBytes([]byte("radioparam name2 -23.512"), &cmd) == nil && cmd.RadioParam != nil && cmd.RadioParam.Param == "name2" && cmd.RadioParam.Sign == "-" && *cmd.RadioParam.Val == 23.512)

	assert.True(t, parseBytes([]byte("scan 1"), &cmd) == nil && cmd.Scan != nil)

	assert.True(t, parseBytes([]byte("speed"), &cmd) == nil && cmd.Speed != nil && cmd.Speed.Speed == nil)
	assert.True(t, parseBytes([]byte("speed 1.5"), &cmd) == nil && cmd.Speed != nil && *cmd.Speed.Speed == 1.5)

	assert.True(t, parseBytes([]byte("time"), &cmd) == nil && cmd.Time != nil)

	assert.True(t, parseBytes([]byte("send udp 12 54 mleid datasize 123"), &cmd) == nil && cmd.Send != nil &&
		len(cmd.Send.DstId) == 1 && cmd.Send.DstId[0].Id == 54 && cmd.Send.AddrType.Type == AddrTypeMleid)
	assert.True(t, parseBytes([]byte("send udp 2 13-18 27-39 ds 123"), &cmd) == nil && cmd.Send != nil &&
		len(cmd.Send.DstId) == 2 && cmd.Send.DstId[0].Id == 13 && cmd.Send.DataSize.Val == 123)

	assert.True(t, parseBytes([]byte("watch"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes == nil)
	assert.True(t, parseBytes([]byte("watch all"), &cmd) == nil && cmd.Watch != nil && len(cmd.Watch.Nodes) == 1 &&
		cmd.Watch.Nodes[0].All != nil)
	assert.True(t, parseBytes([]byte("watch 2 5 6"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil)
	assert.True(t, parseBytes([]byte("watch 2 50 75-80 93-95"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil)
	assert.True(t, parseBytes([]byte("watch 1 2 5 6 debug"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil &&
		len(cmd.Watch.Level) == 5)
	assert.True(t, parseBytes([]byte("watch default T"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes == nil &&
		len(cmd.Watch.Level) == 1)
	assert.True(t, parseBytes([]byte("watch 1 2 5 6 I"), &cmd) == nil && cmd.Watch != nil && cmd.Watch.Nodes != nil &&
		len(cmd.Watch.Level) == 1)
	assert.True(t, parseBytes([]byte("unwatch 2 5 6"), &cmd) == nil && cmd.Unwatch != nil)
	assert.True(t, parseBytes([]byte("unwatch all"), &cmd) == nil && cmd.Unwatch != nil)

	assert.True(t, parseBytes([]byte("web"), &cmd) == nil && cmd.Web != nil)
}

func TestContextlessCommandPat(t *testing.T) {
	assert.True(t, isContextlessCommand("exit"))
	assert.True(t, isContextlessCommand("node 1"))
	assert.True(t, isContextlessCommand("!nodes"))
	assert.True(t, isContextlessCommand("!ping 23 24"))
}

func TestBackgroundCommandPat(t *testing.T) {
	// only node-context commands are required to test here. See isBackgroundCommand() why.
	assert.True(t, backgroundCommandsPat.MatchString("scan"))
	assert.True(t, backgroundCommandsPat.MatchString("ping"))
	assert.True(t, backgroundCommandsPat.MatchString("discover"))
	assert.True(t, backgroundCommandsPat.MatchString("dns resolve 1234"))
	assert.True(t, backgroundCommandsPat.MatchString("dns resolve4 1234"))
	assert.True(t, backgroundCommandsPat.MatchString("dns browse example.com"))
	assert.True(t, backgroundCommandsPat.MatchString("dns service test"))
	assert.True(t, backgroundCommandsPat.MatchString("dns servicehost testlabel testname"))
	assert.True(t, backgroundCommandsPat.MatchString("dns compression enable"))
	assert.True(t, backgroundCommandsPat.MatchString("networkdiagnostic get ff02::1234 23 24 25"))
	assert.True(t, backgroundCommandsPat.MatchString("networkdiagnostic reset fd18::1234 2 3"))
	assert.True(t, backgroundCommandsPat.MatchString("networkdiagnostic qry ff02::1234 23 24 25"))
	assert.True(t, backgroundCommandsPat.MatchString("mdns register anystring anystring"))
	assert.True(t, backgroundCommandsPat.MatchString("mdns   register  anystring anystring"))

	assert.False(t, backgroundCommandsPat.MatchString("state"))
	assert.False(t, backgroundCommandsPat.MatchString("coap get ff02::1234 test"))
	assert.True(t, backgroundCommandsPat.MatchString("mdns config anystring"))
}

type mockCliHandler struct {
	expectedCmd string
	handleError error
	handleCount int
	t           *testing.T
}

func (hnd *mockCliHandler) HandleCommand(cmd string, output io.Writer) error {
	assert.Equal(hnd.t, hnd.expectedCmd, cmd)
	hnd.handleCount += 1
	return hnd.handleError
}

func (hnd *mockCliHandler) GetPrompt() string {
	return "> "
}

func TestCliStartStop(t *testing.T) {
	Cli = newCliInstance()
	handler := mockCliHandler{
		expectedCmd: "help",
		handleError: nil,
		t:           t,
	}

	opt := DefaultCliOptions()
	r, w, _ := os.Pipe()
	opt.Stdin = r
	err := make(chan error, 1)
	go func() {
		err <- Cli.Run(&handler, opt)
	}()
	<-Cli.Started
	fmt.Fprint(w, "help\n")
	time.Sleep(time.Millisecond * 500)
	_ = w.Close()
	Cli.Stop()

	assert.Nil(t, <-err)
	assert.Equal(t, 1, handler.handleCount)
}

func TestCliCommandNotDefined(t *testing.T) {
	Cli = newCliInstance()
	handler := mockCliHandler{
		expectedCmd: "xyz",
		handleError: fmt.Errorf("undefined command"),
		t:           t,
	}

	opt := DefaultCliOptions()
	r, w, _ := os.Pipe()
	opt.Stdin = r
	err := make(chan error, 1)
	go func() {
		err <- Cli.Run(&handler, opt)
	}()
	<-Cli.Started
	fmt.Fprint(w, "xyz\n") // unknown command triggers handle-error, which causes CLI exit.

	assert.NotNil(t, <-err)
	assert.Equal(t, 1, handler.handleCount)

	Cli.Stop() // calling Stop() after CLI has already exited.
}

func TestNodeSelectorUniqueSorted(t *testing.T) {
	var inp, outp, exp []NodeSelector

	inp = []NodeSelector{{Id: 3}, {Id: 3}, {Id: 1}, {Id: 2}, {Id: 1234}}
	exp = []NodeSelector{{Id: 1}, {Id: 2}, {Id: 3}, {Id: 1234}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	inp = []NodeSelector{{Id: 1}, {Id: 18}, {Id: 17}, {Id: 18}, {Id: 2}, {Id: 19}, {Id: 1}}
	exp = []NodeSelector{{Id: 1}, {Id: 2}, {Id: 17}, {Id: 18}, {Id: 19}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	inp = []NodeSelector{{Id: 42}}
	exp = []NodeSelector{{Id: 42}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	inp = []NodeSelector{}
	exp = []NodeSelector{}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	inp = []NodeSelector{{Id: 18, IdRange: 40}, {Id: 8}}
	exp = []NodeSelector{{Id: 8}, {Id: 18, IdRange: 40}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	inp = []NodeSelector{{Id: 200, IdRange: 299}, {Id: 18, IdRange: 40}, {Id: 88}}
	exp = []NodeSelector{{Id: 18, IdRange: 40}, {Id: 88}, {Id: 200, IdRange: 299}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)

	allStr := "all"
	inp = []NodeSelector{{Id: 200, IdRange: 299}, {All: &allStr}, {Id: 88}}
	exp = []NodeSelector{{All: &allStr}}
	outp = getUniqueAndSorted(inp)
	assert.Equal(t, exp, outp)
}
