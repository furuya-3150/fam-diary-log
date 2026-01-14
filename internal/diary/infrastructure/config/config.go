package config

type Config struct {
	DB DBConfig
	TestDB DBConfig
}

var Cfg Config

func Load() Config {
	return Config{
		DB:     loadDB(),
	}
}
