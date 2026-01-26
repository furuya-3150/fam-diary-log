package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/oauth"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/random"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	// OAuth2 server-side flow methods
	InitiateGoogleLogin() (authURL string, state string, err error)
	HandleGoogleCallback(ctx context.Context, code string) (*domain.AuthResponse, error)
}

type authUsecase struct {
	authRepo       repository.UserRepository
	googleProvider oauth.OAuthProvider
	jwtConfig      config.JWTConfig
}

func NewAuthUsecase(userRepo repository.UserRepository, googleProvider oauth.OAuthProvider) AuthUsecase {
	cfg := config.Load()
	return &authUsecase{
		authRepo:       userRepo,
		googleProvider: googleProvider,
		jwtConfig:      cfg.JWT,
	}
}

// generateAccessToken generates a JWT access token
func (u *authUsecase) generateAccessToken(user *domain.User) string {
	now := time.Now()
	expiresAt := now.Add(u.jwtConfig.ExpiresIn)

	// Define JWT claims
	claims := jwt.MapClaims{
		"sub":       user.ID.String(),       // Subject (user ID)
		"email":     user.Email,             // User email
		"name":      user.Name,              // User name
		"provider":  string(user.Provider),  // Auth provider
		"iat":       now.Unix(),             // Issued at
		"exp":       expiresAt.Unix(),       // Expiration time
		"iss":       u.jwtConfig.Issuer,     // Issuer
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	signedToken, err := token.SignedString([]byte(u.jwtConfig.Secret))
	if err != nil {
		// Fallback to simple token if JWT generation fails
		// In production, this should be logged as a critical error
		return fmt.Sprintf("token_%s_%d", user.ID.String(), now.Unix())
	}

	return signedToken
}

// InitiateGoogleLogin generates the Google OAuth authorization URL and state
func (u *authUsecase) InitiateGoogleLogin() (string, string, error) {
	// Generate state for CSRF protection (32 bytes of randomness)
	state, err := random.GenerateRandomBase64String(32)
	if err != nil {
		return "", "", &pkgerrors.InternalError{Message: "failed to generate state"}
	}

	authURL := u.googleProvider.GetAuthURL(state)
	if authURL == "" {
		return "", "", &pkgerrors.InternalError{Message: "Google OAuth not properly configured"}
	}
	return authURL, state, nil
}

// HandleGoogleCallback handles the OAuth callback from Google
func (u *authUsecase) HandleGoogleCallback(ctx context.Context, code string) (*domain.AuthResponse, error) {
	// Exchange code for user info
	userInfo, err := u.googleProvider.ExchangeCode(ctx, code)
	if err != nil {
		return nil, &pkgerrors.ValidationError{Message: fmt.Sprintf("failed to exchange Google code: %v", err)}
	}

	// Check if user exists by provider ID
	existingUser, err := u.authRepo.GetUserByProviderID(ctx, domain.AuthProviderGoogle, userInfo.ProviderID)
	
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to get user"}
	}

	var user *domain.User

	if existingUser != nil {
		// User exists, return existing user
		user = existingUser
	} else {
		// New user, check if email already exists with different provider
		existingEmailUser, err := u.authRepo.GetUserByEmail(ctx, userInfo.Email)
		if err != nil {
			return nil, &pkgerrors.InternalError{Message: "failed to check existing email"}
		}

		if existingEmailUser != nil {
			// Email already exists with different provider
			return nil, &pkgerrors.ValidationError{
				Message: fmt.Sprintf("email %s is already registered with %s", userInfo.Email, existingEmailUser.Provider),
			}
		}

		// Create new user
		user = &domain.User{
			ID:         uuid.New(),
			Email:      userInfo.Email,
			Name:       userInfo.Name,
			Provider:   domain.AuthProviderGoogle,
			ProviderID: userInfo.ProviderID,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		user, err = u.authRepo.CreateUser(ctx, user)
		if err != nil {
			return nil, &pkgerrors.InternalError{Message: "failed to create user"}
		}
	}

	// Generate access token
	accessToken := u.generateAccessToken(user)

	return &domain.AuthResponse{
		User:        user,
		AccessToken: accessToken,
		ExpiresIn:   3600, // 1 hour
	}, nil
}
