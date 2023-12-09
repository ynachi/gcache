package frame

import (
	"bufio"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestSimpleInteger_String(t *testing.T) {
	tests := []struct {
		name string
		give int64
		want string
	}{
		{
			name: "positive integer",
			give: 34,
			want: ":34\r\n",
		},
		{
			name: "negative integer",
			give: -34,
			want: ":-34\r\n",
		},
		{
			name: "integer of value 0",
			give: 0,
			want: ":0\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Integer{
				value: tt.give,
			}
			if got := fmt.Sprint(s); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimpleInteger_Serialize(t *testing.T) {
	tests := []struct {
		name string
		give int64
		want []byte
	}{
		{
			name: "positive integer",
			give: 34,
			want: []byte(":34\r\n"),
		},
		{
			name: "negative integer",
			give: -34,
			want: []byte(":-34\r\n"),
		},
		{
			name: "integer of value 0",
			give: 0,
			want: []byte(":0\r\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Integer{
				value: tt.give,
			}
			if got := s.Serialize(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Serialize(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInteger_DecodeInteger(t *testing.T) {
	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame Integer
		wantErr   bool
	}{
		{
			name:      "basic frame with positive number",
			give:      bufio.NewReader(strings.NewReader("25\r\n")),
			wantFrame: Integer{value: 25},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bufio.NewReader(strings.NewReader("-25\r\nworld")),
			wantFrame: Integer{value: -25},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bufio.NewReader(strings.NewReader("25\rlo\r\n")),
			wantFrame: Integer{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bufio.NewReader(strings.NewReader("25\nlo\r\n")),
			wantFrame: Integer{},
			wantErr:   true,
		},
		{
			name:      "frame data contains valid integer",
			give:      bufio.NewReader(strings.NewReader("25\nlo\r\n")),
			wantFrame: Integer{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeInteger(tt.give)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("DecodeInteger() unexpected error = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeInteger() expected error but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeInteger() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
