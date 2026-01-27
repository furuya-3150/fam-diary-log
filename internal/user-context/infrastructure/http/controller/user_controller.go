package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
)

type UserController interface {
	EditProfile(ctx context.Context, req *dto.EditUserRequest) (*dto.UserResponse, error)
	GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error)
}

func (c *userController) GetProfile(ctx context.Context, userID string) (*dto.UserResponse, error) {
	user, err := c.usecase.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

type userController struct {
	// 依存するusecase等をここに追加
	usecase usecase.UserUsecase
}

func NewUserController(usecase usecase.UserUsecase) UserController {
	return &userController{usecase: usecase}
}

func (c *userController) EditProfile(ctx context.Context, req *dto.EditUserRequest) (*dto.UserResponse, error) {
	input := &usecase.EditUserInput{
		ID:    req.ID.String(),
		Name:  req.Name,
		Email: req.Email,
	}
	user, err := c.usecase.EditUser(ctx, input)
	if err != nil {
		return nil, err
	}
	return &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Provider:  string(user.Provider),
		CreatedAt: user.CreatedAt,
	}, nil
}
