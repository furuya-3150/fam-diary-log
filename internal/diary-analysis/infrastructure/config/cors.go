package config

import "strings"

type CORSConfig struct {
	AllowedOrigins []string
}

func loadCORS() CORSConfig {
	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000")
	origins := strings.Split(originsStr, ",")
	// Trim spaces
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	return CORSConfig{
		AllowedOrigins: origins,
	}
}
