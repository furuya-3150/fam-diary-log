package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
)

type AuthController interface {
	// OAuth2 server-side flow methods
	InitiateGoogleLogin() (authURL string, state string, err error)
	HandleGoogleCallback(ctx context.Context, code string) (isJoined bool, accessToken string, refreshToken string, err error)
	// RefreshToken validates a refresh token and returns a new token pair.
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
}

type authController struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthController(authUsecase usecase.AuthUsecase) AuthController {
	return &authController{
		authUsecase: authUsecase,
	}
}

// InitiateGoogleLogin generates the Google OAuth authorization URL and state
func (c *authController) InitiateGoogleLogin() (string, string, error) {
	return c.authUsecase.InitiateGoogleLogin()
}

// HandleGoogleCallback handles the OAuth callback from Google and returns both tokens.
func (c *authController) HandleGoogleCallback(ctx context.Context, code string) (bool, string, string, error) {
	return c.authUsecase.HandleGoogleCallback(ctx, code)
}

// RefreshToken delegates token refresh to the usecase.
func (c *authController) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	return c.authUsecase.RefreshToken(ctx, refreshToken)
}
