package server

import (
	"github.com/stretchr/testify/assert"
	"github.com/ynachi/gcache/command"
	"github.com/ynachi/gcache/frame"
	"github.com/ynachi/gcache/gerror"
	"log/slog"
	"testing"
)

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{
			name:  "ERROR",
			input: "ERROR",
			want:  slog.LevelError,
		},
		{
			name:  "Debug",
			input: "DEBUG",
			want:  slog.LevelDebug,
		},
		{
			name:  "LowerCaseWarn",
			input: "WARN",
			want:  slog.LevelWarn,
		},
		{
			name:  "Unknown",
			input: "Unknown",
			want:  slog.LevelInfo,
		},
		{
			name:  "EmptyString",
			input: "",
			want:  slog.LevelInfo,
		},
		{
			name:  "WhiteSpace",
			input: " ",
			want:  slog.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input
			assert.Equal(t, tt.want, getLogLevel(got))
		})
	}
}

func TestParseCommandFromFrame(t *testing.T) {
	// We use PING command to test parse command from frame
	simplePingFrameArray := frame.NewArray(1)
	_ = simplePingFrameArray.Append(frame.NewBulkString("PING"))

	complexPingFrameArray := frame.NewArray(2)
	_ = complexPingFrameArray.Append(frame.NewBulkString("PING"))
	_ = complexPingFrameArray.Append(frame.NewBulkString("Yoa"))

	simpleWrongCommand := frame.NewArray(1)
	_ = simpleWrongCommand.Append(frame.NewBulkString("PINGI"))

	complexPingFrameArrayInvalid := frame.NewArray(3)
	_ = complexPingFrameArrayInvalid.Append(frame.NewBulkString("PING"))
	_ = complexPingFrameArrayInvalid.Append(frame.NewBulkString("Yoa"))
	_ = complexPingFrameArrayInvalid.Append(frame.NewBulkString("wrong"))

	returnPing := func(f *frame.Array) *command.Ping {
		cmd := new(command.Ping)
		_ = cmd.FromFrame(f)
		return cmd
	}

	tests := []struct {
		name      string
		give      *frame.Array
		want      *command.Ping
		wantError error
	}{
		{
			name:      "valid simple ping command",
			give:      simplePingFrameArray,
			want:      returnPing(simplePingFrameArray),
			wantError: nil,
		},
		{
			name:      "valid complex ping command",
			give:      complexPingFrameArray,
			want:      returnPing(complexPingFrameArray),
			wantError: nil,
		},
		{
			name:      "simple ping wrong command",
			give:      simpleWrongCommand,
			want:      nil,
			wantError: gerror.ErrInvalidCmdName,
		},
		{
			name:      "complex ping wrong command",
			give:      complexPingFrameArrayInvalid,
			want:      nil,
			wantError: gerror.ErrInvalidPingCommand,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := parseCommandFromFrame(tt.give)
			assert.Equal(t, tt.wantError, err)
			if err == nil {
				assert.Equal(t, tt.want, f)
			}
		})
	}
}
