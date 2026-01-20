package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDiaryAnalysisRepository struct {
	mock.Mock
}

func (m *MockDiaryAnalysisRepository) List(ctx context.Context, criteria *domain.DiaryAnalysisSearchCriteria) ([]*domain.DiaryAnalysis, error) {
	args := m.Called(ctx, criteria)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.DiaryAnalysis), args.Error(1)
}

// GetCharCountByDate with valid date - success
func TestDiaryAnalysisUsecase_GetCharCountByDate_Success(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	dateStr := "2026-01-20"

	createdAt, _ := time.Parse("2006-01-02", dateStr)

	// Mock repository results
	mockResults := []*domain.DiaryAnalysis{
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 10,
			CreatedAt: createdAt.AddDate(0, 0, -1),
		},
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 20,
			CreatedAt: createdAt,
		},
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 30,
			CreatedAt: time.Now().AddDate(0, 0, 1),
		},
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 50,
			CreatedAt: time.Now().AddDate(0, 0, 3),
		},
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 60,
			CreatedAt: time.Now().AddDate(0, 0, 4),
		},
		{
			ID:        uuid.New(),
			DiaryID:   uuid.New(),
			UserID:    userID,
			FamilyID:  uuid.New(),
			CharCount: 70,
			CreatedAt: time.Now().AddDate(0, 0, 5),
		},
	}

	mockRepository.On("List", mock.Anything, mock.MatchedBy(func(criteria *domain.DiaryAnalysisSearchCriteria) bool {
		return criteria.UserID == userID
	})).Return(mockResults, nil)

	// Call GetCharCountByDate
	actual, err := usecase.GetCharCountByDate(context.Background(), userID, dateStr)
	if err != nil {
		t.Fatalf("GetCharCountByDate failed: %v", err)
	}

	// Verify result
	expected := map[string]interface{}{
		"2026-01-19": 10,
		"2026-01-20": 20,
		"2026-01-21": 30,
		"2026-01-22": nil, // No entry for this date
		"2026-01-23": 50,
		"2026-01-24": 60,
		"2026-01-25": 70,
	}
	assert.Equal(t, expected, actual)

	// Verify mock was called
	mockRepository.AssertCalled(t, "List", mock.Anything, mock.Anything)
}

// GetCharCountByDate with invalid date format
func TestDiaryAnalysisUsecase_GetCharCountByDate_InvalidDate(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	invalidDate := "invalid-date"

	// Call GetCharCountByDate with invalid date
	_, err := usecase.GetCharCountByDate(context.Background(), userID, invalidDate)
	if err == nil {
		t.Fatalf("expected error for invalid date format")
	}

	// Verify error type
	_, ok := err.(*errors.ValidationError)
	assert.True(t, ok, "expected ValidationError")
}

// GetCharCountByDate with nil user ID
func TestDiaryAnalysisUsecase_GetCharCountByDate_NilUserID(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	dateStr := "2026-01-20"

	// Call GetCharCountByDate with nil user ID
	_, err := usecase.GetCharCountByDate(context.Background(), uuid.Nil, dateStr)
	if err == nil {
		t.Fatalf("expected error for nil userID")
	}

	// Verify error type
	_, ok := err.(*errors.ValidationError)
	assert.True(t, ok, "expected ValidationError")
}

// GetCharCountByDate with no results from repository
func TestDiaryAnalysisUsecase_GetCharCountByDate_NoResults(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	dateStr := "2026-01-20"

	// Mock empty repository results
	mockRepository.On("List", mock.Anything, mock.Anything).Return([]*domain.DiaryAnalysis{}, nil)

	// Call GetCharCountByDate
	actual, err := usecase.GetCharCountByDate(context.Background(), userID, dateStr)
	if err != nil {
		t.Fatalf("GetCharCountByDate failed: %v", err)
	}

	// Verify result is 0
	expected := map[string]interface{}{
		"2026-01-19": nil,
		"2026-01-20": nil,
		"2026-01-21": nil,
		"2026-01-22": nil,
		"2026-01-23": nil,
		"2026-01-24": nil,
		"2026-01-25": nil,
	}
	assert.Equal(t, expected, actual)
}

// GetCharCountByDate with repository error
func TestDiaryAnalysisUsecase_GetCharCountByDate_RepositoryError(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	dateStr := "2026-01-20"

	// Mock repository error
	mockRepository.On("List", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	// Call GetCharCountByDate
	_, err := usecase.GetCharCountByDate(context.Background(), userID, dateStr)
	if err == nil {
		t.Fatalf("expected error from repository")
	}

	// Verify error
	assert.Error(t, err)
}

// GetCharCountByDate verifies the correct date range is queried
func TestDiaryAnalysisUsecase_GetCharCountByDate_DateRange(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	dateStr := "2026-01-20" // Monday

	mockRepository.On("List", mock.Anything, mock.MatchedBy(func(criteria *domain.DiaryAnalysisSearchCriteria) bool {
		// Verify the week range is correct (Monday to Sunday)
		// For 2026-01-20 (Monday), should query from 2026-01-19 to 2026-01-25
		return criteria.UserID == userID &&
			!criteria.WeekStart.IsZero() &&
			!criteria.WeekEnd.IsZero()
	})).Return([]*domain.DiaryAnalysis{}, nil)

	// Call GetCharCountByDate
	_, err := usecase.GetCharCountByDate(context.Background(), userID, dateStr)
	if err != nil {
		t.Fatalf("GetCharCountByDate failed: %v", err)
	}

	// Verify mock was called with correct criteria
	mockRepository.AssertCalled(t, "List", mock.Anything, mock.Anything)
}

// GetAccuracyScoreByDate with valid date - success
func TestDiaryAnalysisUsecase_GetAccuracyScoreByDate_Success(t *testing.T) {
	t.Parallel()

	mockRepository := new(MockDiaryAnalysisRepository)
	usecase := NewDiaryAnalysisUsecase(mockRepository)

	userID := uuid.New()
	dateStr := "2026-01-20"

	createdAt, _ := time.Parse("2006-01-02", dateStr)

	// Mock repository results
	mockResults := []*domain.DiaryAnalysis{
		{
			ID:            uuid.New(),
			DiaryID:       uuid.New(),
			UserID:        userID,
			FamilyID:      uuid.New(),
			AccuracyScore: 50,
			CreatedAt:     createdAt,
		},
		{
			ID:            uuid.New(),
			DiaryID:       uuid.New(),
			UserID:        userID,
			FamilyID:      uuid.New(),
			AccuracyScore: 60,
			CreatedAt:     createdAt.AddDate(0, 0, 1),
		},
	}

	mockRepository.On("List", mock.Anything, mock.Anything).Return(mockResults, nil)

	// Call GetAccuracyScoreByDate
	result, err := usecase.GetAccuracyScoreByDate(context.Background(), userID, dateStr)
	if err != nil {
		t.Fatalf("GetAccuracyScoreByDate failed: %v", err)
	}

	// Verify results
	assert.NotNil(t, result, "result should not be nil")
	assert.Equal(t, 50, result[createdAt.Format("2006-01-02")], "accuracy score should match")
	assert.Equal(t, 60, result[createdAt.AddDate(0, 0, 1).Format("2006-01-02")], "accuracy score should match")
}
