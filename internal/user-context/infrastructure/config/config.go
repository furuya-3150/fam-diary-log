package config

type Config struct {
	DB      DBConfig
	OAuth   OAuthConfig
	Session SessionConfig
	JWT     JWTConfig
	App     AppConfig
}

var Cfg Config

func Load() Config {
	return Config{
		DB:      loadDB(),
		OAuth:   loadOAuth(),
		Session: loadSession(),
		JWT:     loadJWT(),
		App:     loadApp(),
	}
}
