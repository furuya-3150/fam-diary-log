package handler

import (
	"net/http"
	"strings"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/furuya-3150/fam-diary-log/pkg/response"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler interface {
	EditProfile(c echo.Context) error
	GetProfile(c echo.Context) error
	GetFamilyMembers(c echo.Context) error
}

type userHandler struct {
	userController controller.UserController
	userUsecase    usecase.UserUsecase
}

func NewUserHandler(userController controller.UserController, userUsecase usecase.UserUsecase) UserHandler {
	return &userHandler{
		userController: userController,
		userUsecase:    userUsecase,
	}
}

// EditProfile PUT /users/me

func (h *userHandler) EditProfile(c echo.Context) error {
	var req controller_dto.EditUserRequest
	if err := c.Bind(&req); err != nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request body" + err.Error()})
	}
	userId := c.Request().Context().Value(auth.ContextKeyUserID).(uuid.UUID)
	req.ID = userId

	user, err := h.userController.EditProfile(c.Request().Context(), &req)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return response.RespondSuccess(c, http.StatusOK, user)
}

// GetProfile GET /users/me
func (h *userHandler) GetProfile(c echo.Context) error {
	ctx := c.Request().Context()
	val := ctx.Value(auth.ContextKeyUserID)
	userID, ok := val.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "invalid request"})
	}
	user, err := h.userController.GetProfile(ctx, userID)
	if err != nil {
		return errors.RespondWithError(c, err)
	}
	return response.RespondSuccess(c, http.StatusOK, user)
}

// GetFamilyMembers GET /families/me/members
func (h *userHandler) GetFamilyMembers(c echo.Context) error {
	ctx := c.Request().Context()

	// familyIDをコンテキストから取得
	val := ctx.Value(auth.ContextKeyFamilyID)
	familyID, ok := val.(uuid.UUID)
	if !ok || familyID == uuid.Nil {
		return errors.RespondWithError(c, &errors.BadRequestError{Message: "family_id is required"})
	}

	// クエリパラメータからfieldsを取得
	fieldsParam := c.QueryParam("fields")
	var fields []string
	if fieldsParam != "" {
		// カンマ区切りで分割し、トリムする
		for _, field := range strings.Split(fieldsParam, ",") {
			trimmed := strings.TrimSpace(field)
			if trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
	}

	// usecaseから取得（バリデーションとデフォルト値設定はusecase内で実施）
	users, err := h.userUsecase.GetFamilyMembers(ctx, familyID, fields)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	return response.RespondSuccess(c, http.StatusOK, users)
}
