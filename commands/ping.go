package commands

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
)

type Ping struct {
	message string
}

func (c *Ping) GetMessage() string {
	return c.message
}

func (c *Ping) Apply(_ db.Database, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {

		}
	}(dest)
	resp, err := frame.NewSimpleString(c.message)
	if err != nil {
		// handle error
		return
	}
	_, err = resp.WriteTo(dest)
	if err != nil {
		// handle error
		return
	}
}

func (c *Ping) FromFrame(f *frame.Array) error {
	switch {
	case f.Size() < 1 && f.Size() > 2:
		return ErrInvalidPingCommand
	case f.Size() == 1:
		c.message = "PONG"
	default:
		s, ok := f.Get(1).(*frame.BulkString)
		if !ok {
			return ErrNotGcacheCmd
		}
		c.message = s.Value()
	}
	return nil
}
