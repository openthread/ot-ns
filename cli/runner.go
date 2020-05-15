// Copyright (c) 2020, The OTNS Authors.
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

// Package cli implements the OTNS-CLI. It parses and executes CLI commands.
package cli

import (
	"fmt"
	"github.com/openthread/ot-ns/progctx"
	"github.com/openthread/ot-ns/simulation"
	. "github.com/openthread/ot-ns/types"
	"github.com/pkg/errors"
	"github.com/simonlingoogle/go-simplelogger"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/chzyer/readline"
)

const (
	prompt = "> " // the default CLI prompt
)

var (
	contextNodeId          = InvalidNodeId
	readlineInstance       *readline.Instance
	contextLessCommandsPat = regexp.MustCompile(`(exit|node)\b`)
)

func enterNodeContext(nodeid NodeId) bool {
	simplelogger.AssertTrue(nodeid == InvalidNodeId || nodeid > 0)
	if contextNodeId == nodeid {
		return false
	}

	contextNodeId = nodeid
	if nodeid == InvalidNodeId {
		readlineInstance.SetPrompt(prompt)
	} else {
		readlineInstance.SetPrompt(fmt.Sprintf("node %d%s", contextNodeId, prompt))
	}
	return true
}

// Run runs the CLI console.
func Run(ctx *progctx.ProgCtx, sim *simulation.Simulation) {
	var err error
	defer func() {
		ctx.Cancel(errors.Wrapf(err, "console exit"))
	}()

	ctx.WaitAdd("cli", 1)
	defer ctx.WaitDone("cli")

	err = run(ctx, sim)
}

func run(ctx *progctx.ProgCtx, sim *simulation.Simulation) error {
	cr := newCmdRunner(ctx, sim)

	stdinFd := int(os.Stdin.Fd())
	stdinIsTerminal := readline.IsTerminal(stdinFd)
	if stdinIsTerminal {
		stdinState, err := readline.GetState(stdinFd)
		if err != nil {
			return err
		}

		defer func() {
			_ = readline.Restore(stdinFd, stdinState)
		}()
	}

	stdoutFd := int(os.Stdout.Fd())
	stdoutIsTerminal := readline.IsTerminal(stdoutFd)
	if stdoutIsTerminal {
		stdoutState, err := readline.GetState(stdoutFd)
		if err != nil {
			return err
		}
		defer func() {
			_ = readline.Restore(stdoutFd, stdoutState)
		}()
	}

	l, err := readline.NewEx(&readline.Config{
		Prompt:          prompt,
		HistoryFile:     "/tmp/otns-cmds.tmp",
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
	})

	if err != nil {
		return err
	}
	defer func() {
		_ = l.Close()
	}()
	readlineInstance = l

	for {
		line, err := l.Readline()

		if cr.ctx.Err() != nil {
			// program exited, quit console too
			return nil
		}

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

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if contextNodeId != InvalidNodeId && !isContextlessCommand(line) {
			// run the command in node context
			cmd := &command{
				Node: &NodeCmd{
					Node:    NodeSelector{Id: contextNodeId},
					Command: &line,
				},
			}
			cc := cr.Execute(cmd)

			if cc.Err() != nil {
				if _, err := fmt.Fprintf(os.Stdout, "Error: %v\n", cc.Err()); err != nil {
					return err
				}
			} else {
				if _, err := fmt.Fprintf(os.Stdout, "Done\n"); err != nil {
					return err
				}
			}
		} else {
			// run the OTNS-CLI command
			cmd := &command{}
			if err := parseCmdBytes([]byte(line), cmd); err != nil {
				if _, err := fmt.Fprintf(os.Stdout, "Error: %v\n", err); err != nil {
					return err
				}
			} else {
				cc := cr.Execute(cmd)

				if cc.Err() != nil {
					if _, err := fmt.Fprintf(os.Stdout, "Error: %v\n", cc.Err()); err != nil {
						return err
					}
				} else {
					if _, err := fmt.Fprintf(os.Stdout, "Done\n"); err != nil {
						return err
					}
				}
			}
		}

		_ = os.Stdout.Sync()
	}
}

func isContextlessCommand(line string) bool {
	return contextLessCommandsPat.MatchString(line)
}
