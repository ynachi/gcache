package commands

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"log/slog"
)

type Ping struct {
	message string
	logger  *slog.Logger
}

func (c *Ping) Apply(_ db.Database, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {
			c.logger.Error("unable to write to destination", "error", err)
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
	cmdName, err := GetCmdName(f)
	if err != nil {
		return err
	}
	switch {
	case cmdName != "PING" || f.Size() > 2:
		return ErrInvalidPingCommand
	case f.Size() == 1:
		c.message = "PONG"
	default:
		// GetCmdName already check frame type
		s := f.Get(1).(*frame.BulkString)
		c.message = s.Value()
	}
	return nil
}
