package server

import (
	"bufio"
	"github.com/ynachi/gcache/command"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"net"
)

// Connection is a helper struct that helps propagates embedded the treader and writer of a connection while
// allowing top propagates information about this connection.
// Connection needs a reference to the Cache database to operate on it.
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	storage  *db.Cache
	clientIP string
}

// MakeConnection creates a connection from a net.Conn object.
func MakeConnection(c net.Conn, storage *db.Cache) Connection {
	return Connection{
		reader:   bufio.NewReader(c),
		writer:   bufio.NewWriter(c),
		clientIP: c.RemoteAddr().String(),
		conn:     c,
		storage:  storage,
	}
}

func (c Connection) Close() error {
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return c.conn.Close()
}

// GetCommand handles a command received by the server over an established connection.
func (c Connection) GetCommand() (command.Command, error) {
	cmdFrame, err := c.readCmdFrame()
	if err != nil {
		return nil, err
	}
	return parseCommandFromFrame(cmdFrame)
}

// readCmdFrame reads an array frame from the connection. RESP commands are all represented as Array of frames.
func (c Connection) readCmdFrame() (*frame.Array, error) {
	cmdFrame, err := frame.Decode(c.reader)
	if err != nil {
		return nil, err
	}
	arrayFrame, ok := cmdFrame.(*frame.Array)
	if !ok {
		return nil, gerror.ErrNotAGcacheCommand
	}
	return arrayFrame, nil
}

// parseCommandFromFrame extracts a command from a frame array.
func parseCommandFromFrame(f *frame.Array) (command.Command, error) {
	cmdName, err := command.GetCmdName(f)
	if err != nil {
		return nil, err
	}
	cmd := command.NewCommand(cmdName)
	if cmd == nil {
		return nil, gerror.ErrInvalidCmdName
	}
	err = cmd.FromFrame(f)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}
