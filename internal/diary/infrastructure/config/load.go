package config

import (
	"os"
	"strconv"
)

func loadDB() DBConfig {
	timeout, _ := strconv.Atoi(os.Getenv("DB_TIMEOUT_SEC"))
	return DBConfig{
		Host:            os.Getenv("DB_HOST"),
		Port:            os.Getenv("DB_PORT"),
		User:            os.Getenv("DB_USER"),
		Password:        os.Getenv("DB_PASSWORD"),
		TimeoutSec:      int64(timeout),
		SSLMode:         os.Getenv("POSTGRES_SSL_MODE"),
		DiaryUser:       os.Getenv("DIARY_DB_USER"),
		DiaryPassword:   os.Getenv("DIARY_DB_PASSWORD"),
		DiaryDBName:     os.Getenv("DIARY_DB_NAME"),
		DiaryTestDBName: os.Getenv("DIARY_TEST_DB_NAME"),
	}
}