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

func TestInteger_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		give      *bytes.Buffer
		wantFrame Integer
		wantErr   bool
	}{
		{
			name:      "basic frame with positive number",
			give:      bytes.NewBufferString("25\r\n"),
			wantFrame: Integer{value: 25},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bytes.NewBufferString("-25\r\nworld"),
			wantFrame: Integer{value: -25},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bytes.NewBufferString("25\rlo\r\n"),
			wantFrame: Integer{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bytes.NewBufferString("25\nlo\r\n"),
			wantFrame: Integer{},
			wantErr:   true,
		},
		{
			name:      "frame data contains valid integer",
			give:      bytes.NewBufferString("25\nlo\r\n"),
			wantFrame: Integer{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Integer{}
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

func TestSimpleBulk_String(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{
			name: "simple valid string",
			give: "hello",
			want: "$5\r\nhello\r\n",
		},
		{
			name: "empty string encoding",
			give: "",
			want: "$0\r\n\r\n",
		},
		{
			name: "a valid bulk string can be numeric",
			give: "25",
			want: "$2\r\n25\r\n",
		},
		{
			name: "a valid bulk string can contain CR in the middle",
			give: "hello\rhi",
			want: "$8\r\nhello\rhi\r\n",
		},
		{
			name: "a valid bulk string can contain LF in the middle",
			give: "hello\nhi",
			want: "$8\r\nhello\nhi\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BulkString{
				value: tt.give,
			}
			if got := fmt.Sprint(s); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBulkString_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		give      *bytes.Buffer
		wantFrame BulkString
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bytes.NewBufferString("5\r\nhello\r\n"),
			wantFrame: BulkString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer after read",
			give:      bytes.NewBufferString("5\r\nhello\r\nworld"),
			wantFrame: BulkString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame can contain CR in the middle",
			give:      bytes.NewBufferString("6\r\nhel\rlo\r\n"),
			wantFrame: BulkString{value: "hel\rlo"},
			wantErr:   false,
		},
		{
			name:      "working frame can contain LF in the middle",
			give:      bytes.NewBufferString("6\r\nhel\nlo\r\n"),
			wantFrame: BulkString{value: "hel\nlo"},
			wantErr:   false,
		},
		{
			name:      "frame size does not match data size",
			give:      bytes.NewBufferString("8\r\nhello\r\n"),
			wantFrame: BulkString{},
			wantErr:   true,
		},
		{
			name:      "frame does not end with CRLF",
			give:      bytes.NewBufferString("5\r\nhello\r"),
			wantFrame: BulkString{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BulkString{}
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

func TestBool_String(t *testing.T) {
	tests := []struct {
		name string
		give bool
		want string
	}{
		{
			name: "test valid true",
			give: true,
			want: "#t\r\n",
		},
		{
			name: "test valid false",
			give: false,
			want: "#f\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Bool{
				value: tt.give,
			}
			if got := fmt.Sprint(s); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBool_Deserialize(t *testing.T) {
	tests := []struct {
		name      string
		give      *bytes.Buffer
		wantFrame Bool
		wantErr   bool
	}{
		{
			name:      "get true from buffer",
			give:      bytes.NewBufferString("t\r\n"),
			wantFrame: Bool{value: true},
			wantErr:   false,
		},
		{
			name:      "get false from buffer",
			give:      bytes.NewBufferString("f\r\nworld"),
			wantFrame: Bool{value: false},
			wantErr:   false,
		},
		{
			name:      "invalid bool",
			give:      bytes.NewBufferString("T\r\n"),
			wantFrame: Bool{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Bool{}
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
