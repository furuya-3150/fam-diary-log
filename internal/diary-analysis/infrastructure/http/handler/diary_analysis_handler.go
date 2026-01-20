package handler

import (
	"net/http"

	infctx "github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/context"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type DiaryAnalysisHandler struct {
	dau usecase.DiaryAnalysisUsecase
}

func NewDiaryAnalysisHandler(dau usecase.DiaryAnalysisUsecase) *DiaryAnalysisHandler {
	return &DiaryAnalysisHandler{
		dau: dau,
	}
}

// handleWeekCount is a common handler for week-based counts
func (dah *DiaryAnalysisHandler) handleWeekCount(c echo.Context, usecaseFunc func(echo.Context, uuid.UUID, string) (map[string]interface{}, error)) error {
	// Extract user ID from context
	userID, ok := c.Request().Context().Value(infctx.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.LogicError{Message: "userIDを指定してください"})
	}

	// Extract date from URL param
	date := c.Param("date")

	// Call usecase function
	countByDate, err := usecaseFunc(c, userID, date)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	// Return response
	return response.RespondSuccess(c, http.StatusOK, countByDate)
}

// GetWeekCharCount handles GET /week-char-count/:date
func (dah *DiaryAnalysisHandler) GetWeekCharCount(c echo.Context) error {
	return dah.handleWeekCount(c, func(ctx echo.Context, userID uuid.UUID, date string) (map[string]interface{}, error) {
		return dah.dau.GetCharCountByDate(ctx.Request().Context(), userID, date)
	})
}

// GetWeekSentenceCount handles GET /week-sentence-count/:date
func (dah *DiaryAnalysisHandler) GetWeekSentenceCount(c echo.Context) error {
	return dah.handleWeekCount(c, func(ctx echo.Context, userID uuid.UUID, date string) (map[string]interface{}, error) {
		return dah.dau.GetSentenceCountByDate(ctx.Request().Context(), userID, date)
	})
}

// GetWeekAccuracyScore handles GET /week-accuracy-score/:date
func (dah *DiaryAnalysisHandler) GetWeekAccuracyScore(c echo.Context) error {
	return dah.handleWeekCount(c, func(ctx echo.Context, userID uuid.UUID, date string) (map[string]interface{}, error) {
		return dah.dau.GetAccuracyScoreByDate(ctx.Request().Context(), userID, date)
	})
}
