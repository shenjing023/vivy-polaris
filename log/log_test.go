package log

import (
	"log/slog"
	"testing"

	"github.com/mdobak/go-xerrors"
)

func TestError(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{"TestError"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init()
			err := xerrors.New("test error")
			slog.Error("test error", "err", err)
		})
	}
}

func TestLogOptions(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
		{"TestLogOptions"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(WithLevel(slog.LevelInfo),
				WithMaxFrameDepth(1))
			slog.Info("test info")
			slog.Debug("test debug")
			slog.Warn("test warn")
			slog.Error("test error")

			err := xerrors.New("test error")
			slog.Error("test error", "err", err)
		})
	}
}
