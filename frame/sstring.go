package frame

import (
	"fmt"
	"io"
	"strings"
)

// SimpleString implements Framer interface.
type SimpleString struct {
	value string
}

// NewSimpleString instantiates a new Simple String frame. This function is preferred to create a SimpleString frame
// because it validates the frame upon creation.
func NewSimpleString(s string) (*SimpleString, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, ErrInvalidSimpleString
	}
	return &SimpleString{s}, nil
}

// Serialize turns a SimpleString into a slice of bytes for transfer over a network stream.
func (s *SimpleString) Serialize() []byte {
	return []byte(s.String())
}

// String provides a text representation of a SimpleString frame.
func (s *SimpleString) String() string {
	return fmt.Sprintf("+%s\r\n", s.value)
}

// WriteTo writes a frame to an io.reader.
func (s *SimpleString) WriteTo(w io.Writer) (int64, error) {
	frameToBytes := s.Serialize()
	count, err := w.Write(frameToBytes)
	return int64(count), err
}

// Value returns the value associated with the simple string.
func (s *SimpleString) Value() string {
	return s.value
}
