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

type MockFamilyRepo struct{ mock.Mock }

func (m *MockFamilyRepo) CreateFamily(ctx context.Context, family *domain.Family) (*domain.Family, error) {
	args := m.Called(ctx, family)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Family), args.Error(1)
}

type MockFamilyMemberRepo struct{ mock.Mock }

func (m *MockFamilyMemberRepo) IsUserAlreadyMember(ctx context.Context, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}
func (m *MockFamilyMemberRepo) AddFamilyMember(ctx context.Context, member *domain.FamilyMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

type MockTxManager struct{ mock.Mock }

func (m *MockTxManager) BeginTx(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	return ctx, args.Error(1)
}
func (m *MockTxManager) CommitTx(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}
func (m *MockTxManager) RollbackTx(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestFamilyUsecase_CreateFamily_Success(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, tm)

	ctx := context.Background()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	family := &domain.Family{ID: uuid.New(), Name: "TestFamily"}
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(family, nil)
	fmr.On("AddFamilyMember", ctx, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
	tm.On("CommitTx", ctx).Return(nil)

	result, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.NoError(t, err)
	require.Equal(t, family, result)
	fr.AssertExpectations(t)
	fmr.AssertExpectations(t)
	tm.AssertExpectations(t)
}

func TestFamilyUsecase_CreateFamily_AlreadyMember(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, tm)

	ctx := context.Background()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(true, nil)

	result, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestFamilyUsecase_CreateFamily_RepoError(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, tm)

	ctx := context.Background()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(nil, errors.New("repo error"))
	tm.On("RollbackTx", ctx).Return(nil)

	result, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestFamilyUsecase_CreateFamily_AddMemberError(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, tm)

	ctx := context.Background()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	family := &domain.Family{ID: uuid.New(), Name: "TestFamily"}
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(family, nil)
	fmr.On("AddFamilyMember", ctx, mock.AnythingOfType("*domain.FamilyMember")).Return(errors.New("add member error"))
	tm.On("RollbackTx", ctx).Return(nil)

	result, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
	require.Nil(t, result)
}
