package frame

import "fmt"

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
	return []byte(b.value)
}

// String provides a text representation of an BulkString frame.
func (b *BulkString) String() string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(b.value), b.value)
}

// Value returns the value associated with the simple string.
func (b *BulkString) Value() string {
	return b.value
}
