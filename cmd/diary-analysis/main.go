package main

import (
	"log/slog"
	"os"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http"
	"github.com/furuya-3150/fam-diary-log/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	e := http.NewRouter()

	e.Logger.Fatal(e.Start(":8081"))
}

func init() {
	// ログ設定
	var log *slog.Logger
	if os.Getenv("GO_ENV") == "dev" {
		log = logger.New(slog.LevelDebug)
	} else {
		log = logger.New(slog.LevelInfo)
	}

	slog.SetDefault(log)

	// env読み込み
	if os.Getenv("GO_ENV") == "dev" {
		err := godotenv.Load("./cmd/diary-analysis/.env")
		if err != nil {
			slog.Error("Error loading .env file", "Error", err.Error())
			os.Exit(1)
		}
	}

	// config読み込み
	config.Cfg = config.Load()
}
