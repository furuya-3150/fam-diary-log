package config

type AppConfig struct {
	URL string
}

func loadApp() AppConfig {
	return AppConfig{
		URL: getEnv("APP_URL", "http://localhost:8082"),
	}
}

type ClientAppConfig struct {
	URL string
}

func loadClientApp() ClientAppConfig {
	return ClientAppConfig{
		URL: getEnv("CLIENT_APP_URL", "http://localhost:3000"),
	}
}