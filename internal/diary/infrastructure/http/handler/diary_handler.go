package handler

import (
	"log"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller"
	dto "github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DiaryHandler handles HTTP requests for diary operations
type DiaryHandler struct {
	dc controller.DiaryController
}

// NewDiaryHandler creates a new instance of DiaryHandler
func NewDiaryHandler(dc controller.DiaryController) *DiaryHandler {
	return &DiaryHandler{
		dc: dc,
	}
}

func (dh *DiaryHandler) Create(e echo.Context) error {
	var req *domain.Diary
	if err := e.Bind(&req); err != nil {
		log.Println("bind error", err)
		validationErr := &errors.ValidationError{Message: err.Error()}
		return errors.RespondWithError(e, validationErr)
	}
	req.FamilyID = e.Request().Context().Value(auth.ContextKeyFamilyID).(uuid.UUID)
	req.UserID = e.Request().Context().Value(auth.ContextKeyUserID).(uuid.UUID)

	res, err := dh.dc.Create(e.Request().Context(), req)
	if err != nil {
		log.Println("controller create error", err)
		return errors.RespondWithError(e, err)
	}

	return response.RespondSuccess(e, http.StatusOK, res)
}

func (dh *DiaryHandler) List(e echo.Context) error {
	familyID := e.Request().Context().Value(auth.ContextKeyFamilyID).(uuid.UUID)

	// validate query
	q := dto.DiaryListQuery{TargetDate: e.QueryParam("target_date")}
	v := validator.New()
	if err := v.Struct(q); err != nil {
		validationErr := &errors.ValidationError{Message: "target_date is required and must be YYYY-MM-DD"}
		return errors.RespondWithError(e, validationErr)
	}

	ctx := e.Request().Context()

	res, err := dh.dc.List(ctx, familyID, q.TargetDate)
	if err != nil {
		log.Println("controller list error", err)
		return errors.RespondWithError(e, err)
	}

	return response.RespondSuccess(e, http.StatusOK, res)
}

func (dh *DiaryHandler) GetCount(e echo.Context) error {
	familyID := e.Request().Context().Value(auth.ContextKeyFamilyID).(uuid.UUID)
	yearStr := e.QueryParam("year")
	monthStr := e.QueryParam("month")

	// Validate required query parameters
	if yearStr == "" || monthStr == "" {
		validationErr := &errors.ValidationError{Message: "year and month query parameters are required"}
		return errors.RespondWithError(e, validationErr)
	}

	count, err := dh.dc.GetCount(e.Request().Context(), familyID, yearStr, monthStr)
	if err != nil {
		log.Println("controller get count error", err)
		return errors.RespondWithError(e, err)
	}

	return response.RespondSuccess(e, http.StatusOK, map[string]int{"count": count})
}

func (dh *DiaryHandler) GetStreak(e echo.Context) error {
	userID := e.Request().Context().Value(auth.ContextKeyUserID).(uuid.UUID)
	familyID := e.Request().Context().Value(auth.ContextKeyFamilyID).(uuid.UUID)

	res, err := dh.dc.GetStreak(e.Request().Context(), userID, familyID)
	if err != nil {
		log.Println("controller get streak error", err)
		return errors.RespondWithError(e, err)
	}

	return response.RespondSuccess(e, http.StatusOK, res)
}
