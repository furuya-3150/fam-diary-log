package usecase

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/db"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/datetime"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
)

type DiaryUsecase interface {
	Create(ctx context.Context, d *domain.Diary) (*domain.Diary, error)
	List(ctx context.Context, familyID uuid.UUID) ([]*domain.Diary, error)
}

type diaryUsecase struct {
	tm  db.TransactionManager
	dr  repository.DiaryRepository
	clk clock.Clock
}

func NewDiaryUsecase(tm db.TransactionManager, dr repository.DiaryRepository) DiaryUsecase {
	return &diaryUsecase{
		tm:  tm,
		dr:  dr,
		clk: &clock.Real{},
	}
}

func NewDiaryUsecaseWithClock(tm db.TransactionManager, dr repository.DiaryRepository, clk clock.Clock) DiaryUsecase {
	return &diaryUsecase{
		tm:  tm,
		dr:  dr,
		clk: clk,
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

func (du *diaryUsecase) List(ctx context.Context, familyID uuid.UUID) ([]*domain.Diary, error) {
	// Clock を使用（テストでモック化可能）
	targetDate := du.clk.Now()

	// その週の開始日（月曜日）と終了日（日曜日）を計算
	weekStart, weekEnd := datetime.GetWeekRange(targetDate)

	// DiarySearchCriteria を構築
	query := &domain.DiarySearchCriteria{
		FamilyID:  familyID,
		StartDate: weekStart,
		EndDate:   weekEnd,
	}

	diaries, err := du.dr.List(ctx, query, nil)
	if err != nil {
		return nil, err
	}

	return diaries, nil
}
