package helper

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	// "github.com/joho/godotenv"
)

// setup
func SetupTestDB(t *testing.T) (*db.DBManager) {
	t.Helper()
	config := config.Load()
	fmt.Println(config.DB.TestDatabaseURL)
	dbManager := db.NewDBManager(config.DB.TestDatabaseURL)

	// cleanup
	if err := dbManager.GetGorm().Exec("DELETE FROM diaries").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}
	if err := dbManager.GetGorm().Exec("DELETE FROM streaks").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}

	return dbManager
}

// teardown
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	t.Helper()
	if err := gormDB.Exec("DELETE FROM diaries").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
	if err := gormDB.Exec("DELETE FROM streaks").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
}