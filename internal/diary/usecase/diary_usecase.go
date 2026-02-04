package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/datetime"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/pagination"
	"github.com/furuya-3150/fam-diary-log/pkg/validation"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DiaryUsecase interface {
	Create(ctx context.Context, d *domain.Diary) (*domain.Diary, error)
	List(ctx context.Context, familyID uuid.UUID, targetDate string) ([]*domain.Diary, error)
	GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error)
	GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error)
}

type diaryUsecase struct {
	tm        db.TransactionManager
	dr        repository.DiaryRepository
	sr        repository.StreakRepository
	publisher publisher.Publisher
	clk       clock.Clock
}

// NewDiaryUsecase creates a new DiaryUsecase with all dependencies injected
func NewDiaryUsecase(tm db.TransactionManager, dr repository.DiaryRepository, sr repository.StreakRepository, pub publisher.Publisher, clk clock.Clock) DiaryUsecase {
	return &diaryUsecase{
		tm:        tm,
		dr:        dr,
		sr:        sr,
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

	now := du.clk.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	query := &domain.DiarySearchCriteria{
		FamilyID:  d.FamilyID,
		UserID:    d.UserID,
		StartDate: startOfDay,
		EndDate:   endOfDay,
	}
	pagination := &pagination.Pagination{
		Limit: 1,
	}
	// Check if a diary has already been posted today
	if todaysDiaries, err := du.dr.List(ctx, query, pagination); err != nil {
		return nil, err
	} else if len(todaysDiaries) > 0 {
		return nil, &errors.ValidationError{Message: "diary already posted today"}
	}

	// assign ID if not provided
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}

	du.tm.BeginTx(ctx)

	diary, err := du.dr.Create(ctx, d)
	if err != nil {
		du.tm.RollbackTx(ctx)
		return nil, err
	}

	// Create or update streak
	err = du.updateStreak(ctx, d.UserID, d.FamilyID)
	if err != nil {
		du.tm.RollbackTx(ctx)
		slog.Error("failed to update streak", "error", err.Error())
		// Don't return error, continue with diary creation
		return nil, err
	}

	// Publish diary created event
	event := domain.NewDiaryCreatedEvent(diary.ID, diary.UserID, diary.FamilyID, diary.Content)
	if err := du.publisher.Publish(ctx, event); err != nil {
		du.tm.RollbackTx(ctx)
		slog.Error("failed to publish diary created event", "error", err.Error())
		return nil, err
	}
	defer du.publisher.Close()

	du.tm.CommitTx(ctx)

	return diary, nil
}

func (du *diaryUsecase) updateStreak(ctx context.Context, userID, familyID uuid.UUID) error {
	todayDate := du.clk.Now().Truncate(24 * time.Hour)

	// Get existing streak
	existingStreak, err := du.sr.Get(ctx, userID, familyID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	var currentStreak int = domain.DefaultStreakValue
	var lastPostDate *time.Time = &todayDate

	if existingStreak != nil && existingStreak.LastPostDate != nil {
		lastPostDateTrunc := existingStreak.LastPostDate.Truncate(24 * time.Hour)
		yesterday := todayDate.AddDate(0, 0, -1)

		if lastPostDateTrunc.Equal(yesterday) {
			// Consecutive post: increment streak
			currentStreak = existingStreak.CurrentStreak + 1
		} else if lastPostDateTrunc.Equal(todayDate) {
			return &errors.LogicError{Message: "diary already posted today"}
		}
	}

	streak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: currentStreak,
		LastPostDate:  lastPostDate,
	}

	_, err = du.sr.CreateOrUpdate(ctx, streak)
	if err != nil {
		return err
	}

	return nil
}

func (du *diaryUsecase) List(ctx context.Context, familyID uuid.UUID, targetDate string) ([]*domain.Diary, error) {
	var query *domain.DiarySearchCriteria
	parsedDate, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return nil, &errors.ValidationError{Message: "target_date must be in YYYY-MM-DD format"}
	}
	weekStart, weekEnd := datetime.GetWeekRange(parsedDate)
	query = &domain.DiarySearchCriteria{
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

func (du *diaryUsecase) GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error) {
	// Validate and parse year and month
	_, _, err := validation.ValidateYearMonth(year, month)
	if err != nil {
		return 0, &errors.ValidationError{Message: err.Error()}
	}

	// Combine year and month in YYYY-MM format
	yearMonth := year + "-" + month

	criteria := &domain.DiaryCountCriteria{
		FamilyID:  familyID,
		YearMonth: yearMonth,
	}

	count, err := du.dr.GetCount(ctx, criteria)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (du *diaryUsecase) GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error) {
	// Validate userID
	if userID == uuid.Nil {
		return nil, &errors.ValidationError{Message: "invalid user ID"}
	}

	// Validate familyID
	if familyID == uuid.Nil {
		return nil, &errors.ValidationError{Message: "invalid family ID"}
	}

	streak, err := du.sr.Get(ctx, userID, familyID)
	if err != nil {
		return nil, err
	}

	return streak, nil
}
