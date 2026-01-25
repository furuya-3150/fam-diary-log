package config

const (
	SessionName = "oauth_session"
	SessionKeyOAuthState = "oauth_state"
)

type SessionConfig struct {
	Secret string
	MaxAge int
}
