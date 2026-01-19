package http

import (
	"log/slog"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/broker"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/handler"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/diary/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	config := config.Load()

	dbManager := db.NewDBManager(config.DB.DatabaseURL)
	txManager := db.NewTransaction(dbManager)

	pub := broker.NewDiaryPublisher(slog.Default())

	diaryRepo := repository.NewDiaryRepository(dbManager)
	diaryUsecase := usecase.NewDiaryUsecaseWithPublisher(txManager, diaryRepo, pub)
	diaryController := controller.NewDiaryController(diaryUsecase)
	diaryHandler := handler.NewDiaryHandler(diaryController)

	e := echo.New()

	e.GET("healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// diary
	diary := e.Group("/diaries")
	diary.POST("", diaryHandler.Create)
	diary.GET("", diaryHandler.List)
	diary.GET("/count/:year/:month", diaryHandler.GetCount)

	return e
}
