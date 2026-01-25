package config

const (
	// GoogleUserInfoURL is the endpoint to fetch Google user information
	GoogleUserInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type OAuthConfig struct {
	Google GoogleOAuthConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}
