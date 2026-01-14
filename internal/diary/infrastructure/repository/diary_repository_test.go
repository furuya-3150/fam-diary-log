package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/helper"
)

// diary creation with cancelled context test
func TestDiaryRepository_Create_ContextCancelled(t *testing.T) {

	if testing.Short() {
		t.Skip("skipping integration test")
	}

	gormDB, dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, gormDB)

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

	gormDB, dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, gormDB)

	repo := NewDiaryRepository(dbManager)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(time.Duration(config.Cfg.DB.TimeoutSec+1) * time.Second) // タイムアウトを待つ

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

	gormDB, dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, gormDB)

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
	if err := gormDB.First(&saved, "id = ?", diaryID).Error; err != nil {
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

	gormDB, dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, gormDB)

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
	if err := gormDB.Model(&domain.Diary{}).Where("family_id = ?", familyID).Count(&count).Error; err != nil {
		t.Fatalf("failed to count diaries: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 diaries, got %d", count)
	}
}
