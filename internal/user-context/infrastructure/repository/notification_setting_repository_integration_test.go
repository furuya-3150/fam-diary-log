package repository

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/helper"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDepsForNotificationSettingTest(t *testing.T) *helper.IntegrationTestDeps[NotificationSettingRepository] {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}
	deps := helper.SetupIntegrationTestDeps(t, NewNotificationSettingRepository)
	// ensure cleanup of notification_settings table after test
	t.Cleanup(func() {
		deps.DB.Exec("DELETE FROM notification_settings")
		helper.TeardownIntegrationTest(t, deps)
	})
	return deps
}

func TestNotificationSettingRepository_UpsertAndGet_Success(t *testing.T) {
	deps := setupDepsForNotificationSettingTest(t)
	ctx := context.Background()

	userID := uuid.New()
	familyID := uuid.New()

	ns := &domain.NotificationSetting{
		ID:                 uuid.New(),
		UserID:             userID,
		FamilyID:           familyID,
		PostCreatedEnabled: true,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	created, err := deps.Repo.Upsert(ctx, ns)
	require.NoError(t, err)
	require.NotNil(t, created)

	got, err := deps.Repo.GetByUserAndFamily(ctx, userID, familyID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, true, got.PostCreatedEnabled)
}

func TestNotificationSettingRepository_Upsert_UpdateExisting(t *testing.T) {
	deps := setupDepsForNotificationSettingTest(t)
	ctx := context.Background()

	userID := uuid.New()
	familyID := uuid.New()

	ns := &domain.NotificationSetting{
		UserID:             userID,
		FamilyID:           familyID,
		PostCreatedEnabled: true,
	}

	_, err := deps.Repo.Upsert(ctx, ns)
	require.NoError(t, err)

	// update
	ns.PostCreatedEnabled = false
	log.Println("ns1", ns)
	_, err = deps.Repo.Upsert(ctx, ns)
	require.NoError(t, err)

	log.Println("ns2", ns)

	got, err := deps.Repo.GetByUserAndFamily(ctx, userID, familyID)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, false, got.PostCreatedEnabled)
}

func TestNotificationSettingRepository_GetByUserAndFamily_NotFound(t *testing.T) {
	deps := setupDepsForNotificationSettingTest(t)
	ctx := context.Background()

	userID := uuid.New()
	familyID := uuid.New()

	got, err := deps.Repo.GetByUserAndFamily(ctx, userID, familyID)
	require.NoError(t, err)
	assert.Nil(t, got)
}
