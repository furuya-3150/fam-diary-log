package config

import "os"

func Load() *Config {
	return &Config{
		DB: loadDB(),
		JWT: loadJWT(),
	}
}

func loadDB() DBConfig {
	return DBConfig{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		TestDatabaseURL: os.Getenv("TEST_DATABASE_URL"),
	}
}
