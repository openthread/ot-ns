package simulation

import (
	"github.com/openthread/ot-ns/types"
	"io"
)

type CmdRunner interface {
	RunCommand(cmd string, output io.Writer) error

	// gets the user's current selected node ID context for running commands, or
	// types.InvalidNodeId for no node context selected.
	GetContextNodeId() types.NodeId
}
