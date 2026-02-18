package logger

import (
	"log/slog"
	"os"
	"time"
)

func New(level slog.Level) *slog.Logger {
	// 日本時間（JST）を設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.TimeValue(t.In(jst))
			}
			return a
		},
	})

	return slog.New(handler)
}
