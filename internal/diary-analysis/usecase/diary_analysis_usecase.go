package usecase

import (
	"context"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/datetime"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
)

// DiaryAnalysisUsecase defines the interface for diary analysis operations
type DiaryAnalysisUsecase interface {
	GetCharCountByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error)
	GetSentenceCountByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error)
	GetAccuracyScoreByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error)
}

type diaryAnalysisUsecase struct {
	dar repository.DiaryAnalysisRepository
}

// NewDiaryAnalysisUsecase creates a new DiaryAnalysisUsecase instance
func NewDiaryAnalysisUsecase(dar repository.DiaryAnalysisRepository) DiaryAnalysisUsecase {
	return &diaryAnalysisUsecase{
		dar: dar,
	}
}

// getValueByDateCommon is a helper method for retrieving values for each day of the week
func (dau *diaryAnalysisUsecase) getValueByDateCommon(ctx context.Context, userID uuid.UUID, dateStr string, columnName string, getValue func(*domain.DiaryAnalysis) int) (map[string]interface{}, error) {
	// Validate and parse date
	date, err := domain.ValidateYYYYMMDDFormat(dateStr)
	if err != nil {
		return nil, &errors.ValidationError{Message: err.Error()}
	}

	// Validate userID
	if userID == uuid.Nil {
		return nil, &errors.ValidationError{Message: "invalid user ID"}
	}

	// Get week range
	weekStart, weekEnd := datetime.GetWeekRange(date)

	// Create search criteria
	criteria := &domain.DiaryAnalysisSearchCriteria{
		UserID:    userID,
		WeekStart: weekStart,
		WeekEnd:   weekEnd,
		Columns:   []string{"DATE(created_at) as date", columnName},
	}

	analysis, err := dau.dar.List(ctx, criteria)
	if err != nil {
		return nil, err
	}

	resultMap := initializeWeekResultMap(weekStart)

	// Fill in actual values from repository results
	for _, result := range analysis {
		resultMap[result.CreatedAt.Format("2006-01-02")] = getValue(result)
	}

	return resultMap, nil
}

// GetCharCountByDate retrieves character count for each day of the week containing the specified date
func (dau *diaryAnalysisUsecase) GetCharCountByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error) {
	return dau.getValueByDateCommon(ctx, userID, dateStr, "char_count", func(a *domain.DiaryAnalysis) int {
		return a.CharCount
	})
}

// GetSentenceCountByDate retrieves sentence count for each day of the week containing the specified date
func (dau *diaryAnalysisUsecase) GetSentenceCountByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error) {
	return dau.getValueByDateCommon(ctx, userID, dateStr, "sentence_count", func(a *domain.DiaryAnalysis) int {
		return a.SentenceCount
	})
}

// GetAccuracyScoreByDate retrieves accuracy score for each day of the week containing the specified date
func (dau *diaryAnalysisUsecase) GetAccuracyScoreByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error) {
	return dau.getValueByDateCommon(ctx, userID, dateStr, "accuracy_score", func(a *domain.DiaryAnalysis) int {
		return a.AccuracyScore
	})
}

// Build map with all dates of the week, initializing with nil
func initializeWeekResultMap(weekStart time.Time) map[string]interface{} {
	resultMap := make(map[string]interface{})
	for i := 0; i < 7; i++ {
		currentDate := weekStart.AddDate(0, 0, i)
		dateStr := currentDate.Format("2006-01-02")
		resultMap[dateStr] = nil
	}
	return resultMap
}
