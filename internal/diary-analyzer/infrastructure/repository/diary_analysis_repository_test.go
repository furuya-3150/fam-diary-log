package repository

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/helper"
)

// TestDiaryAnalysisRepositoryCreateSuccess tests successful creation
func TestDiaryAnalysisRepositoryCreateSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	// Arrange
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	analysis := &domain.DiaryAnalysis{
		DiaryID:       diaryID,
		UserID:        userID,
		FamilyID:      familyID,
		CharCount:     150,
		SentenceCount: 5,
		AccuracyScore: 85,
	}

	// Act
	result, err := repo.Create(context.Background(), analysis)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEqual(t, uuid.Nil, result.ID)
	assert.Equal(t, diaryID, result.DiaryID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, familyID, result.FamilyID)
	assert.Equal(t, 150, result.CharCount)
	assert.Equal(t, 5, result.SentenceCount)
	assert.Equal(t, 85, result.AccuracyScore)
}

// TestDiaryAnalysisRepositoryCreateContextCanceled tests context cancellation handling
func TestDiaryAnalysisRepositoryCreateContextCanceled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database test in short mode")
	}

	// Arrange
	dbManager := helper.SetupTestDB(t)
	defer helper.TeardownTestDB(t, dbManager.GetGorm())

	repo := NewDiaryAnalysisRepository(dbManager)

	analysis := &domain.DiaryAnalysis{
		DiaryID:       uuid.New(),
		UserID:        uuid.New(),
		FamilyID:      uuid.New(),
		CharCount:     150,
		SentenceCount: 5,
		AccuracyScore: 85,
	}

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Act
	result, err := repo.Create(ctx, analysis)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}
