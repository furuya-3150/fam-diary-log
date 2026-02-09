
package config

import (
	"os"
	"strconv"
	"time"
)

type JWTConfig struct {
	Secret     string
	ExpiresIn  time.Duration
	Issuer     string
	SignMethod string
}

func loadJWT() JWTConfig {
	expiresInStr := getEnv("JWT_EXPIRES_IN", "3600") // Default: 1 hour
	expiresInSec, err := strconv.Atoi(expiresInStr)
	if err != nil {
		expiresInSec = 3600
	}

	return JWTConfig{
		Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		ExpiresIn:  time.Duration(expiresInSec) * time.Second,
		Issuer:     getEnv("JWT_ISSUER", "fam-diary-log"),
		SignMethod: getEnv("JWT_SIGN_METHOD", "HS256"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
