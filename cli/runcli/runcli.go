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
}

func RunCli(handler CliHandler, options CliOptions) error {
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
	})

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
			if _, err := os.Stdout.WriteString(line + "\n"); err != nil {
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

		_ = os.Stdout.Sync()
	}
}
