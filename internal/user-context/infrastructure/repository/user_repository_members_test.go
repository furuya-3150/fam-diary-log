package repository

import (
	"context"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	integrationTestSkipMsg = "Integration test - requires database setup"
	insertFamilyMemberSQL  = "INSERT INTO family_members (id, family_id, user_id, role, created_at, updated_at) VALUES (?, ?, ?, ?, NOW(), NOW())"
)

func TestUserRepositoryGetUsersByFamilyIDSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipMsg)
	}
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	ctx := context.Background()
	repo := NewUserRepository(dbManager)

	// テストデータの準備
	familyID := uuid.New()
	user1ID := uuid.New()
	user2ID := uuid.New()

	// ユーザーを作成
	user1 := &domain.User{
		ID:         user1ID,
		Email:      "user1@example.com",
		Name:       "User One",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google_user1",
	}
	user2 := &domain.User{
		ID:         user2ID,
		Email:      "user2@example.com",
		Name:       "User Two",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google_user2",
	}

	_, err := repo.CreateUser(ctx, user1)
	require.NoError(t, err)
	_, err = repo.CreateUser(ctx, user2)
	require.NoError(t, err)

	// family_membersテーブルに挿入（直接SQLで）
	db := dbManager.GetGorm()
	err = db.Exec("INSERT INTO families (id, name, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", familyID, "Test Family").Error
	require.NoError(t, err)
	err = db.Exec(insertFamilyMemberSQL, uuid.New(), familyID, user1ID, domain.RoleAdmin).Error
	require.NoError(t, err)
	err = db.Exec(insertFamilyMemberSQL, uuid.New(), familyID, user2ID, domain.RoleMember).Error
	require.NoError(t, err)

	// 全フィールドを指定して取得
	fields := []string{"id", "email", "name", "provider", "provider_id", "created_at", "updated_at"}
	users, err := repo.GetUsersByFamilyID(ctx, familyID, fields)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, user1ID, users[0].ID)
	assert.Equal(t, "User One", users[0].Name)
}

func TestUserRepositoryGetUsersByFamilyIDWithFieldSelection(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipMsg)
	}
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	ctx := context.Background()
	repo := NewUserRepository(dbManager)

	// テストデータの準備
	familyID := uuid.New()
	userID := uuid.New()

	user := &domain.User{
		ID:         userID,
		Email:      "user@example.com",
		Name:       "Test User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google_user",
	}

	_, err := repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// family_membersテーブルに挿入
	db := dbManager.GetGorm()
	err = db.Exec("INSERT INTO families (id, name, created_at, updated_at) VALUES (?, ?, NOW(), NOW())", familyID, "Test Family").Error
	require.NoError(t, err)
	err = db.Exec(insertFamilyMemberSQL, uuid.New(), familyID, userID, domain.RoleAdmin).Error
	require.NoError(t, err)

	// フィールド指定ありで取得
	fields := []string{"id", "name"}
	users, err := repo.GetUsersByFamilyID(ctx, familyID, fields)
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, userID, users[0].ID)
	assert.Equal(t, "Test User", users[0].Name)
}

func TestUserRepositoryGetUsersByFamilyIDNoMembers(t *testing.T) {
	if testing.Short() {
		t.Skip(integrationTestSkipMsg)
	}
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	ctx := context.Background()
	repo := NewUserRepository(dbManager)

	// 存在しないfamilyIDで取得
	familyID := uuid.New()
	fields := []string{"id", "email", "name"}
	users, err := repo.GetUsersByFamilyID(ctx, familyID, fields)
	require.NoError(t, err)
	assert.Len(t, users, 0)
}
