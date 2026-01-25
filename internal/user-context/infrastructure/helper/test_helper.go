package helper

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/joho/godotenv"
)

// SetupTestDB sets up test database and returns DBManager
func SetupTestDB(t *testing.T) *db.DBManager {
	godotenv.Load("../../../../cmd/user-context/.env")
	t.Helper()
	cfg := config.Load()
	fmt.Println(cfg.DB.TestDatabaseURL)
	dbManager := db.NewDBManager(cfg.DB.TestDatabaseURL)

	// cleanup
	if err := dbManager.GetGorm().Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("failed to clean up test database: %v", err)
	}

	return dbManager
}

// TeardownTestDB cleans up test database
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	t.Helper()
	if err := gormDB.Exec("DELETE FROM users").Error; err != nil {
		t.Logf("warning: failed to cleanup test database: %v", err)
	}
}
