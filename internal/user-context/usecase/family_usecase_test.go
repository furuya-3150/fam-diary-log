package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/events"
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

func (m *MockFamilyRepo) GetFamilyByID(ctx context.Context, id uuid.UUID) (*domain.Family, error) {
	args := m.Called(ctx, id)
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
func (m *MockFamilyMemberRepo) GetFamilyMemberByUserID(ctx context.Context, userID uuid.UUID) (*domain.FamilyMember, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyMember), args.Error(1)
}

type MockFamilyInvitationRepository struct{ mock.Mock }

func (m *MockFamilyInvitationRepository) CreateInvitation(ctx context.Context, invitation *domain.FamilyInvitation) error {
	args := m.Called(ctx, invitation)
	return args.Error(0)
}

func (m *MockFamilyInvitationRepository) UpdateInvitationTokenAndExpires(ctx context.Context, familyID, inviterUserID uuid.UUID, token string, expiresAt time.Time, invitedEmailsJSON json.RawMessage) error {
	args := m.Called(ctx, familyID, inviterUserID, token, expiresAt, invitedEmailsJSON)
	return args.Error(0)
}

func (m *MockFamilyInvitationRepository) FindInvitationByFamilyID(ctx context.Context, familyID uuid.UUID) (*domain.FamilyInvitation, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyInvitation), args.Error(1)
}

