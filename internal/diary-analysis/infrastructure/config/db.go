package config

type DBConfig struct {
	DatabaseURL     string
	TestDatabaseURL string
}

type Config struct {
	DB DBConfig
}
