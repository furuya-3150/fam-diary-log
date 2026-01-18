package usecase

import (
	"context"
	"log/slog"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
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
	tm        db.TransactionManager
	dr        repository.DiaryRepository
	publisher publisher.Publisher
	clk       clock.Clock
}

func NewDiaryUsecase(tm db.TransactionManager, dr repository.DiaryRepository) DiaryUsecase {
	return &diaryUsecase{
		tm:        tm,
		dr:        dr,
		publisher: nil,
		clk:       &clock.Real{},
	}
}

func NewDiaryUsecaseWithPublisher(tm db.TransactionManager, dr repository.DiaryRepository, pub publisher.Publisher) DiaryUsecase {
	return &diaryUsecase{
		tm:        tm,
		dr:        dr,
		publisher: pub,
		clk:       &clock.Real{},
	}
}

func NewDiaryUsecaseWithClock(tm db.TransactionManager, dr repository.DiaryRepository, clk clock.Clock) DiaryUsecase {
	return &diaryUsecase{
		tm:        tm,
		dr:        dr,
		publisher: nil,
		clk:       clk,
	}
}

func NewDiaryUsecaseWithPublisherAndClock(tm db.TransactionManager, dr repository.DiaryRepository, pub publisher.Publisher, clk clock.Clock) DiaryUsecase {
	return &diaryUsecase{
		tm:        tm,
		dr:        dr,
		publisher: pub,
		clk:       clk,
	}
}

func (du *diaryUsecase) Create(ctx context.Context, d *domain.Diary) (*domain.Diary, error) {
	err := domain.ValidateCreateDiaryRequest(d)
	if err != nil {
		return nil, &errors.ValidationError{Message: err.Error()}
	}
	if du.publisher == nil {
		return nil, &errors.LogicError{Message: "publisher is not set"}
	}

	diary, err := du.dr.Create(ctx, d)
	if err != nil {
		return nil, err
	}

	// Publish diary created event
	event := domain.NewDiaryCreatedEvent(diary.ID, diary.UserID, diary.FamilyID, diary.Content)
	if err := du.publisher.Publish(ctx, event); err != nil {
		slog.Error("failed to publish diary created event", "error", err.Error())
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
