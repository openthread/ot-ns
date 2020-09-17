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
