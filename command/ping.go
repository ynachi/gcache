package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"log/slog"
)

type Ping struct {
	message string
	logger  *slog.Logger
}

func (c *Ping) Apply(_ *db.Cache, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {
			c.logger.Error("unable to write to destination", "error", err)
		}
	}(dest)
	if c.message == "PONG" {
		if resp, err := frame.NewSimpleString(c.message); err == nil {
			_, err = resp.WriteTo(dest)
			if err != nil {
				// handle error
				return
			}
		}
	}
	resp := frame.NewBulkString(c.message)
	_, err := resp.WriteTo(dest)
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
	case cmdName != "ping" || f.Size() > 2:
		return gerror.ErrInvalidPingCommand
	case f.Size() == 1:
		c.message = "PONG"
	default:
		// GetCmdName already check a frame type
		s := f.Get(1).(*frame.BulkString)
		c.message = s.Value()
	}
	return nil
}

func (c *Ping) SetMessage(message string) {
	c.message = message
}

func (c *Ping) Name() string {
	return "ping"
}
