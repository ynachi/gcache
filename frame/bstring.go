package frame

import (
	"fmt"
	"io"
)

// BulkString implements Framer interface.
type BulkString struct {
	value string
}

// NewBulkString returns a new bulk string.
func NewBulkString(value string) *BulkString {
	return &BulkString{value: value}
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (b *BulkString) Serialize() []byte {
	return []byte(b.String())
}

// String provides a text representation of an BulkString frame.
func (b *BulkString) String() string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(b.value), b.value)
}

// Value returns the value associated with the simple string.
func (b *BulkString) Value() string {
	return b.value
}

func (b *BulkString) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := b.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}
