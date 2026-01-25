package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleProvider struct {
	clientID     string
	clientSecret string
	redirectURL  string
	oauth2Config *oauth2.Config
}

// NewGoogleProviderWithOAuth2 creates a GoogleProvider with full OAuth2 support
func NewGoogleProviderWithOAuth2(clientID, clientSecret, redirectURL string) OAuthProvider {
	oauth2Config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"openid",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleProvider{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURL:  redirectURL,
		oauth2Config: oauth2Config,
	}
}

func (p *GoogleProvider) GetProviderName() string {
	return "google"
}

// GetAuthURL generates the OAuth authorization URL
func (p *GoogleProvider) GetAuthURL(state string) string {
	if p.oauth2Config == nil {
		return ""
	}
	return p.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCode exchanges the authorization code for an access token and retrieves user info
func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error) {
	if p.oauth2Config == nil {
		return nil, &pkgerrors.InternalError{Message: "OAuth2 config not initialized"}
	}

	// Exchange code for token
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return nil, &pkgerrors.ExternalAPIError{
			Message: "failed to exchange authorization code with Google",
			Cause:   err,
		}
	}

	// Get user info from Google's userinfo endpoint
	// Note: Using oauth2.Config.Client() automatically adds authentication headers
	client := p.oauth2Config.Client(ctx, token)
	resp, err := client.Get(config.GoogleUserInfoURL)
	if err != nil {
		return nil, &pkgerrors.ExternalAPIError{
			Message: "failed to get user info from Google",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &pkgerrors.ExternalAPIError{
			Message: fmt.Sprintf("Google API returned status %d", resp.StatusCode),
		}
	}

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to decode user info from Google"}
	}

	return &OAuthUserInfo{
		ProviderID: userInfo.ID,
		Email:      userInfo.Email,
		Name:       userInfo.Name,
		Picture:    userInfo.Picture,
	}, nil
}
