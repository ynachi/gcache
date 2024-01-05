package frame

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

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
				t.Fatalf("DecodeBulkString() unexpected gerror = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeBulkString() expected gerror but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeBulkString() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
