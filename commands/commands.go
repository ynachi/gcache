package commands

import (
	"bufio"
	"errors"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
)

// Add all the command errors here
var (
	ErrInvalidPingCommand = errors.New("ping command is malformed")
	ErrNotGcacheCmd       = errors.New("this frame is not a gcache command")

	ErrInvalidCmdName = errors.New("command not found")
)

// Command represents a command issued to the cache server with their args
type Command interface {
	// Apply applies the command et write back the response to the client
	Apply(db db.Database, dest *bufio.Writer)

	// FromFrame form the commands from a Frame
	FromFrame(f *frame.Array) error
}

var RegisteredCommandName = map[string]struct{}{
	"PING": {},
	"SET":  {},
}

// NewCommand instantiates a concrete command type base on its name. NewCommand should rely on
// GetCmdName to extract the command name from an Array frame in most case. This sould avoid to
// return a nil command struct.
func NewCommand(cmdName string) Command {
	switch cmdName {
	case "PING":
		return new(Ping)
	default:
		return nil
	}
}

// GetCmdName gets a command name from a Frame Array
func GetCmdName(f *frame.Array) (string, error) {
	if f.Size() < 1 {
		return "", ErrNotGcacheCmd
	}
	cmdNameFrame := f.Get(0)
	cmdName, ok := cmdNameFrame.(*frame.BulkString)
	if !ok {
		return "", ErrNotGcacheCmd
	}
	if _, ok := RegisteredCommandName[cmdName.Value()]; !ok {
		return "", ErrInvalidCmdName
	}
	return cmdName.Value(), nil
}
