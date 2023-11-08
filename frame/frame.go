package frame

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidSimpleString = errors.New("input contains invalid characters (CR or LF)")
	ErrNotEnoughData       = errors.New("not enough data to decode a valid frame")
	ErrInvalidError        = errors.New("input contains invalid characters (CR or LF)")
)

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

	// String makes a frame implement the Stringer interface.
	String() string
}

type SimpleString struct {
	value string
}

func NewSimpleString(s string) (*SimpleString, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, ErrInvalidSimpleString
	}
	return &SimpleString{s}, nil
}

func (s *SimpleString) Serialize() []byte {
	return []byte(s.String())
}

func (s *SimpleString) String() string {
	return fmt.Sprintf("+%s\r\n", s.value)
}

// Deserialize implements Deserialize method from Framer interface
// This method reads a CRLF terminated slice of bytes from the buffer and turn it
// to a SimpleString Framer type. The read string should not contain any CR or LF
// as defined by the RESP protocol.
func (s *SimpleString) Deserialize(buff *bytes.Buffer) error {
	frameContent, err := getSimpleStringFromBuffer(buff)
	if err != nil {
		return err
	}
	s.value = frameContent
	return nil
}

func getSimpleStringFromBuffer(buff *bytes.Buffer) (string, error) {
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
	frameContent := string(readBytes[:len(readBytes)-2])
	// if there is any CR in the middle of the string,
	// it is not a frame of type simple string. We do not double-check for LF
	// because from the first read, we are guaranteed to not have any LF in the middle
	if strings.ContainsRune(frameContent, '\r') {
		return "", ErrInvalidSimpleString
	}
	return frameContent, nil
}

type Error struct {
	value string
}

func NewError(s string) (*Error, error) {
	if strings.ContainsAny(s, "\r\n") {
		return nil, ErrInvalidError
	}
	return &Error{s}, nil
}

func (e *Error) Serialize() []byte {
	return []byte(e.String())
}

func (e *Error) String() string {
	return "-" + string(e.value) + "\r\n"
}

// Deserialize implements Deserialize method from Framer interface
// This method reads a CRLF terminated slice of bytes from the buffer and turn it
// to an Error Framer type. The read string should not contain any CR or LF
// as defined by the RESP protocol.
func (e *Error) Deserialize(buff *bytes.Buffer) error {
	frameContent, err := getSimpleStringFromBuffer(buff)
	if err != nil {
		return err
	}
	e.value = frameContent
	return nil
}

//type IntegerFrame int64
//
//func (i IntegerFrame) Serialize() []uint8 {
//	return []uint8(fmt.Sprint(i))
//}
//
//func (i IntegerFrame) String() string {
//	return ":" + strconv.Itoa(int(i)) + "\r\n"
//}
//
//type BulkStringFrame string
//
//func (b BulkStringFrame) Serialize() []uint8 {
//	return []uint8(b)
//}
//
//func (b BulkStringFrame) String() string {
//	frameStr := "$-1\r\n"
//	if b != "" {
//		frameStr = "$" + strconv.Itoa(len(b)) + "\r\n" + string(b) + "\r\n"
//	}
//	return frameStr
//}
//
//func (s *BulkStringFrame) Deserialize(buff *bytes.Buffer) error {
//	if buff.Len() < 2 {
//		return ErrNotEnoughData
//	}
//	tmpRead := make([]byte, 0, 128)
//	for {
//		bs, err := buff.ReadBytes('\n')
//		if err != nil {
//			return err
//		}
//		tmpRead = append(tmpRead, bs...)
//		if tmpRead[len(tmpRead)-2] == '\r' {
//			break
//		}
//	}
//	s.value = string(tmpRead[:len(tmpRead)-2])
//	return nil
//}

//type BoolFrame bool
//
//func (b BoolFrame) Serialize() []uint8 {
//	return []uint8(fmt.Sprint(b))
//}
//
//func (b BoolFrame) String() string {
//	bValue := bool(b)
//	if bValue {
//		return "#t\r\n"
//	}
//	return "#f\r\n"
//}
//
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
