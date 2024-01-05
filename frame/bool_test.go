package frame

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

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
				t.Fatalf("DecodeBool() unexpected gerror = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeBool() expected gerror but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeBool() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
