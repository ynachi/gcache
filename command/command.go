package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"strings"
)

// Command represents a command issued to the cache server with their args.
type Command interface {
	// Name returns the command name
	Name() string

	// Apply applies the command et write back the response to the client
	Apply(db *db.Cache, dest *bufio.Writer)

	// FromFrame form the command from a Frame
	FromFrame(f *frame.Array) error
}

// NewCommand instantiates a concrete command type base on its name.
// NewCommand should rely on
// GetCmdName to extract the command name from an Array frame in most cases.
// This should avoid returning a nil command struct.
func NewCommand(cmdName string) Command {
	switch cmdName {
	case "ping":
		return new(Ping)
	case "set":
		return new(Set)
	case "get":
		return new(Get)
	default:
		return nil
	}
}

// GetCmdName gets a command name from a Frame Array.
func GetCmdName(f *frame.Array) (string, error) {
	if f.Size() < 1 {
		return "", gerror.ErrNotGcacheCmd
	}
	cmdNameFrame := f.Get(0)
	cmdName, ok := cmdNameFrame.(*frame.BulkString)
	if !ok {
		return "", gerror.ErrNotGcacheCmd
	}
	return strings.ToLower(cmdName.Value()), nil
}
