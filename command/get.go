package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"log/slog"
)

type Get struct {
	key    string
	logger *slog.Logger
}

func (c *Get) Apply(cache *db.Cache, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {
			c.logger.Error("unable to write to destination", "error", err)
		}
	}(dest)
	value, ok := cache.Get(c.key)
	if !ok {
		resp := frame.Null{}
		_, err := resp.WriteTo(dest)
		if err != nil {
			c.logger.Error("failed to write response", "error", err)
		}
		return
	}
	resp := frame.NewBulkString(value)
	_, err := resp.WriteTo(dest)
	if err != nil {
		c.logger.Error("failed to write response", "error", err)
	}
}

func (c *Get) FromFrame(f *frame.Array) error {
	if f.Size() != 2 {
		return gerror.ErrInvalidCmdArgs
	}
	c.key = f.Get(1).(*frame.BulkString).Value()

	return nil
}

func (c *Get) Name() string {
	return "get"
}
