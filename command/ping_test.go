package command

import (
	"bufio"
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/ynachi/gcache/frame"
	"log/slog"
	"strings"
	"testing"
)

func TestPing_Apply(t *testing.T) {
	tests := []struct {
		name string
		give string
		want string
	}{
		{name: "RegularMessage", give: "hello", want: "hello"},
		{name: "EmptyMessage", give: ""},
		{name: "LongMessage", give: strings.Repeat("LongMessage", 100), want: strings.Repeat("LongMessage", 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeBuffer := &bytes.Buffer{}
			writer := bufio.NewWriter(writeBuffer)

			ping := Ping{
				message: tt.give,
				logger:  &slog.Logger{},
			}
			ping.Apply(nil, writer)

			f, _ := frame.Decode(bufio.NewReader(writeBuffer))
			got, ok := f.(*frame.SimpleString)
			if !ok {
				t.Fatalf("expected success, got gerror")
			}

			if got.Value() != tt.want {
				t.Errorf("wanted %v but got %v", tt.want, got.Value())
			}
		})
	}
}

func TestPing_FromFrame(t *testing.T) {
	simplePING := frame.NewArray(1)
	_ = simplePING.Append(frame.NewBulkString("PING"))

	pingWithMsg := frame.NewArray(2)
	_ = pingWithMsg.Append(frame.NewBulkString("PING"))
	_ = pingWithMsg.Append(frame.NewBulkString("Hello World"))

	wrongCmd := frame.NewArray(1)
	_ = wrongCmd.Append(frame.NewBulkString("PINg"))

	pingTooManyArgs := frame.NewArray(3)
	_ = pingTooManyArgs.Append(frame.NewBulkString("PING"))
	_ = pingTooManyArgs.Append(frame.NewBulkString("Hello World"))
	_ = pingTooManyArgs.Append(frame.NewBulkString("Wrong"))

	tests := []struct {
		name      string
		frame     *frame.Array
		want      string
		wantError error
	}{
		{name: "ValidSimplePing", frame: simplePING, want: "PONG", wantError: nil},
		{name: "ValidPingWithMessage", frame: pingWithMsg, want: "Hello World", wantError: nil},
		{name: "InvalidWrongCmdWord", frame: wrongCmd, want: "", wantError: ErrInvalidCmdName},
		{name: "InvalidTooManyArgs", frame: pingTooManyArgs, want: "", wantError: ErrInvalidPingCommand},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := Ping{logger: &slog.Logger{}}
			err := cmd.FromFrame(tt.frame)
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.want, cmd.message)
		})
	}
}
