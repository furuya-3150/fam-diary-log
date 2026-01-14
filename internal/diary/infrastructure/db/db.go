package db

import (
	"context"
	"fmt"
	"log"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/config"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	txKey = "tx"
)

type DBManager struct {
	db *gorm.DB
}

func (dm *DBManager) DB(ctx context.Context) *gorm.DB {
	tx, ok := ctx.Value(txKey).(*gorm.DB)
	if !ok {
		return dm.db.WithContext(ctx)
	}

	return tx.WithContext(ctx)
}

func (m *DBManager) WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

func NewDBManger() *DBManager {
	err := godotenv.Load()
	if err != nil {
		log.Fatalln(err)
	}
	url := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?connect_timeout=%s&sslmode=%s", config.Cfg.DB.DiaryUser,
		config.Cfg.DB.DiaryPassword, config.Cfg.DB.Host, config.Cfg.DB.Port,
		config.Cfg.DB.DiaryDBName,
		config.Cfg.DB.TimeoutSec, config.Cfg.DB.SSLMode)
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		log.Fatalln(err)
	}

	// Pingで接続確認
	if _, err := db.DB(); err != nil {
		log.Fatalf("DB ping failed: %v", err)
	}

	fmt.Println("DB Connceted")
	return &DBManager{db}
}

// NewTestDBManager creates a DBManager for testing with provided GORM DB instance
// This allows tests to use actual database connections while maintaining DBManager interface
func NewTestDBManager(testDB *gorm.DB) *DBManager {
	return &DBManager{db: testDB}
}

func NewDB() *gorm.DB {
	dbManger := NewDBManger()
	return dbManger.db
}

func CloseDB(db *gorm.DB) {
	sqlDB, _ := db.DB()
	if err := sqlDB.Close(); err != nil {
		log.Fatalln(err)
	}
}
