package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserUsecase implements usecase.UserUsecase for tests
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) EditUser(ctx context.Context, input *usecase.EditUserInput) (*domain.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func TestUserController_EditProfile_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	controller := NewUserController(mockUsecase)

	id := uuid.New()
	req := &dto.EditUserRequest{
		ID:    id,
		Name:  "Bob",
		Email: "bob@example.com",
	}
	expected := &domain.User{ID: id, Email: req.Email, Name: req.Name}

	// モックのEditUserが呼ばれること・引数が正しいことを検証
	mockUsecase.On("EditUser", mock.Anything, mock.MatchedBy(func(input *usecase.EditUserInput) bool {
		return input.ID == id.String() && input.Name == req.Name && input.Email == req.Email
	})).Return(expected, nil)

	got, err := controller.EditProfile(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, req.Email, got.Email)
	require.Equal(t, req.Name, got.Name)
	mockUsecase.AssertExpectations(t)
}

func TestUserController_EditProfile_Error(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	controller := NewUserController(mockUsecase)

	id := uuid.New()
	req := &dto.EditUserRequest{
		ID:    id,
		Name:  "Bob",
		Email: "bob@example.com",
	}
	mockUsecase.On("EditUser", mock.Anything, mock.Anything).Return(nil, errors.New("fail"))

	_, err := controller.EditProfile(context.Background(), req)
	require.Error(t, err)
	mockUsecase.AssertExpectations(t)
}

func TestUserController_GetProfile_Success(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	controller := NewUserController(mockUsecase)

	id := uuid.New()
	expected := &domain.User{ID: id, Email: "test@example.com"}
	mockUsecase.On("GetUser", mock.Anything, id).Return(expected, nil)

	got, err := controller.GetProfile(context.Background(), id)
	require.NoError(t, err)
	require.Equal(t, expected.ID, got.ID)
	require.Equal(t, expected.Email, got.Email)
	mockUsecase.AssertExpectations(t)
}

func TestUserController_GetProfile_Error(t *testing.T) {
	mockUsecase := new(MockUserUsecase)
	controller := NewUserController(mockUsecase)

	id := uuid.New()
	mockUsecase.On("GetUser", mock.Anything, id).Return(nil, errors.New("fail"))

	_, err := controller.GetProfile(context.Background(), id)
	require.Error(t, err)
	mockUsecase.AssertExpectations(t)
}
