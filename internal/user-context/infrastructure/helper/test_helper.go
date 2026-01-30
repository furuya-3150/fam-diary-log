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

	tables := []string{"users", "families", "family_members", "family_invitations", "family_join_requests"}
	for _, table := range tables {
		if err := dbManager.GetGorm().Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			t.Logf("warning: failed to cleanup table %s: %v", table, err)
		}
	}

	return dbManager
}

// TeardownTestDB cleans up test database
func TeardownTestDB(t *testing.T, gormDB *gorm.DB) {
	t.Helper()
	tables := []string{"users", "families", "family_members", "family_invitations", "family_join_requests"}
	for _, table := range tables {
		if err := gormDB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			t.Logf("warning: failed to cleanup table %s: %v", table, err)
		}
	}
}
