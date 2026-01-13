package logger

import "log/slog"

type Option func(*slog.HandlerOptions)

func WithLevel(level slog.Level) Option {
	return func(o *slog.HandlerOptions) {
		o.Level = level
	}
}
