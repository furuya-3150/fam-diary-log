package repository

import (
	"context"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDepsForTest(t *testing.T) *helper.IntegrationTestDeps[FamilyInvitationRepository] {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyInvitationRepository)
	t.Cleanup(func() { helper.TeardownIntegrationTest(t, deps) })
	return deps
}

func TestFamilyInvitationRepository_CreateInvitation_Success(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	inviterID := uuid.New()
	token := "test-token-create"
	expiresAt := time.Now().Add(24 * time.Hour)

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := deps.Repo.CreateInvitation(ctx, inv)
	require.NoError(t, err)

	got, err := deps.Repo.FindInvitationByFamilyID(ctx, familyID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, token, got.InvitationToken)
}

func TestFamilyInvitationRepository_CreateInvitation_DuplicateToken(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID1 := uuid.New()
	familyID2 := uuid.New()
	inviterID := uuid.New()
	token := "duplicate-token"

	inv1 := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID1,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	inv2 := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID2,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, deps.Repo.CreateInvitation(ctx, inv1))
	err := deps.Repo.CreateInvitation(ctx, inv2)
	require.Error(t, err)
}

func TestFamilyInvitationRepository_FindInvitationByFamilyID_Success(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	inviterID := uuid.New()
	token := "find-token"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, deps.Repo.CreateInvitation(ctx, inv))

	got, err := deps.Repo.FindInvitationByFamilyID(ctx, familyID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, token, got.InvitationToken)
}

func TestFamilyInvitationRepository_FindInvitationByFamilyID_NotFound(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	got, err := deps.Repo.FindInvitationByFamilyID(ctx, familyID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestFamilyInvitationRepository_UpdateInvitationTokenAndExpires_Success(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	inviterID := uuid.New()
	token := "initial-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, deps.Repo.CreateInvitation(ctx, inv))

	newToken := "updated-token"
	newExpires := time.Now().Add(48 * time.Hour)
	require.NoError(t, deps.Repo.UpdateInvitationTokenAndExpires(ctx, familyID, inviterID, newToken, newExpires))

	got, err := deps.Repo.FindInvitationByFamilyID(ctx, familyID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, newToken, got.InvitationToken)
	assert.WithinDuration(t, newExpires, got.ExpiresAt, time.Second)
}

func TestFamilyInvitationRepository_UpdateInvitationTokenAndExpires_NoMatch(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	inviterID := uuid.New()

	// 更新対象が存在しない場合でもエラーにならない実装のため、更新後に取得してnilであることを確認
	err := deps.Repo.UpdateInvitationTokenAndExpires(ctx, familyID, inviterID, "token", time.Now().Add(24*time.Hour))
	require.NoError(t, err)

	got, err := deps.Repo.FindInvitationByFamilyID(ctx, familyID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestFamilyInvitationRepository_FindInvitationByToken_Success(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	inviterID := uuid.New()
	token := "find-by-token"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InviterUserID:   inviterID,
		InvitationToken: token,
		ExpiresAt:       time.Now().Add(24 * time.Hour),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
	require.NoError(t, deps.Repo.CreateInvitation(ctx, inv))

	got, err := deps.Repo.FindInvitationByToken(ctx, token)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, token, got.InvitationToken)
}

func TestFamilyInvitationRepository_FindInvitationByToken_NotFound(t *testing.T) {
	deps := setupDepsForTest(t)
	ctx := context.Background()

	token := "non-existent-token"
	got, err := deps.Repo.FindInvitationByToken(ctx, token)
	require.NoError(t, err)
	assert.Nil(t, got)
}
