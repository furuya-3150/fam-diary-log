package config

// Config holds the application configuration
var Cfg *Config

// Config struct
type Config struct {
	DB DBConfig
	TestDB DBConfig
	ThirdParty ThirdPartyConfig
}

func Load() Config {
	return Config{
		DB:     loadDB(),
		ThirdParty: loadThirdParty(),
	}
}
