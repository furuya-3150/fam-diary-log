package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
)

type DiaryAnalysisRepository interface {
	Create(ctx context.Context, analysis *domain.DiaryAnalysis) (*domain.DiaryAnalysis, error)
}

type diaryAnalysisRepository struct {
	dbManager *db.DBManager
}

func NewDiaryAnalysisRepository(dbManager *db.DBManager) DiaryAnalysisRepository {
	return &diaryAnalysisRepository{
		dbManager: dbManager,
	}
}

func (r *diaryAnalysisRepository) Create(ctx context.Context, analysis *domain.DiaryAnalysis) (*domain.DiaryAnalysis, error) {
	result := r.dbManager.DB(ctx).Create(analysis)
	if result.Error != nil {
		return nil, result.Error
	}

	return analysis, nil
}
