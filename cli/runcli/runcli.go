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

package runcli

import (
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
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

var (
	readlineInstance *readline.Instance
)

func RestorePrompt() {
	readlineInstance.Refresh()
}

func RunCli(handler CliHandler, options *CliOptions) error {
	if options == nil {
		options = DefaultCliOptions()
	}

	stdin := options.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}

	stdout := options.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}

	stdinIsTerminal := readline.IsTerminal(int(stdin.Fd()))
	if stdinIsTerminal {
		stdinState, err := readline.GetState(int(stdin.Fd()))
		if err != nil {
			return err
		}

		defer func() {
			_ = readline.Restore(int(stdin.Fd()), stdinState)
		}()
	}

	stdoutIsTerminal := readline.IsTerminal(int(stdout.Fd()))
	if stdoutIsTerminal {
		stdoutState, err := readline.GetState(int(stdout.Fd()))
		if err != nil {
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
	readlineInstance = l

	if err != nil {
		return err
	}
	defer func() {
		_ = l.Close()
	}()

	for {
		// update the prompt
		l.SetPrompt(handler.GetPrompt())

		line, err := l.Readline()

		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				return nil
			} else {
				continue
			}
		} else if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		if options.EchoInput {
			if _, err := stdout.WriteString(line + "\n"); err != nil {
				return err
			}
		}

		cmd := strings.TrimSpace(line)
		if len(cmd) == 0 {
			continue
		}

		if err = handler.HandleCommand(cmd, l.Stdout()); err != nil {
			return err
		}

		_ = stdout.Sync()
	}
}
