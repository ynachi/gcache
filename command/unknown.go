package command

import (
	"bufio"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
)

type Unknown struct{}

func (c *Unknown) Apply(_ *db.Cache, _ *bufio.Writer) {
}

func (c *Unknown) FromFrame(_ *frame.Array) error {
	return nil
}

func (c *Unknown) String() string {
	return "unknown command"
}
