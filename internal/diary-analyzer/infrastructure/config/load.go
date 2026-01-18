package config

import "os"

func loadDB() DBConfig {
	return DBConfig{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		TestDatabaseURL: os.Getenv("TEST_DATABASE_URL"),
	}
}

func loadThirdParty() ThirdPartyConfig {
	return ThirdPartyConfig{
		YahooNLPAppID: os.Getenv("YAHOO_NLP_APP_ID"),
	}
}
