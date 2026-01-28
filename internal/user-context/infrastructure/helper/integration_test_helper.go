package helper

import (
	"testing"

	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type IntegrationTestDeps[T any] struct {
    DB     *gorm.DB
    Repo    T
}

func SetupIntegrationTestDeps[T any](t *testing.T, repoCtor func(dm *db.DBManager) T) *IntegrationTestDeps[T] {
    dbConn := SetupTestDB(t)
    return &IntegrationTestDeps[T]{
        DB:  dbConn.GetGorm(),
        Repo: repoCtor(dbConn),
    }
}

func TeardownIntegrationTest[T any](t *testing.T, deps *IntegrationTestDeps[T]) {
    sqlDB, err := deps.DB.DB()
    require.NoError(t, err)
    sqlDB.Close()
}