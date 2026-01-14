package http

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/db"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/handler"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/diary/usecase"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	dbManager := db.NewDBManger()
	txManager := db.NewTransaction(dbManager)



	diaryRepo := repository.NewDiaryRepository(dbManager)
	diaryUsecase := usecase.NewDiaryUsecase(txManager, diaryRepo)
	diaryController := controller.NewDiaryController(diaryUsecase)
	diaryHandler := handler.NewDiaryHandler(diaryController)

	e := echo.New()

	e.GET("healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// diary
	diary := e.Group("/diaries")
	diary.POST("", diaryHandler.Create)

	return e
}