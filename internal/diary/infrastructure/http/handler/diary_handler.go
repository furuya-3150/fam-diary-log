package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/context"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
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
	req.FamilyID = e.Request().Context().Value(context.FamilyIDKey).(uuid.UUID)
	req.UserID = e.Request().Context().Value(context.UserIDKey).(uuid.UUID)

	res, err := dh.dc.Create(e.Request().Context(), req)
	if err != nil {
		log.Println("controller create error", err)
		return errors.RespondWithError(e, err)
	}

	return response.RespondSuccess(e, http.StatusOK, res)
}

// List handles GET /diaries request
func (dh *DiaryHandler) List(w http.ResponseWriter, r *http.Request) {
	// query := usecase.DiaryQuery{
	// 	FamilyID:  r.URL.Query().Get("family_id"),
	// 	UserID:    r.URL.Query().Get("user_id"),
	// 	StartDate: r.URL.Query().Get("start_date"),
	// 	EndDate:   r.URL.Query().Get("end_date"),
	// 	Limit:     10,
	// 	Offset:    0,
	// }

	// resp, err := dh.controller.List(r.Context(), query)
	// if err != nil {
	// 	w.WriteHeader(http.StatusInternalServerError)
	// 	json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
	// 	return
	// }

	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(nil)
}
