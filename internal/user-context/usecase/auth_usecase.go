package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	cfg "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	jwtgen "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/jwt"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/oauth"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/random"
	"github.com/google/uuid"
)

type AuthUsecase interface {
	// OAuth2 server-side flow methods
	InitiateGoogleLogin() (authURL string, state string, err error)
	HandleGoogleCallback(ctx context.Context, code string) (isJoined bool, accessToken string, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error)
}

type authUsecase struct {
	authRepo         repository.UserRepository
	familyMemberRepo repository.FamilyMemberRepository
	refreshTokenRepo repository.RefreshTokenRepository
	googleProvider   oauth.OAuthProvider
	tokenGenerator   jwtgen.TokenGenerator
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	familyMemberRepo repository.FamilyMemberRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	googleProvider oauth.OAuthProvider,
	tokenGenerator jwtgen.TokenGenerator,
) AuthUsecase {
	return &authUsecase{
		authRepo:         userRepo,
		familyMemberRepo: familyMemberRepo,
		refreshTokenRepo: refreshTokenRepo,
		googleProvider:   googleProvider,
		tokenGenerator:   tokenGenerator,
	}
}

// generateAccessToken generates a JWT access token
// func (u *authUsecase) generateAccessToken(user *domain.User) string {
// 	now := time.Now()
// 	expiresAt := now.Add(u.jwtConfig.ExpiresIn)

// 	// Define JWT claims
// 	claims := jwt.MapClaims{
// 		"sub":       user.ID.String(),       // Subject (user ID)
// 		"email":     user.Email,             // User email
// 		"name":      user.Name,              // User name
// 		"provider":  string(user.Provider),  // Auth provider
// 		"iat":       now.Unix(),             // Issued at
// 		"exp":       expiresAt.Unix(),       // Expiration time
// 		"iss":       u.jwtConfig.Issuer,     // Issuer
// 	}

// 	// Create token with claims
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	// Sign token with secret
// 	signedToken, err := token.SignedString([]byte(u.jwtConfig.Secret))
// 	if err != nil {
// 		// Fallback to simple token if JWT generation fails
// 		// In production, this should be logged as a critical error
// 		return fmt.Sprintf("token_%s_%d", user.ID.String(), now.Unix())
// 	}

// 	return signedToken
// }

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
	return authURL + "&prompt=select_account", state, nil
}

var Counter = 0

// HandleGoogleCallback handles the OAuth callback from Google
func (u *authUsecase) HandleGoogleCallback(ctx context.Context, code string) (bool, string, string, error) {
	slog.Debug("HandleGoogleCallback: start", "code", Counter)
	Counter++
	// Exchange code for user info
	userInfo, err := u.googleProvider.ExchangeCode(ctx, code)
	if err != nil {
		slog.Error("failed to exchange Google code", "Error", err)
		return false, "", "", &pkgerrors.ValidationError{Message: fmt.Sprintf("failed to exchange Google code: %v", err)}
	}
	slog.Debug("Google user info fetched", "userInfo", userInfo)

	// Check if user exists by provider ID
	existingUser, err := u.authRepo.GetUserByProviderID(ctx, domain.AuthProviderGoogle, userInfo.ProviderID)
	if err != nil {
		return false, "", "", &pkgerrors.InternalError{Message: "failed to get user"}
	}

	var user *domain.User

	if existingUser != nil {
		// User exists, return existing user
		user = existingUser
	} else {
		// New user, check if email already exists with different provider
		existingEmailUser, err := u.authRepo.GetUserByEmail(ctx, userInfo.Email)
		if err != nil {
			return false, "", "", &pkgerrors.InternalError{Message: "failed to check existing email"}
		}

		if existingEmailUser != nil {
			// Email already exists with different provider
			return false, "", "", &pkgerrors.ValidationError{
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
			return false, "", "", &pkgerrors.InternalError{Message: "failed to create user"}
		}
	}

	member, err := u.familyMemberRepo.GetFamilyMemberByUserID(ctx, user.ID)
	if err != nil {
		return false, "", "", err
	}
	var familyId uuid.UUID
	var isJoined bool
	var role domain.Role
	if member != nil {
		familyId = member.FamilyID
		isJoined = true
		role = member.Role
	}

	// Generate access token
	accessToken, err := u.tokenGenerator.GenerateToken(ctx, user.ID, familyId, role)
	if err != nil {
		return false, "", "", &pkgerrors.InternalError{Message: "failed to generate access token"}
	}

	// Generate and store refresh token
	refreshToken, err := u.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return false, "", "", err
	}

	return isJoined, accessToken, refreshToken, nil
}

// RefreshToken validates the given refresh token and issues a new token pair (rotation).
func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	rt, err := u.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return "", "", &pkgerrors.InternalError{Message: "failed to look up refresh token"}
	}
	if rt == nil || rt.Revoked {
		// TODO: 不正なトークンでDBアクセスした際、管理者に通知する仕組みを入れる（攻撃の可能性があるため）
		return "", "", &pkgerrors.UnauthorizedError{Message: "invalid or revoked refresh token"}
	}
	if time.Now().After(rt.ExpiresAt) {
		return "", "", &pkgerrors.UnauthorizedError{Message: "refresh token has expired"}
	}

	// Load user and family membership
	user, err := u.authRepo.GetUserByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return "", "", &pkgerrors.InternalError{Message: "failed to load user"}
	}

	member, err := u.familyMemberRepo.GetFamilyMemberByUserID(ctx, user.ID)
	if err != nil {
		return "", "", err
	}
	var familyId uuid.UUID
	var role domain.Role
	if member != nil {
		familyId = member.FamilyID
		role = member.Role
	}

	// Issue new access token
	newAccessToken, err := u.tokenGenerator.GenerateToken(ctx, user.ID, familyId, role)
	if err != nil {
		return "", "", &pkgerrors.InternalError{Message: "failed to generate access token"}
	}

	// Revoke old refresh token (token rotation)
	if err := u.refreshTokenRepo.Revoke(ctx, rt.ID); err != nil {
		return "", "", &pkgerrors.InternalError{Message: "failed to revoke old refresh token"}
	}

	// Issue new refresh token
	newRefreshToken, err := u.issueRefreshToken(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

// issueRefreshToken generates a random opaque token, stores it, and returns the token string.
func (u *authUsecase) issueRefreshToken(ctx context.Context, userID uuid.UUID) (string, error) {
	rt, err := GenerateRefresToken()
	rt.UserID = userID
	if err != nil {
		return "", err
	}
	if _, err := u.refreshTokenRepo.Create(ctx, rt); err != nil {
		return "", &pkgerrors.InternalError{Message: "failed to store refresh token"}
	}
	return rt.Token, nil
}

func GenerateRefresToken() (*domain.RefreshToken, error) {
	tokenStr, err := random.GenerateRandomBase64String(48)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	c := cfg.Cfg
	rt := &domain.RefreshToken{
		Token:     tokenStr,
		ExpiresAt: time.Now().Add(c.JWT.RefreshExpiresIn),
	}
	return rt, nil
}
