package oauth

import (
	"context"
)

// OAuthUserInfo represents user information from OAuth provider
type OAuthUserInfo struct {
	ProviderID string // User ID from the provider
	Email      string
	Name       string
	Picture    string // Profile picture URL
}

// OAuthProvider defines the interface for OAuth/OpenID Connect providers
type OAuthProvider interface {
	// GetProviderName returns the name of the provider (e.g., "google")
	GetProviderName() string

	// GetAuthURL generates the OAuth authorization URL for the user to visit
	GetAuthURL(state string) string

	// ExchangeCode exchanges the authorization code for an ID token and returns user info
	ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error)
}
