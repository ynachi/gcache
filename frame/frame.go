package frame

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrInvalidSimpleString = errors.New("input contains invalid characters (CR or LF)")
	ErrNotEnoughData       = errors.New("not enough data to decode a valid frame")
	ErrInvalidError        = errors.New("input contains invalid characters (CR or LF)")
	ErrMalformedFrame      = errors.New("unable to decode a valid frame from data")
)

// initialTemporaryBufferSize is the initial size of temporary buffers used in Deserialize methods.
const initialTemporaryBufferSize = 1024

type Framer interface {
	// Serialize returns a slice of bytes representation of this frame. It produces a slice of bytes
	// readies to be transferred other the network.
	Serialize() []byte

	// Deserialize tries to read a frame from a buffer. It returns an error if no frame can be read from the buffer.
	// In case an error occur, the bytes read until the error are discarded. The reason to discard is that the
	//  buffer is reused to read multiple frames. So if there is an invalid frame, discarding it give the chance to
	// read a good one in subsequent calls. Also note that the buffer is supposed to be filled with stream of data.
	// Lastly, after a successful read, the read cursor is positioned after the bytes read for subsequent reads.
	// Care should be taken to check the incoming frame type before calling Deserialize as bytes read upon error
	// are lost.
	Deserialize(b *bytes.Buffer) error

	// String makes a frame implement the Stringer interface. There is no such thing as nil frame.
	// It should be considered to call any methods of this interface on a nil instance a programming error.
	String() string
}

// SimpleString implements Framer interface
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

// Deserialize implements Deserialize method from Framer interface
// This method reads a CRLF terminated slice of bytes from the buffer and turn it
// to a SimpleString Framer type. The read string should not contain any CR or LF
// as defined by the RESP protocol.
func (s *SimpleString) Deserialize(buff *bytes.Buffer) error {
	frameContent, err := simpleStringFromBuffer(buff)
	if err != nil {
		return err
	}
	s.value = frameContent
	return nil
}

// simpleStringFromBuffer tries a simple string from a buffer. It error if it cannot immediately read one.
func simpleStringFromBuffer(buff *bytes.Buffer) (string, error) {
	frameContent, err := readUntilCRLFSimple(buff)
	if err != nil {
		return "", err
	}
	// if there is any CR in the middle of the string,
	// it is not a frame of type simple string. We do not double-check for LF
	// because from the first read, we are guaranteed to not have any LF in the middle
	if strings.ContainsRune(frameContent, '\r') {
		return "", ErrInvalidSimpleString
	}
	return frameContent, nil
}

// readUntilCRLFSimple a string with does not contain any CR or LF until it reach.
// It returns an error if the immediately coming string does not match the requirements.
// The result is stripped from the CRLF. In case of an error, the bytes read are lost.
func readUntilCRLFSimple(buff *bytes.Buffer) (string, error) {
	if buff.Len() < 2 {
		return "", ErrNotEnoughData
	}
	readBytes, err := buff.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if readBytes[len(readBytes)-2] != '\r' {
		return "", ErrInvalidSimpleString
	}
	simpleString := string(readBytes[:len(readBytes)-2])
	return simpleString, nil
}

// Error implements Framer interface
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

// Deserialize implements Deserialize method from Framer interface
// This method reads a CRLF terminated slice of bytes from the buffer and turn it
// to an Error Framer type. The read string should not contain any CR or LF
// as defined by the RESP protocol.
func (e *Error) Deserialize(buff *bytes.Buffer) error {
	frameContent, err := simpleStringFromBuffer(buff)
	if err != nil {
		return err
	}
	e.value = frameContent
	return nil
}

// Integer implements Framer interface
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

// Deserialize reads an Integer frame from a buffer. It error if it cannot get a valid frame
// from the immediate incoming bytes. Bytes reads are lost in case of error. Upon successful
// read, the io read cursor of the buffer is properly set to try to read the next valid frame.
func (i *Integer) Deserialize(buff *bytes.Buffer) error {
	nextString, err := readUntilCRLFSimple(buff)
	if err != nil {
		return err
	}
	frameContent, err := strconv.Atoi(nextString)
	if err != nil {
		return err
	}
	i.value = int64(frameContent)
	return nil
}

