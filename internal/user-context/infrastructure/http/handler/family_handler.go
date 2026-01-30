package handler

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FamilyHandler interface {
	CreateFamily(c echo.Context) error
	InviteMembers(c echo.Context) error
	ApplyToFamily(c echo.Context) error
}

type familyHandler struct {
	familyController controller.FamilyController
}

func NewFamilyHandler(familyController controller.FamilyController) FamilyHandler {
	return &familyHandler{familyController: familyController}
}

// CreateFamily POST /families
func (h *familyHandler) CreateFamily(c echo.Context) error {
	var req dto.CreateFamilyRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body" + err.Error()})
	}

	ctx := c.Request().Context()
	val := ctx.Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request"})
	}

	family, err := h.familyController.CreateFamily(ctx, &req, userID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	return c.JSON(http.StatusOK, family)
}

func (h *familyHandler) InviteMembers(c echo.Context) error {
	var req dto.InviteMembersRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body: " + err.Error()})
	}

	ctx := c.Request().Context()
	val := ctx.Value("family_id")
	familyID, ok := val.(uuid.UUID)
	if !ok || familyID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid family_id context"})
	}
	req.FamilyID = familyID

	// user_idもcontextから取得
	val = ctx.Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	validate := validator.New()
	type emailList struct {
		Emails []string `validate:"required,min=1,dive,email"`
	}
	if err := validate.Struct(emailList{Emails: req.Emails}); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid emails: " + err.Error()})
	}

	err := h.familyController.InviteMembers(ctx, &req)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
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
	val := ctx.Value("user_id")
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid user_id context"})
	}

	if err := h.familyController.ApplyToFamily(ctx, &req, userID); err != nil {
		return errors.RespondWithError(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
