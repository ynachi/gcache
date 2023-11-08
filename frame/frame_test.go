package frame

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func TestSimpleString_String(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{
			name: "simple string should work",
			give: "OK",
			want: "+OK\r\n",
		},
		{
			name: "empty string should work",
			give: "",
			want: "+\r\n",
		},
		{
			name: "integer based string should work",
			give: "732",
			want: "+732\r\n",
		},
		{
			name: "non ascii simple string should work",
			give: "æ",
			want: "+æ\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleString{
				value: tt.give,
			}
			if got := fmt.Sprint(s); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleString_Serialize(t *testing.T) {
	tests := []struct {
		name string
		give string
		want []byte
	}{
		{
			name: "simple string should work",
			give: "OK",
			want: []byte("+OK\r\n"),
		},
		{
			name: "empty string should work",
			give: "",
			want: []byte("+\r\n"),
		},
		{
			name: "integer based string should work",
			give: "732",
			want: []byte("+732\r\n"),
		},
		{
			name: "non ascii simple string should work",
			give: "æ",
			want: []byte("+æ\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleString{
				value: tt.give,
			}
			if got := s.Serialize(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Serialize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSimpleString(t *testing.T) {
	tests := []struct {
		name    string
		give    string
		want    *SimpleString
		wantErr bool
	}{
		{
			name:    "a string without CR or LF should work",
			give:    "OK",
			want:    &SimpleString{"OK"},
			wantErr: false,
		},
		{
			name:    "a string with CRLF should fail",
			give:    "OK\r\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "a string with CR should fail",
			give:    "O\rK",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "a string with LF should fail",
			give:    "OK\n",
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSimpleString(tt.give)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSimpleString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSimpleString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleString_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		give      *bytes.Buffer
		wantFrame SimpleString
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bytes.NewBufferString("hello\r\n"),
			wantFrame: SimpleString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bytes.NewBufferString("hello\r\nworld"),
			wantFrame: SimpleString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bytes.NewBufferString("hel\rlo\r\n"),
			wantFrame: SimpleString{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bytes.NewBufferString("hel\nlo\r\n"),
			wantFrame: SimpleString{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SimpleString{}
			err := s.Deserialize(tt.give)
			if *s != tt.wantFrame {
				t.Errorf("Deserialize() got = %v, want %v", *s, tt.wantFrame)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestError_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		give      *bytes.Buffer
		wantFrame Error
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bytes.NewBufferString("hello\r\n"),
			wantFrame: Error{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bytes.NewBufferString("hello\r\nworld"),
			wantFrame: Error{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bytes.NewBufferString("hel\rlo\r\n"),
			wantFrame: Error{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bytes.NewBufferString("hel\nlo\r\n"),
			wantFrame: Error{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Error{}
			err := s.Deserialize(tt.give)
			if *s != tt.wantFrame {
				t.Errorf("Deserialize() got = %v, want %v", *s, tt.wantFrame)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

//func TestBulkString_Deserialize(t *testing.T) {
//	tests := []struct {
//		name      string
//		give      *bytes.Buffer
//		wantFrame SimpleString
//		wantErr   bool
//	}{
//		// TODO: Add test cases.
//		{
//			name:      "basic working frame",
//			give:      bytes.NewBufferString("hello\r\n"),
//			wantFrame: SimpleString{value: "hello"},
//			wantErr:   false,
//		},
//		{
//			name:      "basic working frame with data left in the buffer",
//			give:      bytes.NewBufferString("hello\r\nworld"),
//			wantFrame: SimpleString{value: "hello"},
//			wantErr:   false,
//		},
//		{
//			name:      "working frame can contain CR in the middle",
//			give:      bytes.NewBufferString("hel\rlo\r\n"),
//			wantFrame: SimpleString{value: "hel\rlo"},
//			wantErr:   false,
//		},
//		{
//			name:      "working frame can contain LF in the middle",
//			give:      bytes.NewBufferString("hel\nlo\r\n"),
//			wantFrame: SimpleString{value: "hel\nlo"},
//			wantErr:   false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &SimpleString{}
//			err := s.Deserialize(tt.give)
//			if *s != tt.wantFrame {
//				t.Errorf("Deserialize() got = %v, want %v", *s, tt.wantFrame)
//			}
//			if (err != nil) != tt.wantErr {
//				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//		})
//	}
//}
