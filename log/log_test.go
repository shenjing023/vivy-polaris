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
