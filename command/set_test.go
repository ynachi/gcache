package command

import (
	"bufio"
	"bytes"
	"github.com/ynachi/gcache/db"
	"github.com/ynachi/gcache/frame"
	"log/slog"
	"testing"
)

var _cache, _ = db.NewCache(5, "LRU")

func TestSet_Apply_Ok(t *testing.T) {
	tests := []struct {
		name      string
		giveKey   string
		giveValue string
		want      string
	}{
		{name: "Success", giveKey: "hello", giveValue: "world", want: "ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeBuffer := &bytes.Buffer{}
			writer := bufio.NewWriter(writeBuffer)

			ping := Set{
				key:    tt.giveKey,
				value:  tt.giveValue,
				logger: &slog.Logger{},
			}
			ping.Apply(_cache, writer)

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

//func TestSet_FromFrame(t *testing.T) {
//	simplePING := frame.NewArray(1)
//	_ = simplePING.Append(frame.NewBulkString("PING"))
//
//	pingWithMsg := frame.NewArray(2)
//	_ = pingWithMsg.Append(frame.NewBulkString("PING"))
//	_ = pingWithMsg.Append(frame.NewBulkString("Hello World"))
//
//	pingTooManyArgs := frame.NewArray(3)
//	_ = pingTooManyArgs.Append(frame.NewBulkString("PING"))
//	_ = pingTooManyArgs.Append(frame.NewBulkString("Hello World"))
//	_ = pingTooManyArgs.Append(frame.NewBulkString("Wrong"))
//
//	tests := []struct {
//		name      string
//		frame     *frame.Array
//		want      string
//		wantError error
//	}{
//		{name: "ValidSimplePing", frame: simplePING, want: "PONG", wantError: nil},
//		{name: "ValidPingWithMessage", frame: pingWithMsg, want: "Hello World", wantError: nil},
//		{name: "InvalidTooManyArgs", frame: pingTooManyArgs, want: "", wantError: gerror.ErrInvalidPingCommand},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			cmd := Ping{logger: &slog.Logger{}}
//			err := cmd.FromFrame(tt.frame)
//			assert.Equal(t, tt.wantError, err)
//			assert.Equal(t, tt.want, cmd.message)
//		})
//	}
//}
