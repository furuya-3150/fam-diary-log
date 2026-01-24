package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/diary/usecase"
	"github.com/google/uuid"
)

type DiaryController interface {
	Create(ctx context.Context, d *domain.Diary) (*dto.DiaryResponse, error)
	List(ctx context.Context, familyID uuid.UUID) ([]dto.DiaryResponse, error)
	GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error)
	GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*dto.StreakResponse, error)
}

type diaryController struct {
	du usecase.DiaryUsecase
}

func NewDiaryController(du usecase.DiaryUsecase) DiaryController {
	return &diaryController{du: du}
}

func (dc *diaryController) Create(ctx context.Context, d *domain.Diary) (*dto.DiaryResponse, error) {
	diary, err := dc.du.Create(ctx, d)
	if err != nil {
		return nil, err
	}
	res := &dto.DiaryResponse{
		ID:        diary.ID,
		FamilyID:  diary.FamilyID,
		UserID:    diary.UserID,
		Title:     diary.Title,
		Content:   diary.Content,
		CreatedAt: diary.CreatedAt,
	}
	return res, err
}

func (dc *diaryController) List(ctx context.Context, familyID uuid.UUID) ([]dto.DiaryResponse, error) {
	diaries, err := dc.du.List(ctx, familyID)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.DiaryResponse, len(diaries))
	for i, diary := range diaries {
		responses[i] = dto.DiaryResponse{
			ID:        diary.ID,
			FamilyID:  diary.FamilyID,
			UserID:    diary.UserID,
			Title:     diary.Title,
			Content:   diary.Content,
			CreatedAt: diary.CreatedAt,
		}
	}
	return responses, nil
}

func (dc *diaryController) GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error) {
	count, err := dc.du.GetCount(ctx, familyID, year, month)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (dc *diaryController) GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*dto.StreakResponse, error) {
	streak, err := dc.du.GetStreak(ctx, userID, familyID)
	if err != nil {
		return nil, err
	}

	// streak が nil の場合（レコードが存在しない場合）
	if streak == nil {
		return &dto.StreakResponse{
			UserID:        userID,
			FamilyID:      familyID,
			CurrentStreak: 0,
			LastPostDate:  nil,
		}, nil
	}

	res := &dto.StreakResponse{
		UserID:        streak.UserID,
		FamilyID:      streak.FamilyID,
		CurrentStreak: streak.CurrentStreak,
		LastPostDate:  streak.LastPostDate,
	}
	return res, nil
}
