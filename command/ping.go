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
			c.logger.Error("unable to write to destination", "gerror", err)
		}
	}(dest)
	resp, err := frame.NewSimpleString(c.message)
	if err != nil {
		// handle gerror
		return
	}
	_, err = resp.WriteTo(dest)
	if err != nil {
		// handle gerror
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

func (c *Ping) String() string {
	bs := frame.NewBulkString(c.message)
	return bs.String()
}

func (c *Ping) SetMessage(message string) {
	c.message = message
}
