package config

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	TimeoutSec      int64
	SSLMode         string

	DiaryUser       string
	DiaryPassword   string
	DiaryDBName     string
	DiaryTestDBName string
}
