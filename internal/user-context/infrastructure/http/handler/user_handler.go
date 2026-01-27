package handler

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/labstack/echo/v4"
)

type UserHandler interface {
	EditProfile(c echo.Context) error
	GetProfile(c echo.Context) error
}

type userHandler struct {
	userController controller.UserController
}

func NewUserHandler(userController controller.UserController) UserHandler {
	return &userHandler{userController: userController}
}

// EditProfile PUT /users/me

func (h *userHandler) EditProfile(c echo.Context) error {
	var req controller_dto.EditUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body" + err.Error()})
	}
	user, err := h.userController.EditProfile(c.Request().Context(), &req)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}

// GetProfile GET /users/me
func (h *userHandler) GetProfile(c echo.Context) error {
	ctx := c.Request().Context()
	val := ctx.Value("user_id")
	userID, ok := val.(string)
	if !ok || userID == "" {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request"})
	}
	user, err := h.userController.GetProfile(ctx, userID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return c.JSON(http.StatusOK, user)
}
