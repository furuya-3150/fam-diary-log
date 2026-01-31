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

func setupDepsForJoinRequestTest(t *testing.T) *helper.IntegrationTestDeps[FamilyJoinRequestRepository] {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewFamilyJoinRequestRepository)
	t.Cleanup(func() { helper.TeardownIntegrationTest(t, deps) })
	return deps
}

func TestFamilyJoinRequestRepository_CreateJoinRequest_Success(t *testing.T) {
	deps := setupDepsForJoinRequestTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:        uuid.New(),
		FamilyID:  familyID,
		UserID:    userID,
		Status:    domain.JoinRequestStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := deps.Repo.CreateJoinRequest(ctx, jr)
	require.NoError(t, err)

	got, err := deps.Repo.FindPendingRequest(ctx, familyID, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int(domain.JoinRequestStatusPending), int(got.Status))
}

func TestFamilyJoinRequestRepository_FindPendingRequest_NotFound(t *testing.T) {
	deps := setupDepsForJoinRequestTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()

	got, err := deps.Repo.FindPendingRequest(ctx, familyID, userID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestFamilyJoinRequestRepository_UpdateStatusByID_UpdatesFields(t *testing.T) {
	deps := setupDepsForJoinRequestTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()
	responderID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:        uuid.New(),
		FamilyID:  familyID,
		UserID:    userID,
		Status:    domain.JoinRequestStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, deps.Repo.CreateJoinRequest(ctx, jr))

	updates := map[string]interface{}{
		"status":            int(domain.JoinRequestStatusApproved),
		"responded_user_id": responderID,
		"responded_at":      time.Now(),
		"updated_at":        time.Now(),
	}

	require.NoError(t, deps.Repo.UpdateStatusByID(ctx, jr.ID, updates))

	got, err := deps.Repo.FindByID(ctx, jr.ID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int(domain.JoinRequestStatusApproved), int(got.Status))
	require.NotNil(t, got.RespondedAt)
}

func TestFamilyJoinRequestRepository_FindApprovedByUser_Success(t *testing.T) {
	deps := setupDepsForJoinRequestTest(t)
	ctx := context.Background()

	familyID := uuid.New()
	userID := uuid.New()
	responderID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:              uuid.New(),
		FamilyID:        familyID,
		UserID:          userID,
		Status:          domain.JoinRequestStatusApproved,
		RespondedAt:     func() time.Time { tt := time.Now(); return tt }(),
		RespondedUserID: responderID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	require.NoError(t, deps.Repo.CreateJoinRequest(ctx, jr))

	got, err := deps.Repo.FindApprovedByUser(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, int(domain.JoinRequestStatusApproved), int(got.Status))
}

func TestFamilyJoinRequestRepository_FindApprovedByUser_NotFound(t *testing.T) {
	deps := setupDepsForJoinRequestTest(t)
	ctx := context.Background()

	userID := uuid.New()

	got, err := deps.Repo.FindApprovedByUser(ctx, userID)
	require.NoError(t, err)
	assert.Nil(t, got)
}
