package simulation

import "io"

type CmdRunner interface {
	RunCommand(cmd string, output io.Writer) error
}
