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
				t.Errorf("NewSimpleString() error = %v, wantErr %v", err, tt.wantErr)
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
			if err != nil { // an error occurred
				if tt.wantErr { // but it is expected
					return // so the test is successful.
				}
				// not expected though, fail the test.
				t.Fatalf("DecodeSimpleString() unexpected error = %v", err)
			} else if tt.wantErr { // no error but one was expected!
				t.Fatalf("DecodeSimpleString() expected error but got none.")
			}
			// finally, if no errors and none are expected, check the result:
			if *f != tt.wantFrame {
				t.Errorf("DecodeSimpleString() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}

func TestError_DecodeError(t *testing.T) {
	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame Error
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bufio.NewReader(strings.NewReader("hello\r\n")),
			wantFrame: Error{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer",
			give:      bufio.NewReader(strings.NewReader("hello\r\nworld")),
			wantFrame: Error{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame should not contain CR in the middle",
			give:      bufio.NewReader(strings.NewReader("hel\rlo\r\n")),
			wantFrame: Error{},
			wantErr:   true,
		},
		{
			name:      "working frame should not contain LF in the middle",
			give:      bufio.NewReader(strings.NewReader("hel\nlo\r\n")),
			wantFrame: Error{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeError(tt.give)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("DecodeError() unexpected error = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeError() expected error but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeError() got = %v, want %v", *f, tt.wantFrame)
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

func TestBulk_String(t *testing.T) {
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

func TestBulkString_DecodeBulkString(t *testing.T) {
	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame BulkString
		wantErr   bool
	}{
		{
			name:      "basic working frame",
			give:      bufio.NewReader(strings.NewReader("5\r\nhello\r\n")),
			wantFrame: BulkString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "basic working frame with data left in the buffer after read",
			give:      bufio.NewReader(strings.NewReader("5\r\nhello\r\nworld")),
			wantFrame: BulkString{value: "hello"},
			wantErr:   false,
		},
		{
			name:      "working frame can contain CR in the middle",
			give:      bufio.NewReader(strings.NewReader("6\r\nhel\rlo\r\n")),
			wantFrame: BulkString{value: "hel\rlo"},
			wantErr:   false,
		},
		{
			name:      "working frame can contain LF in the middle",
			give:      bufio.NewReader(strings.NewReader("6\r\nhel\nlo\r\n")),
			wantFrame: BulkString{value: "hel\nlo"},
			wantErr:   false,
		},
		{
			name:      "frame size does not match data size",
			give:      bufio.NewReader(strings.NewReader("8\r\nhello\r\n")),
			wantFrame: BulkString{},
			wantErr:   true,
		},
		{
			name:      "frame does not end with CRLF",
			give:      bufio.NewReader(strings.NewReader("5\r\nhello\r")),
			wantFrame: BulkString{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeBulkString(tt.give)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("DecodeBulkString() unexpected error = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeBulkString() expected error but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeBulkString() got = %v, want %v", *f, tt.wantFrame)
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

func TestBool_DecodeBool(t *testing.T) {
	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame Bool
		wantErr   bool
	}{
		{
			name:      "get true from buffer",
			give:      bufio.NewReader(strings.NewReader("t\r\n")),
			wantFrame: Bool{value: true},
			wantErr:   false,
		},
		{
			name:      "get false from buffer",
			give:      bufio.NewReader(strings.NewReader("f\r\nworld")),
			wantFrame: Bool{value: false},
			wantErr:   false,
		},
		{
			name:      "invalid bool",
			give:      bufio.NewReader(strings.NewReader("T\r\n")),
			wantFrame: Bool{},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeBool(tt.give)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("DecodeBool() unexpected error = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeBool() expected error but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeBool() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}

func TestArray_String(t *testing.T) {
	tests := []struct {
		name string
		give Array
		want string
	}{
		{
			name: "empty array",
			give: Array{},
			want: "*0\r\n",
		},
		{
			name: "array of bulk strings",
			give: Array{
				value: []Framer{
					&BulkString{value: "hello"},
					&BulkString{value: "world"},
				},
			},
			want: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
		},
		{
			name: "mixed frame types",
			give: Array{
				value: []Framer{
					&BulkString{value: "hello"},
					&SimpleString{value: "world"},
					&Integer{value: 25},
				},
			},
			want: "*3\r\n$5\r\nhello\r\n+world\r\n:25\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &tt.give
			if got := s.String(); got != tt.want {
				t.Errorf("String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArray_DecodeArray(t *testing.T) {
	arrayOfBulks := NewArray(2)
	_ = arrayOfBulks.Append(&BulkString{value: "hello"})
	_ = arrayOfBulks.Append(&BulkString{value: "world"})

	arrayOfMixed := NewArray(2)
	_ = arrayOfMixed.Append(&BulkString{value: "hello"})
	_ = arrayOfMixed.Append(&Integer{value: 28})
	_ = arrayOfMixed.Append(&SimpleString{value: "simple"})
	_ = arrayOfMixed.Append(&Null{})
	_ = arrayOfMixed.Append(&Bool{value: true})

	tests := []struct {
		name      string
		give      *bufio.Reader
		wantFrame Array
		wantErr   bool
	}{
		{
			name:      "array of bulk",
			give:      bufio.NewReader(strings.NewReader("2\r\n$5\r\nhello\r\n$5\r\nworld\r\n")),
			wantFrame: *arrayOfBulks,
			wantErr:   false,
		},
		{
			name:      "array of mixed frame types",
			give:      bufio.NewReader(strings.NewReader("2\r\n$5\r\nhello\r\n:28\r\n+simple\r\n_\r\n#t\r\n")),
			wantFrame: *arrayOfMixed,
			wantErr:   false,
		},
		{
			name:      "array of mixed frame types with extra invalid data",
			give:      bufio.NewReader(strings.NewReader("2\r\n$5\r\nhello\r\n:28\r\n+simple\r\n_\r\n+simple2")),
			wantFrame: *arrayOfMixed,
			wantErr:   false,
		},
		{
			name:      "empty array",
			give:      bufio.NewReader(strings.NewReader("0\r\n")),
			wantFrame: *NewArray(0),
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := DecodeArray(tt.give)
			if err != nil {
				if tt.wantErr {
					return
				}
				t.Fatalf("DecodeArray() unexpected error = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeArray() expected error but got none.")
			}
			if !reflect.DeepEqual(*f, tt.wantFrame) {
				t.Errorf("DecodeArray() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
