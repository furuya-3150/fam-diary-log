package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFamilyUsecase struct {
	mock.Mock
}

func (m *MockFamilyUsecase) CreateFamily(ctx context.Context, name string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == "" {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockFamilyUsecase) InviteMembers(ctx context.Context, in usecase.InviteMembersInput) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockFamilyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, token, userID)
	if args.Get(0) == "" {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func TestFamilyController_InviteMembers_Success(t *testing.T) {
	mockUsecase := new(MockFamilyUsecase)
	controller := NewFamilyController(mockUsecase)

	familyID := uuid.New()
	userID := uuid.New()
	emails := []string{"test1@example.com", "test2@example.com"}

	req := &dto.InviteMembersRequest{
		FamilyID: familyID,
		UserID:   userID,
		Emails:   emails,
	}

	expectedInput := usecase.InviteMembersInput{
		FamilyID:      familyID,
		InviterUserID: userID,
		Emails:        emails,
	}

	mockUsecase.On("InviteMembers", mock.Anything, expectedInput).Return(nil)

	err := controller.InviteMembers(context.Background(), req)
	require.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestFamilyController_InviteMembers_UsecaseError(t *testing.T) {
	mockUsecase := new(MockFamilyUsecase)
	controller := NewFamilyController(mockUsecase)

	familyID := uuid.New()
	userID := uuid.New()
	emails := []string{"test@example.com"}

	req := &dto.InviteMembersRequest{
		FamilyID: familyID,
		UserID:   userID,
		Emails:   emails,
	}

	mockUsecase.On("InviteMembers", mock.Anything, mock.AnythingOfType("usecase.InviteMembersInput")).Return(errors.New("usecase error"))

	err := controller.InviteMembers(context.Background(), req)
	require.Error(t, err)
	require.Contains(t, err.Error(), "usecase error")
	mockUsecase.AssertExpectations(t)
}
