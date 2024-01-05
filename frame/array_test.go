package frame

import (
	"bufio"
	"reflect"
	"strings"
	"testing"
)

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
			s := tt.give
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
				t.Fatalf("DecodeArray() unexpected gerror = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeArray() expected gerror but got none.")
			}
			if !reflect.DeepEqual(*f, tt.wantFrame) {
				t.Errorf("DecodeArray() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
