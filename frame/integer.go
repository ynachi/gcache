package frame

import (
	"fmt"
	"io"
	"strconv"
)

// Integer implements Framer interface.
type Integer struct {
	value int64
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (i *Integer) Serialize() []byte {
	return []byte(i.String())
}

// String provides a text representation of an Integer frame.
func (i *Integer) String() string {
	return fmt.Sprintf(":%s\r\n", strconv.Itoa(int(i.value)))
}

// WriteTo writes a frame to an io.reader.
func (i *Integer) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := i.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}
