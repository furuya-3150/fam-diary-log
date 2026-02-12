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

	// diary analyses - authenticated user's family context
	analyses := e.Group("/families/me/diaries/analyses")
	analyses.Use(auth.JWTAuthMiddleware(cfg.JWT.Secret), auth.RequireFamily())
	analyses.GET("/weekly-char-count", diaryAnalysisHandler.GetWeekCharCount)
	analyses.GET("/weekly-sentence-count", diaryAnalysisHandler.GetWeekSentenceCount)
	analyses.GET("/weekly-accuracy-score", diaryAnalysisHandler.GetWeekAccuracyScore)
	analyses.GET("/weekly-writing-time", diaryAnalysisHandler.GetWeekWritingTime)

	return e
}
