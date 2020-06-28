package main

import (
	"github.com/chzyer/readline"
	"github.com/simonlingoogle/go-simplelogger"
	"io"
	"os"
	"strings"
)

const (
	Prompt = "> "
)

func runCli() error {
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
		Prompt:          Prompt,
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

		os.Stdout.WriteString(line + "\n")

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if line == "exit" {
			os.Exit(0)
		}

		_ = os.Stdout.Sync()
	}

}

func main() {
	err := runCli()
	if err != nil {
		simplelogger.Error(err)
	}
}
