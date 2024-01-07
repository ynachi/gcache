package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"log/slog"
)

type Set struct {
	key    string
	value  string
	logger *slog.Logger
}

func (c *Set) Apply(cache *db.Cache, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {
			c.logger.Error("unable to write to destination", "error", err)
		}
	}(dest)
	cache.Set(c.key, c.value)
	resp, _ := frame.NewSimpleString("ok")
	_, err := resp.WriteTo(dest)
	if err != nil {
		c.logger.Error("failed to write response", "error", err)
	}
}

// @TODO Minimal implementation for now, to fix.
func (c *Set) FromFrame(f *frame.Array) error {
	if f.Size() != 3 {
		return gerror.ErrInvalidCmdArgs
	}
	c.key = f.Get(1).(*frame.BulkString).Value()
	c.value = f.Get(2).(*frame.BulkString).Value()

	return nil
}

func (c *Set) Name() string {
	return "set"
}
