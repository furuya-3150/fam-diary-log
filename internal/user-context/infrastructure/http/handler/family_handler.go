package handler

import (
	"log/slog"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FamilyHandler interface {
	CreateFamily(c echo.Context) error
	InviteMembers(c echo.Context) error
	ApplyToFamily(c echo.Context) error
	RespondToJoinRequest(c echo.Context) error
	JoinFamily(c echo.Context) error
}

type familyHandler struct {
	fc controller.FamilyController
	fu usecase.FamilyUsecase
}

func NewFamilyHandler(familyController controller.FamilyController, familyUsecase usecase.FamilyUsecase) FamilyHandler {
	return &familyHandler{
		fc: familyController,
		fu: familyUsecase,
	}
}

// CreateFamily POST /families
func (h *familyHandler) CreateFamily(c echo.Context) error {
	var req dto.CreateFamilyRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body" + err.Error()})
	}

	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request"})
	}

	token, err := h.fu.CreateFamily(ctx, req.Name, userID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	accessTokenCookie := &http.Cookie{
		Name:     auth.FamilyCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(config.Cfg.JWT.ExpiresIn.Seconds()),
	}
	c.SetCookie(accessTokenCookie)


	return response.RespondSuccess(c, http.StatusNoContent, nil)
}

func (h *familyHandler) InviteMembers(c echo.Context) error {
	var req dto.InviteMembersRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body: " + err.Error()})
	}

	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyFamilyID)
	familyID, ok := val.(uuid.UUID)
	if !ok || familyID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid family_id context"})
	}
	req.FamilyID = familyID

	// user_idもcontextから取得
	val = ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}
	req.UserID = userID

	validate := validator.New()
	type emailList struct {
		Emails []string `validate:"required,min=1,dive,email"`
	}
	if err := validate.Struct(emailList{Emails: req.Emails}); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid emails: " + err.Error()})
	}

	err := h.fc.InviteMembers(ctx, &req)
	slog.Debug("InviteMembers: after fc.InviteMembers", "error", err)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return response.RespondSuccess(c, http.StatusNoContent, nil)
}

// ApplyToFamily POST /invitations/apply
func (h *familyHandler) ApplyToFamily(c echo.Context) error {
	var req dto.ApplyRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body: " + err.Error()})
	}
	if req.Token == "" {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "token is required"})
	}

	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	if err := h.fc.ApplyToFamily(ctx, &req, userID); err != nil {
		return errors.RespondWithError(c, err)
	}
	return response.RespondSuccess(c, http.StatusNoContent, nil)
}

// RespondToJoinRequest POST /families/respond
func (h *familyHandler) RespondToJoinRequest(c echo.Context) error {
	var req dto.RespondJoinRequestRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body: " + err.Error()})
	}

	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	if err := h.fc.RespondToJoinRequest(ctx, &req, userID); err != nil {
		return errors.RespondWithError(c, err)
	}
	return response.RespondSuccess(c, http.StatusNoContent, nil)
}

// JoinFamily POST /families/join
func (h *familyHandler) JoinFamily(c echo.Context) error {
	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	token, err := h.fc.JoinFamily(ctx, userID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	accessTokenCookie := &http.Cookie{
		Name:     auth.FamilyCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(config.Cfg.JWT.ExpiresIn.Seconds()),
	}
	c.SetCookie(accessTokenCookie)

	return response.RespondSuccess(c, http.StatusNoContent, nil)
}
