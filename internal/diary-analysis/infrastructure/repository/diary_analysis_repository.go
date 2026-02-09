package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analysis/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"gorm.io/gorm"
)

type DiaryAnalysisRepository interface {
	List(ctx context.Context, criteria *domain.DiaryAnalysisSearchCriteria) ([]*domain.DiaryAnalysis, error)
}

type diaryAnalysisRepository struct {
	dm *db.DBManager
}

func NewDiaryAnalysisRepository(dm *db.DBManager) DiaryAnalysisRepository {
	return &diaryAnalysisRepository{
		dm: dm,
	}
}

// List retrieves diary analyses based on the search criteria
func (dar *diaryAnalysisRepository) List(ctx context.Context, criteria *domain.DiaryAnalysisSearchCriteria) ([]*domain.DiaryAnalysis, error) {
	db := dar.dm.DB(ctx)

	var diaryAnalysis []*domain.DiaryAnalysis

	q := db.Model(&domain.DiaryAnalysis{}).
		Where("user_id = ?", criteria.UserID)

	if !criteria.WeekStart.IsZero() {
		q = q.Where("DATE(created_at) >= ?", criteria.WeekStart)
	}

	if !criteria.WeekEnd.IsZero() {
		q = q.Where("DATE(created_at) <= ?", criteria.WeekEnd)
	}

	// Apply columns selection if specified
	if len(criteria.Columns) > 0 {
		q = q.Select(criteria.Columns)
	}

	err := q.
		Order("created_at ASC").
		Find(&diaryAnalysis).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return []*domain.DiaryAnalysis{}, nil
		}
		return nil, err
	}

	return diaryAnalysis, nil
}
