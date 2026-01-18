package helper

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
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
func SetupTestDB(t *testing.T) (*db.DBManager) {
	config := config.Load()
	fmt.Println(config.DB.TestDatabaseURL)
	dbManager := db.NewDBManager(config.DB.TestDatabaseURL)

	// cleanup
	if err := dbManager.GetGorm().Exec("DELETE FROM diaries").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}

	return dbManager
}

// teardown
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	if err := gormDB.Exec("DELETE FROM diaries").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
}
