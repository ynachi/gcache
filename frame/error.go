package frame

import (
	"fmt"
	"io"
	"strings"
)

// Error implements Framer interface.
type Error struct {
	value string
}

// NewError creates a new Error frame. It's make sure the frame is valid upon creation.
func NewError(s string) (*Error, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, ErrInvalidError
	}
	return &Error{s}, nil
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (e *Error) Serialize() []byte {
	return []byte(e.String())
}

// String provides a text representation of an Error frame.
func (e *Error) String() string {
	return fmt.Sprintf("-%s\r\n", e.value)
}

// WriteTo writes a frame to an io.reader.
func (e *Error) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := e.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}
