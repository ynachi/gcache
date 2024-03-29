package frame

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidSimpleString = errors.New("input contains invalid characters (CR or LF)")
	ErrNotEnoughData       = errors.New("not enough data to decode a valid frame")
	ErrInvalidError        = errors.New("input contains invalid characters (CR or LF)")
	ErrMalformedFrame      = errors.New("unable to decode a valid frame from data")
	ErrArrayIsFull         = errors.New("array reached its maximum capacity")
	ErrUnknownFrameType    = errors.New("unknown frame type")
)

// initialTemporaryBufferSize is the initial size of temporary buffers used in Deserialize methods.
const initialTemporaryBufferSize = 1024

type Framer interface {
	// Serialize returns a slice of bytes' representation of this frame.
	// It produces a slice of bytes that is ready to be transferred other the network.
	Serialize() []byte

	// Stringer needs to be implemented to provide a string representation of a frame.
	fmt.Stringer

	// WriterTo WriteTo writes the frames as bytes to an io.Writer
	io.WriterTo
}

// Decode tries to read a frame from a buffer. It returns an error if no frame
// can be read from the buffer.
// In case an error occurs, the bytes read before getting the
// error are discarded. The reason to discard is that the buffer is reused to
// read multiple frames.
// So if there is an invalid frame, discarding it gives the
// chance to read a good one in subsequent calls. Also note that the buffer is
// supposed to be filled with stream of data. Lastly, after a successful read,
// the read cursor is positioned after the bytes read for subsequent reads. Care
// should be taken to check the incoming frame type before calling Deserialize as
// bytes read upon error are lost.
// Decode relies on some decoding function defined at frames' level. For instance, when decode identify that it
// needs to decode a simple string Frame, it would call DecodeSimpleString.
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
		return nil, ErrUnknownFrameType
	}
}

// DecodeSimpleString consume a simple string from a buffer. If an
// gerror occurs, it is returned, and the bytes read so far are lost.
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

// simpleStringFromBuffer tries a simple string from a buffer. It gerror if it cannot immediately read one.
func simpleStringFromBuffer(rd *bufio.Reader) (string, error) {
	frameContent, err := readUntilCRLFSimple(rd)
	if err != nil {
		return "", err
	}
	// If there is any CR in the middle of the string,
	// it is not a frame of type simple string. We do not double-check for LF
	// because from the first read, we are guaranteed to not have any LF in the middle
	if strings.ContainsRune(frameContent, '\r') {
		return "", ErrInvalidSimpleString
	}
	return frameContent, nil
}

// readUntilCRLFSimple a string with does not contain any CR or LF until it reach.
// It returns an error if the immediately coming string does not match the requirements.
// The result is stripped from the CRLF.
// This function returns various errors which should be taken care of by the caller.
// Io.EOF when we reach the end of the stream without any CRLF
// ErrNotEnoughData when there are less than two digits
// ErrInvalidSimpleString when we encounter LF in the middle.
// In fact, simple string should not contain any LF in the middle of our protocol.
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

// DecodeError decode an Error from a buffer.
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

// DecodeInteger decodes an int from a buffer.
func DecodeInteger(rd *bufio.Reader) (*Integer, error) {
	frameContent, err := getInt(rd)
	if err != nil {
		return nil, err
	}
	i := Integer{int64(frameContent)}
	return &i, nil
}

// getInt read an int from a buffer.
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
