package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNotificationRepo struct{ mock.Mock }

func (m *MockNotificationRepo) GetByUserAndFamily(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error) {
	args := m.Called(ctx, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotificationSetting), args.Error(1)
}

func (m *MockNotificationRepo) Upsert(ctx context.Context, setting *domain.NotificationSetting) (*domain.NotificationSetting, error) {
	args := m.Called(ctx, setting)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotificationSetting), args.Error(1)
}

func TestNotificationUsecase_GetNotificationSetting_DefaultsWhenNotFound(t *testing.T) {
	repo := new(MockNotificationRepo)
	u := NewNotificationUsecase(repo)

	ctx := context.Background()
	userID := uuid.New()
	familyID := uuid.New()

	repo.On("GetByUserAndFamily", mock.Anything, userID, familyID).Return(nil, nil)

	s, err := u.GetNotificationSetting(ctx, userID, familyID)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, true, s.PostCreatedEnabled)
}

func TestNotificationUsecase_GetNotificationSetting_FromRepo(t *testing.T) {
	repo := new(MockNotificationRepo)
	u := NewNotificationUsecase(repo)

	ctx := context.Background()
	userID := uuid.New()
	familyID := uuid.New()

	exist := &domain.NotificationSetting{UserID: userID, FamilyID: familyID, PostCreatedEnabled: false}
	repo.On("GetByUserAndFamily", mock.Anything, userID, familyID).Return(exist, nil)

	s, err := u.GetNotificationSetting(ctx, userID, familyID)
	require.NoError(t, err)
	require.NotNil(t, s)
	require.Equal(t, false, s.PostCreatedEnabled)
}

func TestNotificationUsecase_GetNotificationSetting_RepoError(t *testing.T) {
	repo := new(MockNotificationRepo)
	u := NewNotificationUsecase(repo)

	ctx := context.Background()
	userID := uuid.New()
	familyID := uuid.New()

	repo.On("GetByUserAndFamily", mock.Anything, userID, familyID).Return(nil, errors.New("db err"))

	s, err := u.GetNotificationSetting(ctx, userID, familyID)
	require.Error(t, err)
	require.Nil(t, s)
}

func TestNotificationUsecase_UpdateNotificationSetting_Success(t *testing.T) {
	repo := new(MockNotificationRepo)
	u := NewNotificationUsecase(repo)

	ctx := context.Background()
	ns := &domain.NotificationSetting{UserID: uuid.New(), FamilyID: uuid.New(), PostCreatedEnabled: true}

	repo.On("Upsert", mock.Anything, ns).Return(ns, nil)

	err := u.UpdateNotificationSetting(ctx, ns)
	require.NoError(t, err)
}

func TestNotificationUsecase_UpdateNotificationSetting_Error(t *testing.T) {
	repo := new(MockNotificationRepo)
	u := NewNotificationUsecase(repo)

	ctx := context.Background()
	ns := &domain.NotificationSetting{UserID: uuid.New(), FamilyID: uuid.New(), PostCreatedEnabled: true}

	repo.On("Upsert", mock.Anything, ns).Return(nil, errors.New("upsert err"))

	err := u.UpdateNotificationSetting(ctx, ns)
	require.Error(t, err)
}
