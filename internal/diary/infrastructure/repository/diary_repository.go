package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/pagination"
)

type DiaryRepository interface {
	Create(ctx context.Context, diary *domain.Diary) (*domain.Diary, error)
	List(ctx context.Context, criteria *domain.DiarySearchCriteria, pag *pagination.Pagination) ([]*domain.Diary, error)
}

type diaryRepository struct {
	dm *db.DBManager
}

func NewDiaryRepository(dm *db.DBManager) DiaryRepository {
	return &diaryRepository{
		dm: dm,
	}
}

func (dr *diaryRepository) Create(ctx context.Context, diary *domain.Diary) (*domain.Diary, error) {
	db := dr.dm.DB(ctx)
	err := db.Create(diary).Error
	if err != nil {
		return nil, err
	}
	return diary, nil
}

func (dr *diaryRepository) List(ctx context.Context, criteria *domain.DiarySearchCriteria, pag *pagination.Pagination) ([]*domain.Diary, error) {
	db := dr.dm.DB(ctx)
	var diaries []*domain.Diary

	q := db.Where("family_id = ?", criteria.FamilyID)

	if !criteria.StartDate.IsZero() {
		q = q.Where("created_at >= ?", criteria.StartDate)
	}

	if !criteria.EndDate.IsZero() {
		q = q.Where("created_at <= ?", criteria.EndDate)
	}

	if pag != nil {
		if pag.Limit > 0 {
			q = q.Limit(pag.Limit)
		}
		if pag.Offset > 0 {
			q = q.Offset(pag.Offset)
		}
	}

	// created_at で降順ソート
	err := q.Order("created_at DESC").Find(&diaries).Error
	if err != nil {
		return nil, err
	}
	return diaries, nil
}
