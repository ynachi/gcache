package server

import (
	"bufio"
	"net"
)

// Connection is a helper struct that helps propagates embedded the treader and writer of a connection while
// allowing top propagates information about this connection.
type Connection struct {
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
	clientIP string
}

// MakeConnection creates a connection from a net.Conn object.
func MakeConnection(c net.Conn) Connection {
	return Connection{
		reader:   bufio.NewReader(c),
		writer:   bufio.NewWriter(c),
		clientIP: c.RemoteAddr().String(),
		conn:     c,
	}
}

func (c Connection) Close() error {
	if err := c.writer.Flush(); err != nil {
		return err
	}
	return c.conn.Close()
}
