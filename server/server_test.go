package server

import (
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
			name:  "Warn",
			input: "Warn",
			want:  slog.LevelWarn,
		},
		{
			name:  "ERROR",
			input: "ERROR",
			want:  slog.LevelError,
		},
		{
			name:  "Debug",
			input: "Debug",
			want:  slog.LevelDebug,
		},
		{
			name:  "Unknown",
			input: "Unknown",
			want:  slog.LevelInfo,
		},
		{
			name:  "LowerCaseWarn",
			input: "warn",
			want:  slog.LevelWarn,
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
			if tt.want != getLogLevel(got) {
				t.Errorf("GetLogLevel got = %v, want %v", got, tt.want)
			}
		})
	}
}
