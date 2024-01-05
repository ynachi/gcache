package frame

import (
	"bufio"
	"strings"
	"testing"
)

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
				t.Fatalf("DecodeError() unexpected gerror = %v", err)
			} else if tt.wantErr {
				t.Fatalf("DecodeError() expected gerror but got none.")
			}
			if *f != tt.wantFrame {
				t.Errorf("DecodeError() got = %v, want %v", *f, tt.wantFrame)
			}
		})
	}
}
