package log

import (
	"log/slog"

	"github.com/shenjing023/vivy-polaris/options"
)

type loggerOptions struct {
	maxFrameDepth int
	level         slog.Level
}

// WithMaxFrameDepth sets the maximum frame depth for the logger.
func WithMaxFrameDepth(depth int) options.Option[loggerOptions] {
	return options.NewFuncOption(func(o *loggerOptions) {
		o.maxFrameDepth = depth
	})
}

// WithLevel sets the log level for the logger.
func WithLevel(level slog.Level) options.Option[loggerOptions] {
	return options.NewFuncOption(func(o *loggerOptions) {
		o.level = level
	})
}
