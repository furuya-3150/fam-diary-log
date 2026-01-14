package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/diary/usecase"
)

type DiaryController interface {
	Create(ctx context.Context, d *domain.Diary) (*dto.DiaryResponse, error)
	List(ctx context.Context, query domain.DiarySearchCriteria) ([]dto.DiaryResponse, error)
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

func (dc *diaryController) List(ctx context.Context, query domain.DiarySearchCriteria) ([]dto.DiaryResponse, error) {
	// TODO: Implement List
	return []dto.DiaryResponse{}, nil
}
