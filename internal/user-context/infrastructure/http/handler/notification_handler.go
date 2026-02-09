package handler

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationHandler interface {
	GetNotificationSetting(c echo.Context) error
	UpdateNotificationSetting(c echo.Context) error
}

type notificationHandler struct {
	nu usecase.NotificationUsecase
}

func NewNotificationHandler(nu usecase.NotificationUsecase) NotificationHandler {
	return &notificationHandler{nu: nu}
}

func (h *notificationHandler) GetNotificationSetting(c echo.Context) error {
	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	val = ctx.Value(auth.ContextKeyFamilyID)
	familyID, ok := val.(uuid.UUID)
	if !ok || familyID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid family_id context"})
	}

	s, err := h.nu.GetNotificationSetting(ctx, userID, familyID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	resp := dto.NotificationSettingResponse{
		FamilyID: s.FamilyID,
		PostCreatedEnabled: s.PostCreatedEnabled,
	}
	return response.RespondSuccess(c, http.StatusOK, resp)
}

func (h *notificationHandler) UpdateNotificationSetting(c echo.Context) error {
	var req dto.NotificationSettingRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body: " + err.Error()})
	}
	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	val = ctx.Value(auth.ContextKeyFamilyID)
	familyID, ok := val.(uuid.UUID)
	if !ok || familyID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid family_id context"})
	}
	req.FamilyID = familyID

	ns := &domain.NotificationSetting{
		UserID:             userID,
		FamilyID:           req.FamilyID,
		PostCreatedEnabled: req.PostCreatedEnabled,
	}
	if err := h.nu.UpdateNotificationSetting(ctx, ns); err != nil {
		return errors.RespondWithError(c, err)
	}

	return response.RespondSuccess(c, http.StatusNoContent, nil)
}