// BulkString implements Framer interface
type BulkString struct {
	value string
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (b *BulkString) Serialize() []byte {
	return []byte(b.value)
}

// String provides a text representation of an BulkString frame.
func (b *BulkString) String() string {
	return fmt.Sprintf("$%d\r\n%s\r\n", len(b.value), b.value)
}

// Deserialize reads an Integer frame from a buffer. It error if it cannot get a valid frame
// from the immediate incoming bytes. Bytes reads are lost in case of error. Upon successful
// read, the io read cursor of the buffer is properly set to try to read the next valid frame.
// Unlike a SimpleString, the BulkString can contain CR or LF in the middle.
func (b *BulkString) Deserialize(buff *bytes.Buffer) error {
	if buff.Len() < 2 {
		return ErrNotEnoughData
	}
	// read the size of bulk string before
	sizeString, err := readUntilCRLFSimple(buff)
	if err != nil {
		return err
	}
	size, err := strconv.Atoi(sizeString)
	if err != nil {
		return err
	}
	tmpRead := make([]byte, 0, initialTemporaryBufferSize)
	for {
		bs, err := buff.ReadBytes('\n')
		if err != nil {
			return err
		}
		tmpRead = append(tmpRead, bs...)
		if len(tmpRead) >= 2 && tmpRead[len(tmpRead)-2] == '\r' {
			break
		}
	}
	if len(tmpRead) != size+2 {
		return ErrMalformedFrame
	}
	b.value = string(tmpRead[:size])
	return nil
}

// Bool implements Framer interface.
type Bool struct {
	value bool
}

// Serialize turns the frame into a slice of byte for transfer over a network stream.
func (b *Bool) Serialize() []byte {
	return []byte(b.String())
}

// String provides a text representation of a Bool frame.
func (b *Bool) String() string {
	if b.value == true {
		return "#t\r\n"
	}
	return "#f\r\n"
}

// Deserialize reads a Bool frame from a buffer. It error if it cannot get a valid frame
// from the immediate incoming bytes. Bytes reads are lost in case of error. Upon successful
// read, the io read cursor of the buffer is properly set to try to read the next valid frame.
func (b *Bool) Deserialize(buff *bytes.Buffer) error {
	w, err := simpleStringFromBuffer(buff)
	if err != nil {
		return err
	}
	switch w {
	case "t":
		b.value = true
	case "f":
		b.value = false
	default:
		return ErrMalformedFrame
	}
	return nil
}

//type NullFrame string
//
//func (n NullFrame) Serialize() []uint8 {
//	return []uint8(n)
//}
//
//func (n NullFrame) String() string {
//	return "_\r\n"
//}

// example to decode frames based on the type
//
//
// Define your Framer interface
//type Framer interface {
//	Deserialize(b *bytes.Buffer) error
//	// other methods...
//}
//
//// Suppose you have two different frame types
//type FrameType1 struct { /* fields */ }
//type FrameType2 struct { /* fields */ }
//
//// You define Deserialize differently for each type
//func (f *FrameType1) Deserialize(b *bytes.Buffer) error {
//	// deserialization for FrameType1
//	// ...
//	return nil
//}
//
//func (f *FrameType2) Deserialize(b *bytes.Buffer) error {
//	// deserialization for FrameType2
//	// ...
//	return nil
//}
//
//// Then, in your function dealing with the incoming data
//switch frameTypeByte {
//case byte1:
//frame := &FrameType1{}
//err := frame.Deserialize(b)
//// handle error and use the frame
//case byte2:
//frame := &FrameType2{}
//err := frame.Deserialize(b)
//// handle error and use the frame
//default:
//// handle error: unrecognized frame type
//}
