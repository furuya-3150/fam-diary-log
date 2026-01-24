package helper

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/joho/godotenv"
)

// setup
func SetupTestDB(t *testing.T) (*db.DBManager) {
	if os.Getenv("GO_ENV") == "dev" {
		_ = godotenv.Load("../../.env")
	}
	config := config.Load()
	fmt.Println(config.DB.TestDatabaseURL)
	dbManager := db.NewDBManager(config.DB.TestDatabaseURL)

	// cleanup
	if err := dbManager.GetGorm().Exec("DELETE FROM diary_analyses").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}

	return dbManager
}

// teardown
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	if err := gormDB.Exec("DELETE FROM diary_analyses").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
}