func (m *MockFamilyInvitationRepository) FindInvitationByToken(ctx context.Context, token string) (*domain.FamilyInvitation, error) {
	args := m.Called(ctx, token)
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

type MockWSPblisher struct {
	mock.Mock
}

func (m *MockWSPblisher) Publish(ctx context.Context, uuid uuid.UUID, payload interface{}) error {
	args := m.Called(ctx, uuid, payload)
	return args.Error(0)
}

func (m *MockWSPblisher) CloseUserConnections(userID uuid.UUID) {
	m.Called(userID)
}

type MockMailPublisher struct {
	mock.Mock
}

func (m *MockMailPublisher) Publish(ctx context.Context, event events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockMailPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockTokenGen struct{ mock.Mock }

func (m *MockTokenGen) GenerateToken(ctx context.Context, userID uuid.UUID, familyID uuid.UUID, role domain.Role) (string, error) {
	args := m.Called(ctx, userID, familyID, role)
	return args.String(0), args.Error(1)
}

type MockPublisher struct{ mock.Mock }

func (m *MockPublisher) Publish(ctx context.Context, userID uuid.UUID, payload interface{}) error {
	args := m.Called(ctx, userID, payload)
	return args.Error(0)
}

func (m *MockPublisher) CloseUserConnections(userID uuid.UUID) {
	m.Called(userID)
}

// newTestEnv centralizes mock creation and usecase instantiation for tests
func newTestEnv() (context.Context,
	*MockFamilyRepo,
	*MockFamilyMemberRepo,
	*MockFamilyInvitationRepository,
	*MockUserRepository,
	*MockTxManager,
	*MockTokenGen,
	*MockWSPblisher,
	*MockMailPublisher,
	FamilyUsecase) {

	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	ur := new(MockUserRepository)
	tm := new(MockTxManager)
	tg := new(MockTokenGen)
	wsp := new(MockWSPblisher)
	mp := new(MockMailPublisher)

	u := NewFamilyUsecase(fr, fmr, fir, ur, tm, &clock.Fixed{}, tg, mp)
	ctx := context.Background()

	return ctx, fr, fmr, fir, ur, tm, tg, wsp, mp, u
}

func TestFamilyUsecase_CreateFamily_Success(t *testing.T) {
	ctx, fr, fmr, _, _, tm, tg, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	family := &domain.Family{ID: uuid.New(), Name: "TestFamily"}
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(family, nil)
	fmr.On("AddFamilyMember", ctx, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
	tm.On("CommitTx", ctx).Return(nil)
	tg.On("GenerateToken", ctx, mock.Anything, mock.Anything, mock.Anything).Return("signed", nil)

	_, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.NoError(t, err)
	fr.AssertExpectations(t)
	fmr.AssertExpectations(t)
	tm.AssertExpectations(t)
	tg.AssertExpectations(t)
}

func TestFamilyUsecase_CreateFamily_AlreadyMember(t *testing.T) {
	ctx, _, fmr, _, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(true, nil)

	_, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
}

func TestFamilyUsecase_CreateFamily_RepoError(t *testing.T) {
	ctx, fr, fmr, _, _, tm, _, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(nil, errors.New("repo error"))
	tm.On("RollbackTx", ctx).Return(nil)

	_, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
}

func TestFamilyUsecase_CreateFamily_AddMemberError(t *testing.T) {
	ctx, fr, fmr, _, _, tm, _, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	family := &domain.Family{ID: uuid.New(), Name: "TestFamily"}
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(family, nil)
	fmr.On("AddFamilyMember", ctx, mock.AnythingOfType("*domain.FamilyMember")).Return(errors.New("add member error"))
	tm.On("RollbackTx", ctx).Return(nil)

	_, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
}

func TestFamilyUsecase_CreateFamily_TokenGenerationError(t *testing.T) {
	ctx, fr, fmr, _, _, tm, tg, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(false, nil)
	tm.On("BeginTx", ctx).Return(ctx, nil)
	family := &domain.Family{ID: uuid.New(), Name: "TestFamily"}
	fr.On("CreateFamily", ctx, mock.AnythingOfType("*domain.Family")).Return(family, nil)
	fmr.On("AddFamilyMember", ctx, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
	tm.On("CommitTx", ctx).Return(nil)
	tg.On("GenerateToken", ctx, mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("token generation error"))

	_, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token generation error")
}

func TestFamilyUsecase_InviteMembers_CreateSuccess(t *testing.T) {
	ctx, fr, _, fir, ur, _, _, _, mp, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	// 既存レコードなし
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, nil)
	fir.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.FamilyInvitation")).Return(nil)
	mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mp.On("Close").Return(nil)
	ur.On("GetUserByID", mock.Anything, inviterID).Return(&domain.User{ID: inviterID, Email: "hoge@example.com"}, nil)
	fr.On("GetFamilyByID", mock.Anything, familyID).Return(&domain.Family{ID: familyID, Name: "TestFamily"}, nil)

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.NoError(t, err)
	fir.AssertExpectations(t)
	mp.AssertExpectations(t)
	ur.AssertExpectations(t)
	fr.AssertExpectations(t)
}

// InviteMembers: 正常系 - 既存更新
func TestFamilyUsecase_InviteMembers_UpdateExistingSuccess(t *testing.T) {
	ctx, fr, _, fir, ur, _, _, _, mp, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	existing := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InviterUserID: inviterID, InvitationToken: "old", ExpiresAt: time.Now()}
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(existing, nil)
	fir.On("UpdateInvitationTokenAndExpires", mock.Anything, familyID, inviterID, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("json.RawMessage")).Return(nil)
	ur.On("GetUserByID", mock.Anything, inviterID).Return(&domain.User{ID: inviterID, Email: "hoge@example.com"}, nil)
	fr.On("GetFamilyByID", mock.Anything, familyID).Return(&domain.Family{ID: familyID, Name: "TestFamily"}, nil)
	mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mp.On("Close").Return(nil)

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.NoError(t, err)
	fir.AssertExpectations(t)
	ur.AssertExpectations(t)
	fr.AssertExpectations(t)
	mp.AssertExpectations(t)
}

// InviteMembers: 異常系 - Findでエラー
func TestFamilyUsecase_InviteMembers_FindError(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, errors.New("find error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

// InviteMembers: 異常系 - Createでエラー
func TestFamilyUsecase_InviteMembers_CreateError(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, nil)
	fir.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.FamilyInvitation")).Return(errors.New("create error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

// InviteMembers: 異常系 - Updateでエラー
func TestFamilyUsecase_InviteMembers_UpdateError(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	existing := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InviterUserID: inviterID, InvitationToken: "old", ExpiresAt: time.Now()}
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(existing, nil)
	fir.On("UpdateInvitationTokenAndExpires", mock.Anything, familyID, inviterID, mock.AnythingOfType("string"), mock.Anything, mock.AnythingOfType("json.RawMessage")).Return(errors.New("update error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_Success(t *testing.T) {
	ctx, _, fmr, fir, ur, _, tg, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-123"
	familyID := uuid.New()
	userEmail := "user@example.com"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InvitationToken: token,
		InvitedEmails:   []string{userEmail}, // User's email is in the invited list
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	ur.On("GetUserByID", mock.Anything, userID).Return(&domain.User{ID: userID, Email: userEmail}, nil)
	fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
	tg.On("GenerateToken", ctx, userID, familyID, mock.Anything).Return("test-token", nil)

	signed, err := u.ApplyToFamily(ctx, token, userID)

	require.NotEmpty(t, signed)
	require.NoError(t, err)
	fir.AssertExpectations(t)
	fmr.AssertExpectations(t)
	ur.AssertExpectations(t)
	tg.AssertExpectations(t)
}

func TestFamilyUsecase_ApplyToFamily_NotFound(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-404"

	fir.On("FindInvitationByToken", mock.Anything, token).Return(nil, nil)

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid invitation token")
}

func TestFamilyUsecase_ApplyToFamily_FindError(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-err"

	fir.On("FindInvitationByToken", mock.Anything, token).Return(nil, errors.New("db error"))

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_AlreadyMember(t *testing.T) {
	ctx, _, fmr, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-exist"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InvitationToken: token,
		InvitedEmails:   []string{"user@example.com"},
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(true, nil)

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already a member")
}

func TestFamilyUsecase_ApplyToFamily_EmailNotInvited(t *testing.T) {
	ctx, _, fmr, fir, ur, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-not-invited"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InvitationToken: token,
		InvitedEmails:   []string{"invited@example.com"}, // Different email
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	ur.On("GetUserByID", mock.Anything, userID).Return(&domain.User{ID: userID, Email: "user@example.com"}, nil)

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not invited")
}

func TestFamilyUsecase_ApplyToFamily_AddMemberError(t *testing.T) {
	ctx, _, fmr, fir, ur, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-add-err"
	familyID := uuid.New()
	userEmail := "user@example.com"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InvitationToken: token,
		InvitedEmails:   []string{userEmail},
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	ur.On("GetUserByID", mock.Anything, userID).Return(&domain.User{ID: userID, Email: userEmail}, nil)
	fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(errors.New("add member error"))

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_ExpiredToken(t *testing.T) {
	// 期限切れをテストするため、固定時刻を設定したClockを使用
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	ur := new(MockUserRepository)
	tm := new(MockTxManager)
	tg := new(MockTokenGen)
	mp := new(MockMailPublisher)

	// Clockを2025年1月1日に設定
	fixedClock := &clock.Fixed{Time: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	u := NewFamilyUsecase(fr, fmr, fir, ur, tm, fixedClock, tg, mp)

	ctx := context.Background()
	userID := uuid.New()
	token := "tok-expired"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        uuid.New(),
		InvitationToken: token,
		InvitedEmails:   []string{"user@example.com"},
		ExpiresAt:       time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), // Clockの時刻より古い
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	// 期限切れの場合、IsUserAlreadyMemberは呼ばれない

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expired")
}

func TestFamilyUsecase_ApplyToFamily_TokenGenerationError(t *testing.T) {
	ctx, _, fmr, fir, ur, _, tg, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-123"
	familyID := uuid.New()
	userEmail := "user@example.com"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        familyID,
		InvitationToken: token,
		InvitedEmails:   []string{userEmail},
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	ur.On("GetUserByID", mock.Anything, userID).Return(&domain.User{ID: userID, Email: userEmail}, nil)
	fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
	tg.On("GenerateToken", ctx, userID, familyID, mock.Anything).Return("", errors.New("token error"))

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "token error")
}

func TestFamilyUsecase_ApplyToFamily_InvalidUser(t *testing.T) {
	ctx, _, fmr, fir, ur, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-123"

	inv := &domain.FamilyInvitation{
		ID:              uuid.New(),
		FamilyID:        uuid.New(),
		InvitationToken: token,
		InvitedEmails:   []string{"user@example.com"},
		ExpiresAt:       time.Now().Add(24 * time.Hour),
	}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	ur.On("GetUserByID", mock.Anything, userID).Return(nil, nil)

	_, err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid user")
}
