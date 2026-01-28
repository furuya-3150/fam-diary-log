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

func TestFamilyRepositoryIntegration_CreateFamily_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyRepository)
	defer helper.TeardownIntegrationTest(t, deps)

	ctx := context.Background()
	family := &domain.Family{
		ID:        uuid.New(),
		Name:      "TestFamily",
	}
	result, err := deps.Repo.CreateFamily(ctx, family)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, family.ID, result.ID)
	assert.Equal(t, family.Name, result.Name)
}

func TestFamilyRepositoryIntegration_CreateFamily_DuplicateID(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyRepository)
	defer helper.TeardownIntegrationTest(t, deps)

	ctx := context.Background()
	family := &domain.Family{
		ID:        uuid.New(),
		Name:      "TestFamily",
	}
	_, err := deps.Repo.CreateFamily(ctx, family)
	require.NoError(t, err)
	// 異常系: 同じIDで再作成（主キー重複）
	_, err = deps.Repo.CreateFamily(ctx, family)
	assert.Error(t, err)
}

func TestFamilyMemberRepositoryIntegration_AddFamilyMember_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyMemberRepository)
	defer helper.TeardownIntegrationTest(t, deps)

	ctx := context.Background()
	familyID := uuid.New()
	userID := uuid.New()
	member := &domain.FamilyMember{
		FamilyID:  familyID,
		UserID:    userID,
		Role:      domain.RoleAdmin,
	}
	err := deps.Repo.AddFamilyMember(ctx, member)
	require.NoError(t, err)

	var persisted domain.FamilyMember
	err = deps.DB.First(&persisted, "family_id = ? AND user_id = ?", familyID, userID).Error
	require.NoError(t, err)
	assert.Equal(t, familyID, persisted.FamilyID)
	assert.Equal(t, userID, persisted.UserID)
	assert.Equal(t, domain.RoleAdmin, persisted.Role)
}

func TestFamilyMemberRepositoryIntegration_AddFamilyMember_Duplicate(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyMemberRepository)
	defer helper.TeardownIntegrationTest(t, deps)

	ctx := context.Background()
	familyID := uuid.New()
	userID := uuid.New()
	member := &domain.FamilyMember{
		FamilyID:  familyID,
		UserID:    userID,
		Role:      domain.RoleAdmin,
	}
	err := deps.Repo.AddFamilyMember(ctx, member)
	require.NoError(t, err)
	// 異常系: 同じfamily_id, user_idで再作成（UNIQUE制約違反）
	err = deps.Repo.AddFamilyMember(ctx, member)
	assert.Error(t, err)
}