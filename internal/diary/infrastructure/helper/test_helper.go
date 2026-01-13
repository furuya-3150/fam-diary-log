package helper

import (
	"fmt"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/db"
)

const (
	DB_USER        = "postgres"
	DB_PASSWORD    = "password"
	DB_NAME        = "test_diary"
	DB_HOST        = "db"
	DB_PORT        = "5432"
	DB_TIMEOUT_SEC = "5"
	DB_SSLMODE     = "disable"
)

// setup
func SetupTestDB(t *testing.T) (*gorm.DB, *db.DBManager) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?connect_timeout=%s&sslmode=%s",
		DB_USER,
		DB_PASSWORD,
		DB_HOST,
		DB_PORT,
		DB_NAME,
		DB_TIMEOUT_SEC,
		DB_SSLMODE,
	)
	fmt.Println(dsn)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// cleanup
	if err := gormDB.Exec("DELETE FROM diaries").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}

	dbManager := db.NewTestDBManager(gormDB)

	return gormDB, dbManager
}

// teardown
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	if err := gormDB.Exec("DELETE FROM diaries").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
}
