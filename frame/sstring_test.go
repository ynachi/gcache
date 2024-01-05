package frame

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
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
				t.Errorf("NewSimpleString() gerror = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSimpleString() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDecodeSimpleString(t *testing.T) {
	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame SimpleString
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bufio.NewReader(strings.NewReader("hello\r\n")),
			wantFrame: SimpleString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bufio.NewReader(strings.NewReader("hello\r\nworld")),
			wantFrame: SimpleString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bufio.NewReader(strings.NewReader("hel\rlo\r\n")),
			wantFrame: SimpleString{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bufio.NewReader(strings.NewReader("hel\nlo\r\n")),
			wantFrame: SimpleString{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeSimpleString(tt.give)
			if err != nil { // an gerror occurred
				if tt.wantErr { // but it is expected
					return // so the test is successful.
				}
				// not expected though, fail the test.
				t.Fatalf("DecodeSimpleString() unexpected gerror = %v", err)
			} else if tt.wantErr { // no gerror but one was expected!
				t.Fatalf("DecodeSimpleString() expected gerror but got none.")
			}
			// finally, if no errors and none are expected, check the result:
			if *f != tt.wantFrame {
				t.Errorf("DecodeSimpleString() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
