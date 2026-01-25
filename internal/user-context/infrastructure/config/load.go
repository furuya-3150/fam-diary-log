package config

import (
	"os"
	"strconv"
)

func loadDB() DBConfig {
	return DBConfig{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		TestDatabaseURL: os.Getenv("TEST_DATABASE_URL"),
	}
}

func loadOAuth() OAuthConfig {
	return OAuthConfig{
		Google: GoogleOAuthConfig{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		},
	}
}

func loadSession() SessionConfig {
	maxAge := 86400 * 7 // Default: 7 days
	if val := os.Getenv("SESSION_MAX_AGE"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			maxAge = parsed
		}
	}

	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "default-secret-change-in-production" // Fallback (開発用のみ)
	}

	return SessionConfig{
		Secret: secret,
		MaxAge: maxAge,
	}
}
