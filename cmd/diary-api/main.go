package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http"
	"github.com/joho/godotenv"
)

func main() {
	slog.Info("Starting diary API server... 123")
	e := http.NewRouter()

	e.Logger.Fatal(e.Start(":8080"))
}

func init() {
	// 日本時間（JST）を設定
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	// ログ設定
	var handler slog.Handler
	if os.Getenv("GO_ENV") == "dev" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					t := a.Value.Time()
					a.Value = slog.TimeValue(t.In(jst))
				}
				return a
			},
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey {
					t := a.Value.Time()
					a.Value = slog.TimeValue(t.In(jst))
				}
				return a
			},
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	// env読み込み
	if os.Getenv("GO_ENV") == "dev" {
		err = godotenv.Load("./cmd/diary-api/.env")
		if err != nil {
			slog.Error("Error loading .env file", "Error", err.Error())
			os.Exit(1)
		}
	}

	// config読み込み
	config.Cfg = config.Load()
}
