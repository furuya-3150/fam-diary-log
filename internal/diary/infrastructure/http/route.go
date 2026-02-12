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
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	config := config.Load()

	dbManager := db.NewDBManager(config.DB.DatabaseURL)
	txManager := db.NewTransaction(dbManager)

	pub := broker.NewDiaryPublisher(slog.Default())

	clock := &clock.Real{}
	diaryRepo := repository.NewDiaryRepository(dbManager)
	streakRepo := repository.NewStreakRepository(dbManager)
	diaryUsecase := usecase.NewDiaryUsecase(txManager, diaryRepo, streakRepo, pub, clock)
	diaryController := controller.NewDiaryController(diaryUsecase)
	diaryHandler := handler.NewDiaryHandler(diaryController)

	e := echo.New()

	e.GET("healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// family diaries - authenticated user's family context
	diaries := e.Group("/families/me/diaries")
	diaries.Use(auth.JWTAuthMiddleware(config.JWT.Secret), auth.RequireFamily())
	diaries.POST("", diaryHandler.Create)
	diaries.GET("", diaryHandler.List)
	diaries.GET("/count", diaryHandler.GetCount)
	diaries.GET("/streak", diaryHandler.GetStreak)

	return e
}
