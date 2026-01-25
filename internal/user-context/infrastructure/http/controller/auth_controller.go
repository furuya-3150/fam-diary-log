package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
)

type AuthController interface {
	// OAuth2 server-side flow methods
	InitiateGoogleLogin() (authURL string, state string, err error)
	HandleGoogleCallback(ctx context.Context, code string) (*dto.AuthResponse, error)
}

type authController struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthController(authUsecase usecase.AuthUsecase) AuthController {
	return &authController{
		authUsecase: authUsecase,
	}
}

func (c *authController) toAuthResponse(domainResp *domain.AuthResponse) *dto.AuthResponse {
	return &dto.AuthResponse{
		User: &dto.UserResponse{
			ID:        domainResp.User.ID,
			Email:     domainResp.User.Email,
			Name:      domainResp.User.Name,
			Provider:  string(domainResp.User.Provider),
			CreatedAt: domainResp.User.CreatedAt,
		},
		AccessToken: domainResp.AccessToken,
		ExpiresIn:   domainResp.ExpiresIn,
	}
}

// InitiateGoogleLogin generates the Google OAuth authorization URL and state
func (c *authController) InitiateGoogleLogin() (string, string, error) {
	return c.authUsecase.InitiateGoogleLogin()
}

// HandleGoogleCallback handles the OAuth callback from Google
func (c *authController) HandleGoogleCallback(ctx context.Context, code string) (*dto.AuthResponse, error) {
	authResp, err := c.authUsecase.HandleGoogleCallback(ctx, code)
	if err != nil {
		return nil, err
	}
	return c.toAuthResponse(authResp), nil
}
