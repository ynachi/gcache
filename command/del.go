package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"log/slog"
)

type Del struct {
	keys   []string
	logger *slog.Logger
}

func (c *Del) Apply(cache *db.Cache, dest *bufio.Writer) {
	defer func(dest *bufio.Writer) {
		err := dest.Flush()
		if err != nil {
			c.logger.Error("unable to write to destination", "error", err)
		}
	}(dest)
	numKeys := cache.Delete(c.keys...)
	resp := frame.NewInteger(int64(numKeys))
	_, err := resp.WriteTo(dest)
	if err != nil {
		c.logger.Error("failed to write response", "error", err)
	}
}

func (c *Del) FromFrame(f *frame.Array) error {
	if f.Size() < 2 {
		return gerror.ErrInvalidCmdArgs
	}
	for i := 1; i < f.Size(); i++ {
		c.keys = append(c.keys, f.Get(i).(*frame.BulkString).Value())
	}
	return nil
}

func (c *Del) Name() string {
	return "del"
}
