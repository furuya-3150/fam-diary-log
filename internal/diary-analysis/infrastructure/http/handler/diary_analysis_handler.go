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

// GetWeekCharCount handles GET /week-char-count/:date
func (dah *DiaryAnalysisHandler) GetWeekCharCount(c echo.Context) error {
	// Extract user ID from context
	userID, ok := c.Request().Context().Value(infctx.UserIDKey).(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.LogicError{Message: "userIDを指定してください"})
	}

	// Extract date from URL param
	date := c.Param("date")

	// Call usecase to get char count by date
	charCountByDate, err := dah.dau.GetCharCountByDate(c.Request().Context(), userID, date)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	// Return response with date-based char counts
	return response.RespondSuccess(c, http.StatusOK, charCountByDate)
}
