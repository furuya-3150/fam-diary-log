package repository
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/helper"
)

// diary creation with cancelled context test
func TestDiaryRepository_Create_ContextCancelled(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	diary := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "Test content",
	}

	_, err := repo.Create(ctx, diary)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// diary creation with timed out context test
func TestDiaryRepository_Create_WithTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// .envでtimeoutを5秒で設定している
	time.Sleep(time.Duration(5+1) * time.Second) // タイムアウトを待つ

	diary := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "Test content",
	}

	_, err := repo.Create(ctx, diary)
	if err == nil {
		t.Fatal("expected error for timed out context")
	}
}

// diary creation success test
func TestDiaryRepository_Create_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	diary := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	result, err := repo.Create(context.Background(), diary)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.ID != diaryID {
		t.Errorf("ID mismatch: got %v, want %v", result.ID, diaryID)
	}

	if result.Title != diary.Title {
		t.Errorf("Title mismatch: got %q, want %q", result.Title, diary.Title)
	}

	if result.Content != diary.Content {
		t.Errorf("Content mismatch: got %q, want %q", result.Content, diary.Content)
	}

	var saved domain.Diary
	if err := dbManager.GetGorm().First(&saved, "id = ?", diaryID).Error; err != nil {
		t.Fatalf("failed to retrieve saved diary: %v", err)
	}

	if saved.Title != diary.Title {
		t.Errorf("saved title mismatch: got %q, want %q", saved.Title, diary.Title)
	}
}

// multiple records creation test
func TestDiaryRepository_Create_MultipleRecords(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	familyID := uuid.New()
	diaries := []*domain.Diary{
		{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			FamilyID: familyID,
			Title:    "First Diary",
			Content:  "First content",
		},
		{
			ID:       uuid.New(),
			UserID:   uuid.New(),
			FamilyID: familyID,
			Title:    "Second Diary",
			Content:  "Second content",
		},
	}

	for _, diary := range diaries {
		_, err := repo.Create(context.Background(), diary)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	var count int64
	if err := dbManager.GetGorm().Model(&domain.Diary{}).Where("family_id = ?", familyID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count diaries: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 diaries, got %d", count)
	}
}

// ------------
// List Diaries
// ------------

// diary list retrieval with date range and boundary check
func TestDiaryRepository_List_SuccessWithDateRange(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	familyID, userID := uuid.New(), uuid.New()
	repo := NewDiaryRepository(dbManager)

	// Setup date range
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), now.Day()-(int(now.Weekday())-1), 0, 0, 0, 0, now.Location())
	endDate := startDate.AddDate(0, 0, 6).Add(time.Duration(23*3600+59*60+59)*time.Second + 999999999*time.Nanosecond)

	// Test cases: (title, offset from startDate)
	testCases := []struct {
		title  string
		offset time.Duration
		want   bool
	}{
		{"Before Start Date", -24 * time.Hour, false},
		{"At Start Date", 0, true},
		{"Inside Range", 3 * 24 * time.Hour, true},
		{"At End Date", endDate.Sub(startDate), true},
		{"After End Date", endDate.Sub(startDate) + 24*time.Hour, false},
	}

	// Create and insert test diaries
	for _, tc := range testCases {
		diary := &domain.Diary{
			ID:       uuid.New(),
			UserID:   userID,
			FamilyID: familyID,
			Title:    tc.title,
			Content:  tc.title,
		}
		if _, err := repo.Create(context.Background(), diary); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if err := dbManager.GetGorm().Model(diary).Update("created_at", startDate.Add(tc.offset)).Error; err != nil {
			t.Fatalf("failed to update created_at: %v", err)
		}
	}

	// Fetch results
	result, err := repo.List(context.Background(), &domain.DiarySearchCriteria{
		FamilyID:  familyID,
		StartDate: startDate,
		EndDate:   endDate,
	}, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify results
	if result == nil || len(result) != 3 {
		t.Errorf("expected 3 diaries in range, got %d", len(result))
	}

	resultTitles := make(map[string]bool)
	for _, d := range result {
		resultTitles[d.Title] = true
	}

	for _, tc := range testCases {
		if tc.want != resultTitles[tc.title] {
			if tc.want {
				t.Errorf("expected %q to be included in results", tc.title)
			} else {
				t.Errorf("expected %q to be excluded from results", tc.title)
			}
		}
	}

	// Verify DESC order by created_at
	if len(result) >= 2 && result[0].CreatedAt.Before(result[1].CreatedAt) {
		t.Error("expected results ordered by created_at DESC")
	}
}

// diary list retrieval for different family
func TestDiaryRepository_List_DifferentFamily(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	familyID1 := uuid.New()
	familyID2 := uuid.New()
	userID := uuid.New()

	// Create diaries for two different families
	diary1 := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID1,
		Title:    "Family 1 Diary",
		Content:  "Content for family 1",
	}

	diary2 := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID2,
		Title:    "Family 2 Diary",
		Content:  "Content for family 2",
	}

	_, err1 := repo.Create(context.Background(), diary1)
	_, err2 := repo.Create(context.Background(), diary2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Create failed: %v, %v", err1, err2)
	}

	// List diaries for family 1
	criteria := &domain.DiarySearchCriteria{
		FamilyID: familyID1,
	}

	result, err := repo.List(context.Background(), criteria, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result) != 1 {
		t.Errorf("expected 1 diary for family 1, got %d", len(result))
	}

	if len(result) > 0 && result[0].FamilyID != familyID1 {
		t.Errorf("expected family ID %v, got %v", familyID1, result[0].FamilyID)
	}
}

// diary list retrieval with empty result
func TestDiaryRepository_List_EmptyResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryRepository(dbManager)

	familyID := uuid.New()

	// List diaries for family with no diaries
	criteria := &domain.DiarySearchCriteria{
		FamilyID: familyID,
	}

	result, err := repo.List(context.Background(), criteria, nil)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil for empty list")
	}

	if len(result) != 0 {
		t.Errorf("expected 0 diaries, got %d", len(result))
	}
}
