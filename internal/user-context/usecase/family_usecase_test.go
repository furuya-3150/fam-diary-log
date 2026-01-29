package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
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

type MockFamilyInvitationRepository struct{ mock.Mock }

func (m *MockFamilyInvitationRepository) CreateInvitation(ctx context.Context, invitation *domain.FamilyInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockFamilyInvitationRepository) UpdateInvitationTokenAndExpires(ctx context.Context, familyID, inviterUserID uuid.UUID, token string, expiresAt time.Time) error {
	args := m.Called(ctx, familyID, inviterUserID, token, expiresAt)
	return args.Error(0)
}

func (m *MockFamilyInvitationRepository) FindInvitationByFamilyID(ctx context.Context, familyID uuid.UUID) (*domain.FamilyInvitation, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyInvitation), args.Error(1)
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
	tir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, tir, tm, &clock.Fixed{})

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
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

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
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

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
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

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

// InviteMembers: 正常系 - 新規作成
func TestFamilyUsecase_InviteMembers_CreateSuccess(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

	ctx := context.Background()
	familyID := uuid.New()
	inviterID := uuid.New()

	// 既存レコードなし
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, nil)
	fir.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.FamilyInvitation")).Return(nil)

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.NoError(t, err)
	fir.AssertExpectations(t)
}

// InviteMembers: 正常系 - 既存更新
func TestFamilyUsecase_InviteMembers_UpdateExistingSuccess(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

	ctx := context.Background()
	familyID := uuid.New()
	inviterID := uuid.New()

	existing := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InviterUserID: inviterID, InvitationToken: "old", ExpiresAt: time.Now()}
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(existing, nil)
	fir.On("UpdateInvitationTokenAndExpires", mock.Anything, familyID, inviterID, mock.AnythingOfType("string"), mock.Anything).Return(nil)

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.NoError(t, err)
	fir.AssertExpectations(t)
}

// InviteMembers: 異常系 - Findでエラー
func TestFamilyUsecase_InviteMembers_FindError(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

	ctx := context.Background()
	familyID := uuid.New()
	inviterID := uuid.New()

	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, errors.New("find error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

// InviteMembers: 異常系 - Createでエラー
func TestFamilyUsecase_InviteMembers_CreateError(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

	ctx := context.Background()
	familyID := uuid.New()
	inviterID := uuid.New()

	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, nil)
	fir.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.FamilyInvitation")).Return(errors.New("create error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

// InviteMembers: 異常系 - Updateでエラー
func TestFamilyUsecase_InviteMembers_UpdateError(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tm := new(MockTxManager)
	u := NewFamilyUsecase(fr, fmr, fir, tm, &clock.Fixed{})

	ctx := context.Background()
	familyID := uuid.New()
	inviterID := uuid.New()

	existing := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InviterUserID: inviterID, InvitationToken: "old", ExpiresAt: time.Now()}
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(existing, nil)
	fir.On("UpdateInvitationTokenAndExpires", mock.Anything, familyID, inviterID, mock.AnythingOfType("string"), mock.Anything).Return(errors.New("update error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}
