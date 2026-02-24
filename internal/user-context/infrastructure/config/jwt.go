package config

import (
	"os"
	"strconv"
	"time"
)

type JWTConfig struct {
	Secret           string
	ExpiresIn        time.Duration
	RefreshExpiresIn time.Duration
	Issuer           string
	SignMethod       string
}

func loadJWT() JWTConfig {
	expiresInStr := getEnv("JWT_EXPIRES_IN", "3600") // Default: 1 hour
	expiresInSec, err := strconv.Atoi(expiresInStr)
	if err != nil {
		expiresInSec = 3600
	}

	refreshExpiresInStr := getEnv("JWT_REFRESH_EXPIRES_IN", "2592000") // Default: 30 days
	refreshExpiresInSec, err := strconv.Atoi(refreshExpiresInStr)
	if err != nil {
		refreshExpiresInSec = 2592000
	}

	return JWTConfig{
		Secret:           getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ExpiresIn:        time.Duration(expiresInSec) * time.Second,
		RefreshExpiresIn: time.Duration(refreshExpiresInSec) * time.Second,
		Issuer:           getEnv("JWT_ISSUER", "fam-diary-log"),
		SignMethod:       getEnv("JWT_SIGN_METHOD", "HS256"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
