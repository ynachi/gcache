package server

import (
	"github.com/stretchr/testify/assert"
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
