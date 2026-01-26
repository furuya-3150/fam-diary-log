package handler

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/labstack/echo/v4"
)

type UserHandler interface {
	// EditUser(c echo.Context) error
	EditProfile(c echo.Context) error
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
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	user, err := h.userController.EditProfile(c.Request().Context(), &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, user)
}
