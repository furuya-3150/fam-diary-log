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

	// diary
	diary := e.Group("/diaries")
	diary.Use(auth.JWTAuthMiddleware(config.JWT.Secret, auth.FamilyCookieName))
	diary.POST("", diaryHandler.Create)
	diary.GET("", diaryHandler.List)
	diary.GET("/count/:year/:month", diaryHandler.GetCount)

	// streak
	diary.GET("/streak", diaryHandler.GetStreak)

	return e
}
