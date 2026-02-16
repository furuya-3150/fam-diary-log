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

func setupFamilyMemberDepsForTest(t *testing.T) *helper.IntegrationTestDeps[FamilyMemberRepository] {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyMemberRepository)
	t.Cleanup(func() { helper.TeardownIntegrationTest(t, deps) })
	return deps
}

func TestFamilyMemberRepository_AddFamilyMember_Success(t *testing.T) {
	deps := setupFamilyMemberDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()

	member := &domain.FamilyMember{
		ID:       uuid.New(),
		FamilyID: familyID,
		UserID:   userID,
		Role:     domain.RoleMember,
	}

	err := deps.Repo.AddFamilyMember(ctx, member)
	require.NoError(t, err)

	// 追加されたメンバーを取得して確認
	got, err := deps.Repo.GetFamilyMemberByUserID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, familyID, got.FamilyID)
	assert.Equal(t, domain.RoleMember, got.Role)
}

func TestFamilyMemberRepository_IsUserAlreadyMember_True(t *testing.T) {
	deps := setupFamilyMemberDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()

	member := &domain.FamilyMember{
		ID:       uuid.New(),
		FamilyID: familyID,
		UserID:   userID,
		Role:     domain.RoleAdmin,
	}

	err := deps.Repo.AddFamilyMember(ctx, member)
	require.NoError(t, err)

	// メンバーであることを確認
	isMember, err := deps.Repo.IsUserAlreadyMember(ctx, userID)
	require.NoError(t, err)
	assert.True(t, isMember)
}

func TestFamilyMemberRepository_IsUserAlreadyMember_False(t *testing.T) {
	deps := setupFamilyMemberDepsForTest(t)
	ctx := context.Background()

	userID := uuid.New()

	// メンバーでないことを確認
	isMember, err := deps.Repo.IsUserAlreadyMember(ctx, userID)
	require.NoError(t, err)
	assert.False(t, isMember)
}

func TestFamilyMemberRepository_GetFamilyMemberByUserID_Success(t *testing.T) {
	deps := setupFamilyMemberDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()

	member := &domain.FamilyMember{
		ID:       uuid.New(),
		FamilyID: familyID,
		UserID:   userID,
		Role:     domain.RoleAdmin,
	}

	err := deps.Repo.AddFamilyMember(ctx, member)
	require.NoError(t, err)

	got, err := deps.Repo.GetFamilyMemberByUserID(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, userID, got.UserID)
	assert.Equal(t, familyID, got.FamilyID)
	assert.Equal(t, domain.RoleAdmin, got.Role)
}

func TestFamilyMemberRepository_GetFamilyMemberByUserID_NotFound(t *testing.T) {
	deps := setupFamilyMemberDepsForTest(t)
	ctx := context.Background()

	userID := uuid.New()

	got, err := deps.Repo.GetFamilyMemberByUserID(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, got)
}