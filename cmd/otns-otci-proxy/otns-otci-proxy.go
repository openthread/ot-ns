package main

import (
	"io"
	"os"

	"github.com/pkg/errors"

	"github.com/openthread/ot-ns/cli/runcli"
	"github.com/simonlingoogle/go-simplelogger"
)

type cliHandler struct{}

func (c cliHandler) GetPrompt() string {
	return "> "
}

func (c cliHandler) HandleCommand(cmd string, output io.Writer) error {
	if _, err := output.Write([]byte("Done\n")); err != nil {
		return err
	}

	if cmd == "exit" {
		os.Exit(0)
	} else {
		panic(errors.Errorf("can not handle other command"))
	}

	return nil
}

func main() {
	err := runcli.RunCli(&cliHandler{}, &runcli.CliOptions{
		EchoInput: true,
	})

	if err != nil {
		simplelogger.Error(err)
	}
}
