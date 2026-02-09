package usecase

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/google/uuid"
)

type NotificationUsecase interface {
	GetNotificationSetting(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error)
	UpdateNotificationSetting(ctx context.Context, setting *domain.NotificationSetting) error
}

type notificationUsecase struct {
	repo repository.NotificationSettingRepository
}

func NewNotificationUsecase(repo repository.NotificationSettingRepository) NotificationUsecase {
	return &notificationUsecase{repo: repo}
}

func (u *notificationUsecase) GetNotificationSetting(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error) {
	s, err := u.repo.GetByUserAndFamily(ctx, userID, familyID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		// default setting
		return &domain.NotificationSetting{UserID: userID, FamilyID: familyID, PostCreatedEnabled: false}, nil
	}
	return s, nil
}

func (u *notificationUsecase) UpdateNotificationSetting(ctx context.Context, setting *domain.NotificationSetting) error {
	_, err := u.repo.Upsert(ctx, setting)
	return err
}
