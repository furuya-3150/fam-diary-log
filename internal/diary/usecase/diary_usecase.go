package usecase

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/db"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
)

type DiaryUsecase interface {
	Create(ctx context.Context, d *domain.Diary) (*domain.Diary,error)
	List(ctx context.Context, query *domain.DiarySearchCriteria) ([]*domain.Diary, error)
}

type diaryUsecase struct {
	tm db.TransactionManager
	dr repository.DiaryRepository
}

func NewDiaryUsecase(tm db.TransactionManager, dr repository.DiaryRepository) DiaryUsecase {
	return &diaryUsecase{
		tm: tm,
		dr: dr,
	}
}

func (du *diaryUsecase) Create(ctx context.Context, d *domain.Diary) (*domain.Diary, error) {
	err := domain.ValidateCreateDiaryRequest(d)
	if err != nil {
		return nil, &errors.ValidationError{Message: err.Error()}
	}

	diary, err := du.dr.Create(ctx, d)
	if err != nil {
		return nil, err
	}

	return diary, nil
}

func (du *diaryUsecase) List(ctx context.Context, query *domain.DiarySearchCriteria) ([]*domain.Diary, error) {
	// return du.unitOfWork.DiaryRepository().List(ctx, query.FamilyID, query.StartDate, query.EndDate)
	return []*domain.Diary{}, nil
}
