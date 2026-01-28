package handler

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type FamilyHandler interface {
	CreateFamily(c echo.Context) error
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
