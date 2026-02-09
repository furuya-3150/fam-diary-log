package config

type AppConfig struct {
	URL string
}

func loadApp() AppConfig {
	return AppConfig{
		URL: getEnv("APP_URL", "http://localhost:8082"),
	}
}
