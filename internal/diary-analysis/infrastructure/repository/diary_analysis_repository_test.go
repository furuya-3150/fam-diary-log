package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/helper"
)

// List diary analyses with cancelled context
func TestDiaryAnalysisRepository_List_ContextCancelled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	godotenv.Load("../../../../cmd/diary-analysis/.env")

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	userID := uuid.New()
	criteria := &domain.DiaryAnalysisSearchCriteria{
		UserID:  userID,
		Columns: []string{"*"},
	}

	_, err := repo.List(ctx, criteria)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// List diary analyses - success test
func TestDiaryAnalysisRepository_List_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	godotenv.Load("../../../../cmd/diary-analysis/.env")

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()
	baseDate := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	// Create test data
	analysis1 := &domain.DiaryAnalysis{
		ID:            uuid.New(),
		DiaryID:       uuid.New(),
		UserID:        userID,
		FamilyID:      familyID,
		CharCount:     100,
		SentenceCount: 10,
		AccuracyScore: 0.95,
		CreatedAt:     baseDate,
	}

	analysis2 := &domain.DiaryAnalysis{
		ID:            uuid.New(),
		DiaryID:       uuid.New(),
		UserID:        userID,
		FamilyID:      familyID,
		CharCount:     150,
		SentenceCount: 15,
		AccuracyScore: 0.92,
		CreatedAt:     baseDate.AddDate(0, 0, 1),
	}

	// Create records in database
	if err := dbManager.GetGorm().Create(analysis1).Error; err != nil {
		t.Fatalf("failed to create test diary analysis: %v", err)
	}
	if err := dbManager.GetGorm().Create(analysis2).Error; err != nil {
		t.Fatalf("failed to create test diary analysis: %v", err)
	}

	// Create search criteria
	criteria := &domain.DiaryAnalysisSearchCriteria{
		UserID:    userID,
		WeekStart: baseDate.AddDate(0, 0, -1),
		WeekEnd:   baseDate.AddDate(0, 0, 6),
		Columns:   []string{"*"},
	}

	// Call List
	results, err := repo.List(context.Background(), criteria)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify results
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Verify data is sorted by created_at ascending
	if results[0].ID != analysis1.ID {
		t.Errorf("first result ID mismatch: got %v, want %v", results[0].ID, analysis1.ID)
	}
	if results[1].ID != analysis2.ID {
		t.Errorf("second result ID mismatch: got %v, want %v", results[1].ID, analysis2.ID)
	}

	// Verify char counts
	if results[0].CharCount != 100 {
		t.Errorf("first result CharCount mismatch: got %d, want 100", results[0].CharCount)
	}
	if results[1].CharCount != 150 {
		t.Errorf("second result CharCount mismatch: got %d, want 150", results[1].CharCount)
	}
}

// List diary analyses with no matching records
func TestDiaryAnalysisRepository_List_NoRecords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	godotenv.Load("../../../../cmd/diary-analysis/.env")

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	userID := uuid.New()

	// Create search criteria for non-existent user
	criteria := &domain.DiaryAnalysisSearchCriteria{
		UserID:  userID,
		Columns: []string{"*"},
	}

	// Call List
	results, err := repo.List(context.Background(), criteria)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify empty results
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

// List diary analyses with date range filter
func TestDiaryAnalysisRepository_List_WithDateRangeFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	godotenv.Load("../../../../cmd/diary-analysis/.env")

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	userID := uuid.New()
	familyID := uuid.New()
	baseDate := time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC)

	// Create test data with different dates
	analysis1 := &domain.DiaryAnalysis{
		ID:        uuid.New(),
		DiaryID:   uuid.New(),
		UserID:    userID,
		FamilyID:  familyID,
		CharCount: 100,
		CreatedAt: baseDate,
	}

	analysis2 := &domain.DiaryAnalysis{
		ID:        uuid.New(),
		DiaryID:   uuid.New(),
		UserID:    userID,
		FamilyID:  familyID,
		CharCount: 150,
		CreatedAt: baseDate.AddDate(0, 0, 3),
	}

	analysis3 := &domain.DiaryAnalysis{
		ID:        uuid.New(),
		DiaryID:   uuid.New(),
		UserID:    userID,
		FamilyID:  familyID,
		CharCount: 200,
		CreatedAt: baseDate.AddDate(0, 0, 10),
	}

	// Create all records
	for _, analysis := range []*domain.DiaryAnalysis{analysis1, analysis2, analysis3} {
		if err := dbManager.GetGorm().Create(analysis).Error; err != nil {
			t.Fatalf("failed to create test diary analysis: %v", err)
		}
	}

	// Create search criteria with narrow date range
	criteria := &domain.DiaryAnalysisSearchCriteria{
		UserID:    userID,
		WeekStart: baseDate,
		WeekEnd:   baseDate.AddDate(0, 0, 5),
		Columns:   []string{"*"},
	}

	// Call List
	results, err := repo.List(context.Background(), criteria)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify only records within date range are returned
	if len(results) != 2 {
		t.Errorf("expected 2 results within date range, got %d", len(results))
	}

	// Verify the records are the correct ones
	if results[0].CharCount != 100 || results[1].CharCount != 150 {
		t.Errorf("unexpected char counts in filtered results")
	}
}
