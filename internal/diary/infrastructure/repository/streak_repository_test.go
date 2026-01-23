package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/helper"
)

// streak creation with cancelled context test
func TestStreakRepository_CreateOrUpdate_ContextCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	streak := &domain.Streak{
		UserID:        uuid.New(),
		FamilyID:      uuid.New(),
		CurrentStreak: 5,
		LastPostDate:  nil,
	}

	_, err := repo.CreateOrUpdate(ctx, streak)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// streak creation success test
func TestStreakRepository_CreateOrUpdate_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()
	now := time.Now().Truncate(time.Hour)
	lastPostDate := &now

	streak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 5,
		LastPostDate:  lastPostDate,
	}

	result, err := repo.CreateOrUpdate(context.Background(), streak)
	if err != nil {
		t.Fatalf("CreateOrUpdate failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", result.UserID, userID)
	}

	if result.FamilyID != familyID {
		t.Errorf("FamilyID mismatch: got %v, want %v", result.FamilyID, familyID)
	}

	if result.CurrentStreak != 5 {
		t.Errorf("CurrentStreak mismatch: got %d, want %d", result.CurrentStreak, 5)
	}

	var saved domain.Streak
	if err := dbManager.GetGorm().First(&saved, "user_id = ? AND family_id = ?", userID, familyID).Error; err != nil {
		t.Fatalf("failed to retrieve saved streak: %v", err)
	}

	if saved.CurrentStreak != 5 {
		t.Errorf("saved current_streak mismatch: got %d, want %d", saved.CurrentStreak, 5)
	}
}

// streak update (upsert) test
func TestStreakRepository_CreateOrUpdate_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()

	// Create initial streak
	initialStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 3,
		LastPostDate:  nil,
	}

	_, err := repo.CreateOrUpdate(context.Background(), initialStreak)
	if err != nil {
		t.Fatalf("CreateOrUpdate failed: %v", err)
	}

	// Update streak
	now := time.Now().Truncate(time.Hour)
	updatedStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 10,
		LastPostDate:  &now,
	}

	result, err := repo.CreateOrUpdate(context.Background(), updatedStreak)
	if err != nil {
		t.Fatalf("CreateOrUpdate failed: %v", err)
	}

	if result.CurrentStreak != 10 {
		t.Errorf("CurrentStreak mismatch after update: got %d, want %d", result.CurrentStreak, 10)
	}

	var saved domain.Streak
	if err := dbManager.GetGorm().First(&saved, "user_id = ? AND family_id = ?", userID, familyID).Error; err != nil {
		t.Fatalf("failed to retrieve updated streak: %v", err)
	}

	if saved.CurrentStreak != 10 {
		t.Errorf("saved current_streak mismatch after update: got %d, want %d", saved.CurrentStreak, 10)
	}

	// Verify only one record exists (upsert, not insert)
	var count int64
	if err := dbManager.GetGorm().Model(&domain.Streak{}).Where("user_id = ? AND family_id = ?", userID, familyID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count streaks: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 streak record, got %d", count)
	}
}

// streak get success test
func TestStreakRepository_Get_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()
	now := time.Now().Truncate(time.Hour)

	streak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 7,
		LastPostDate:  &now,
	}

	_, err := repo.CreateOrUpdate(context.Background(), streak)
	if err != nil {
		t.Fatalf("CreateOrUpdate failed: %v", err)
	}

	result, err := repo.Get(context.Background(), userID, familyID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.UserID != userID {
		t.Errorf("UserID mismatch: got %v, want %v", result.UserID, userID)
	}

	if result.FamilyID != familyID {
		t.Errorf("FamilyID mismatch: got %v, want %v", result.FamilyID, familyID)
	}

	if result.CurrentStreak != 7 {
		t.Errorf("CurrentStreak mismatch: got %d, want %d", result.CurrentStreak, 7)
	}
}

// streak get with non-existent record test (should return nil, nil)
func TestStreakRepository_Get_NotFound_NoError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()

	result, err := repo.Get(context.Background(), userID, familyID)

	// Should return (nil, nil) for non-existent record
	if err != nil {
		t.Fatalf("Get should not return error for non-existent record, got: %v", err)
	}

	if result != nil {
		t.Errorf("expected nil result for non-existent record, got %v", result)
	}
}

// streak get with different user/family combination test
func TestStreakRepository_Get_DifferentUserFamilyCombinations(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewStreakRepository(dbManager)

	userID1 := uuid.New()
	userID2 := uuid.New()
	familyID1 := uuid.New()
	familyID2 := uuid.New()

	// Create streaks for different combinations
	testCases := []struct {
		userID   uuid.UUID
		familyID uuid.UUID
		streak   int
	}{
		{userID1, familyID1, 5},
		{userID1, familyID2, 3},
		{userID2, familyID1, 7},
		{userID2, familyID2, 2},
	}

	for _, tc := range testCases {
		streak := &domain.Streak{
			UserID:        tc.userID,
			FamilyID:      tc.familyID,
			CurrentStreak: tc.streak,
		}
		_, err := repo.CreateOrUpdate(context.Background(), streak)
		if err != nil {
			t.Fatalf("CreateOrUpdate failed: %v", err)
		}
	}

	// Verify each combination retrieves the correct streak
	for _, tc := range testCases {
		result, err := repo.Get(context.Background(), tc.userID, tc.familyID)
		if err != nil && err != gorm.ErrRecordNotFound {
			t.Fatalf("Get failed for user %v, family %v: %v", tc.userID, tc.familyID, err)
		}

		if result == nil {
			t.Fatalf("expected result for user %v, family %v, got nil", tc.userID, tc.familyID)
		}

		if result.CurrentStreak != tc.streak {
			t.Errorf("CurrentStreak mismatch for user %v, family %v: got %d, want %d", tc.userID, tc.familyID, result.CurrentStreak, tc.streak)
		}
	}
}
