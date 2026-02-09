package config

type Config struct {
	DB DBConfig
	TestDB DBConfig
	JWT JWTConfig
}

var Cfg Config

func Load() Config {
	return Config{
		DB:     loadDB(),
		JWT:    loadJWT(),
	}
}
