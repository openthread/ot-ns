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
	"errors"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"

	"github.com/openthread/ot-ns/logger"
)

type CliHandler interface {
	HandleCommand(cmd string, output io.Writer) error
	GetPrompt() string
}

type CliOptions struct {
	EchoInput bool
	Stdin     *os.File
	Stdout    *os.File
}

func DefaultCliOptions() *CliOptions {
	return &CliOptions{
		EchoInput: false,
		Stdin:     nil,
		Stdout:    nil,
	}
}

// CliInstance is the singleton CLI instance
type CliInstance struct {
	Started          chan struct{}
	Options          *CliOptions
	readlineInstance *readline.Instance
	waitCliClosed    chan struct{}
}

var Cli = newCliInstance()

func (cli *CliInstance) RestorePrompt() {
	if cli.readlineInstance != nil {
		cli.readlineInstance.Refresh()
	}
}

func newCliInstance() *CliInstance {
	return &CliInstance{
		Started:       make(chan struct{}),
		waitCliClosed: make(chan struct{}),
	}
}

func getCliOptions(options *CliOptions) *CliOptions {
	if options == nil {
		options = DefaultCliOptions()
	}
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}

	return options
}

func (cli *CliInstance) Stop() {
	<-cli.Started
	// cannot call readlineInstance.Close() from here, as it can block (RunCli() will call it)
	// (https://github.com/chzyer/readline/issues/217)
	// send ETX(Ctrl-C, 0x03, readline.CharInterrupt) to avoid readline internally blocking on Runes() select.
	_, _ = cli.Options.Stdin.WriteString("\003\n")
	_ = cli.Options.Stdin.Close() // trigger RunCli() readline call to stop
	logger.Tracef("Waiting for CLI to stop ...")
	<-cli.waitCliClosed
	logger.Tracef("CLI wait-for-stop done.")
}

func (cli *CliInstance) Run(handler CliHandler, options *CliOptions) error {
	defer logger.Debugf("CLI exit.")
	defer close(cli.waitCliClosed)

	options = getCliOptions(options)
	cli.Options = options

	stdin := options.Stdin
	stdinIsTerminal := readline.IsTerminal(int(stdin.Fd()))
	if stdinIsTerminal {
		stdinState, err := readline.GetState(int(stdin.Fd()))
		if err != nil {
			close(cli.Started)
			return err
		}
		defer func() {
			_ = readline.Restore(int(stdin.Fd()), stdinState)
		}()
	}

	stdout := options.Stdout
	stdoutIsTerminal := readline.IsTerminal(int(stdout.Fd()))
	if stdoutIsTerminal {
		stdoutState, err := readline.GetState(int(stdout.Fd()))
		if err != nil {
			close(cli.Started)
			return err
		}
		defer func() {
			_ = readline.Restore(int(stdout.Fd()), stdoutState)
		}()
	}

	readlineConfig := &readline.Config{
		Prompt:          handler.GetPrompt(),
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
		FuncFilterInputRune: func(r rune) (rune, bool) {
			switch r {
			// block CtrlZ feature
			case readline.CharCtrlZ:
				return r, false
			}
			return r, true
		},
	}

	if options.Stdin != nil {
		readlineConfig.Stdin = options.Stdin
	}

	if options.Stdout != nil {
		readlineConfig.Stdout = options.Stdout
	}

	l, err := readline.NewEx(readlineConfig)

	if err != nil {
		close(cli.Started)
		return err
	}

	defer func() {
		_ = l.Close()
	}()
	cli.readlineInstance = l
	close(cli.Started)

	for {
		// update the prompt and read a line
		l.SetPrompt(handler.GetPrompt())
		line, err := l.Readline()

		if len(line) > 0 && line[0] == readline.CharInterrupt {
			return nil
		} else if errors.Is(err, readline.ErrInterrupt) {
			if len(line) == 0 {
				return nil
			} else {
				continue // Ctrl-C in midline edit only cancels the present cmd line.
			}
		} else if err == io.EOF { // typical way to close the CLI
			return nil
		} else if err != nil {
			return err
		}

		if options.EchoInput {
			if _, err := stdout.WriteString(line + "\n"); err != nil {
				_ = stdout.Sync()
				return err
			}
		}

		cmd := strings.TrimSpace(line)
		if len(cmd) == 0 {
			stdout.WriteString("")
			_ = stdout.Sync()
			continue
		}

		if err = handler.HandleCommand(cmd, l.Stdout()); err != nil {
			_ = stdout.Sync()
			return err
		}

		_ = stdout.Sync()
	}
}

// OnStdout is the handler called when new Stdout/Stderr output occurred.
func (cli *CliInstance) OnStdout() {
	cli.RestorePrompt()
}
