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
        ID:         uuid.New(),
        FamilyID:   familyID,
        UserID:     userID,
        Status:     domain.JoinRequestStatusPending,
        CreatedAt:  time.Now(),
        UpdatedAt:  time.Now(),
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
