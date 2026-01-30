package controller

import (
	"context"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFamilyUsecase struct {
	mock.Mock
}

func (m *MockFamilyUsecase) CreateFamily(ctx context.Context, name string, userID uuid.UUID) (*domain.Family, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Family), args.Error(1)
}

func (m *MockFamilyUsecase) InviteMembers(ctx context.Context, in usecase.InviteMembersInput) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}	

func (m *MockFamilyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) error {
	args := m.Called(ctx, token, userID)
	return args.Error(0)
}

func TestFamilyController_CreateFamily_Success(t *testing.T) {
	mockUsecase := new(MockFamilyUsecase)
	controller := NewFamilyController(mockUsecase)

	userID := uuid.New()
	req := &dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	expected := &domain.Family{
		ID:   uuid.New(),
		Name: req.Name,
	}
	mockUsecase.On("CreateFamily", mock.Anything, req.Name, userID).Return(expected, nil)

	got, err := controller.CreateFamily(context.Background(), req, userID)
	require.NoError(t, err)
	require.Equal(t, req.Name, got.Name)
	mockUsecase.AssertExpectations(t)
}

func TestFamilyController_CreateFamily_Error(t *testing.T) {
	mockUsecase := new(MockFamilyUsecase)
	controller := NewFamilyController(mockUsecase)

	userID := uuid.New()
	req := &dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	mockUsecase.On("CreateFamily", mock.Anything, req.Name, userID).Return(nil, assert.AnError)

	got, err := controller.CreateFamily(context.Background(), req, userID)
	require.Error(t, err)
	require.Nil(t, got)
	mockUsecase.AssertExpectations(t)
}
