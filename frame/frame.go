package frame

import (
	"bufio"
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
	ErrArrayIsFull         = errors.New("array reached its maximum capacity")
)

// initialTemporaryBufferSize is the initial size of temporary buffers used in Deserialize methods.
const initialTemporaryBufferSize = 1024

type Framer interface {
	// Serialize returns a slice of bytes representation of this frame. It produces a slice of bytes
	// readies to be transferred other the network.
	Serialize() []byte

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

// Null frame
type Null struct{}

func (n *Null) Serialize() []byte {
	return []byte(n.String())
}

// String provides a text representation of a Null frame.
func (n *Null) String() string {
	return "_\r\n"
}

// Array represents an array of Frames, which can be mixed types.
// It looks redundant to set the size as it could be deduced from the length
// of the slice this is needed. For example, to decode an Array from a buffer, we
// will need to know its length in advance.
type Array struct {
	size  int
	value []Framer
}

// NewArray create a new Array. This array is not reallocated (means its length is 0) but the
// capacity is set to avoid sliding. The reason the initial len is 0 is that we expect it to be
// filled via append.
func NewArray(size int) *Array {
	value := make([]Framer, 0, size)
	return &Array{size: size, value: value}
}

// Append add a new frame to an Array. It error when there is not enough capacity to add more.
// It does not grow the Array automatically.
func (a *Array) Append(f Framer) error {
	if len(a.value) >= a.size {
		return ErrArrayIsFull
	}
	a.value = append(a.value, f)
	return nil
}

// String provides a text representation of an  Array frame.
func (a *Array) String() string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(a.value)))
	for _, f := range a.value {
		sb.WriteString(f.String())
	}
	return sb.String()
}

func (a *Array) Serialize() []byte {
	return []byte(a.String())
}

// Decode tries to read a frame from a buffer. It returns an error if no frame
// can be read from the buffer. In case an error occur, the bytes read until the
// error are discarded. The reason to discard is that the buffer is reused to
// read multiple frames. So if there is an invalid frame, discarding it give the
// chance to read a good one in subsequent calls. Also note that the buffer is
// supposed to be filled with stream of data. Lastly, after a successful read,
// the read cursor is positioned after the bytes read for subsequent reads. Care
// should be taken to check the incoming frame type before calling Deserialize as
// bytes read upon error are lost. Decode relies on some decoding function define
// at frames level. For instance, when decode identify that it need to decode a simple
// string Frame, it would call DecodeSimpleString.
func Decode(rd *bufio.Reader) (Framer, error) {
	frameID, err := rd.ReadByte()
	if err != nil {
		return nil, err
	}
	switch frameID {
	case '+':
		return DecodeSimpleString(rd)
	case '-':
		return DecodeError(rd)
	case ':':
		return DecodeInteger(rd)
	case '$':
		return DecodeBulkString(rd)
	case '#':
		return DecodeBool(rd)
	case '_':
		return DecodeNull(rd)
	case '*':
		return DecodeArray(rd)
	default:
		return nil, fmt.Errorf("unknown frameID: %v", frameID)
	}
}

// DecodeSimpleString consume a simple string from a buffer. If an
// error occur, it is returned and the bytes read so  far are lost.
func DecodeSimpleString(buff *bufio.Reader) (*SimpleString, error) {
	frameContent, err := simpleStringFromBuffer(buff)
	if err != nil {
		return nil, err
	}
	s, err := NewSimpleString(frameContent)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// simpleStringFromBuffer tries a simple string from a buffer. It error if it cannot immediately read one.
func simpleStringFromBuffer(rd *bufio.Reader) (string, error) {
	frameContent, err := readUntilCRLFSimple(rd)
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
// The result is stripped from the CRLF. This function return various errors which should
// be taken care of by the caller.
// io.EOF when we reach the end of the stream without any CRLF
// ErrNotEnoughData when there is less than 2 digits
// ErrInvalidSimpleString when we encounter LF in the middle. In fact, simple string should
// not contain any LF in the middle in our protocol.
func readUntilCRLFSimple(rd *bufio.Reader) (string, error) {
	readBytes, err := rd.ReadBytes('\n')
	if err != nil {
		return "", err
	}
	if len(readBytes) < 2 {
		return "", ErrNotEnoughData
	}
	if readBytes[len(readBytes)-2] != '\r' {
		return "", ErrInvalidSimpleString
	}
	simpleString := string(readBytes[:len(readBytes)-2])
	return simpleString, nil
}

// DecodeError decode an Error from a buffer
func DecodeError(rd *bufio.Reader) (*Error, error) {
	frameContent, err := simpleStringFromBuffer(rd)
	if err != nil {
		return nil, err
	}
	e, err := NewError(frameContent)
	if err != nil {
		return nil, err
	}
	return e, nil
}

// DecodeInteger decodes an int from a buffer
func DecodeInteger(rd *bufio.Reader) (*Integer, error) {
	frameContent, err := getInt(rd)
	if err != nil {
		return nil, err
	}
	i := Integer{int64(frameContent)}
	return &i, nil
}

// getInt read an int from a buffer
func getInt(rd *bufio.Reader) (int, error) {
	nextString, err := readUntilCRLFSimple(rd)
	if err != nil {
		return 0, err
	}
	ans, err := strconv.Atoi(nextString)
	if err != nil {
		return 0, err
	}
	return ans, nil
}

// DecodeBulkString decodes a bulk string from a buffer.
func DecodeBulkString(rd *bufio.Reader) (*BulkString, error) {
	// read the size of bulk string before
	sizeString, err := readUntilCRLFSimple(rd)
	if err != nil {
		return nil, err
	}
	size, err := strconv.Atoi(sizeString)
	if err != nil {
		return nil, err
	}
	tmpRead := make([]byte, 0, initialTemporaryBufferSize)
	for {
		bs, err := rd.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
		tmpRead = append(tmpRead, bs...)
		if len(tmpRead) >= 2 && tmpRead[len(tmpRead)-2] == '\r' {
			break
		}
	}
	// we already know the size, so let's compare
	if len(tmpRead) != size+2 {
		return nil, ErrMalformedFrame
	}
	bs := BulkString{string(tmpRead[:size])}
	return &bs, nil
}

// DecodeBool decodes a bool from a buffer.
func DecodeBool(rd *bufio.Reader) (*Bool, error) {
	w, err := simpleStringFromBuffer(rd)
	if err != nil {
		return nil, err
	}
	b := Bool{}
	switch w {
	case "t":
		b.value = true
	case "f":
		b.value = false
	default:
		return nil, ErrMalformedFrame
	}
	return &b, nil
}

// DecodeNull decodes a null frame from a buffer.
func DecodeNull(rd *bufio.Reader) (*Null, error) {
	w, err := simpleStringFromBuffer(rd)
	if err != nil {
		return nil, err
	}
	if w != "" {
		return nil, ErrMalformedFrame
	}
	return &Null{}, nil
}

func DecodeArray(rd *bufio.Reader) (*Array, error) {
	length, err := getInt(rd)
	if err != nil {
		return nil, err
	}
	array := NewArray(length)
	for i := 0; i < length; i++ {
		frame, err := Decode(rd)
		if err != nil {
			return nil, err
		}
		if err := array.Append(frame); err != nil {
			return nil, err
		}
	}
	return array, nil
}

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
