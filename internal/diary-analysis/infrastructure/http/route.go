package http

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/http/handler"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	cfg := config.Load()

	dbManager := db.NewDBManager(cfg.DB.DatabaseURL)

	diaryAnalysisRepo := repository.NewDiaryAnalysisRepository(dbManager)
	diaryAnalysisUsecase := usecase.NewDiaryAnalysisUsecase(diaryAnalysisRepo)
	diaryAnalysisHandler := handler.NewDiaryAnalysisHandler(diaryAnalysisUsecase)

	e := echo.New()

	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// diary analyses
	analyses := e.Group("/analyses")
	analyses.Use(auth.JWTAuthMiddleware(cfg.JWT.Secret, auth.FamilyCookieName))
	analyses.GET("/week-char-count/:date", diaryAnalysisHandler.GetWeekCharCount)
	analyses.GET("/week-sentence-count/:date", diaryAnalysisHandler.GetWeekSentenceCount)
	analyses.GET("/week-accuracy-score/:date", diaryAnalysisHandler.GetWeekAccuracyScore)
	analyses.GET("/week-writing-time-seconds/:date", diaryAnalysisHandler.GetWeekWritingTime)

	return e
}
