package config

import (
	"os"
)

type JWTConfig struct {
	Secret     string
}

func loadJWT() JWTConfig {
	return JWTConfig{
		Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
