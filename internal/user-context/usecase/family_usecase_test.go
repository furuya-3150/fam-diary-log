package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	apperrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
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

func (m *MockFamilyInvitationRepository) FindInvitationByToken(ctx context.Context, token string) (*domain.FamilyInvitation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyInvitation), args.Error(1)
}

type MockFamilyJoinRequestRepository struct{ mock.Mock }

func (m *MockFamilyJoinRequestRepository) CreateJoinRequest(ctx context.Context, req *domain.FamilyJoinRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockFamilyJoinRequestRepository) FindPendingRequest(ctx context.Context, familyID, userID uuid.UUID) (*domain.FamilyJoinRequest, error) {
	args := m.Called(ctx, familyID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyJoinRequest), args.Error(1)
}

func (m *MockFamilyJoinRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.FamilyJoinRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyJoinRequest), args.Error(1)
}

func (m *MockFamilyJoinRequestRepository) UpdateStatusByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockFamilyJoinRequestRepository) FindApprovedByUser(ctx context.Context, userID uuid.UUID) (*domain.FamilyJoinRequest, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FamilyJoinRequest), args.Error(1)
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

// newTestEnv centralizes mock creation and usecase instantiation for tests
func newTestEnv() (context.Context,
	*MockFamilyRepo,
	*MockFamilyMemberRepo,
	*MockFamilyInvitationRepository,
	*MockFamilyJoinRequestRepository,
	*MockTxManager,
	*MockTokenGen,
	*MockWSPblisher,
	*MockMailPublisher,
	FamilyUsecase) {

	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tjr := new(MockFamilyJoinRequestRepository)
	tm := new(MockTxManager)
	tg := new(MockTokenGen)
	wsp := new(MockWSPblisher)
	mp := new(MockMailPublisher)

	u := NewFamilyUsecase(fr, fmr, fir, tjr, tm, &clock.Fixed{}, tg, wsp, mp)
	ctx := context.Background()

	return ctx, fr, fmr, fir, tjr, tm, tg, wsp, mp, u
}

func TestFamilyUsecase_CreateFamily_Success(t *testing.T) {
	ctx, fr, fmr, _, _, tm, _, _, _, u := newTestEnv()
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
	ctx, _, fmr, _, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()

	fmr.On("IsUserAlreadyMember", ctx, userID).Return(true, nil)

	result, err := u.CreateFamily(ctx, "TestFamily", userID)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestFamilyUsecase_CreateFamily_RepoError(t *testing.T) {
	ctx, fr, fmr, _, _, tm, _, _, _, u := newTestEnv()
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
	ctx, fr, fmr, _, _, tm, _, _, _, u := newTestEnv()
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

func TestFamilyUsecase_InviteMembers_CreateSuccess(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, mp, u := newTestEnv()
	familyID := uuid.New()
	inviterID := uuid.New()

	// 既存レコードなし
	fir.On("FindInvitationByFamilyID", mock.Anything, familyID).Return(nil, nil)
	fir.On("CreateInvitation", mock.Anything, mock.AnythingOfType("*domain.FamilyInvitation")).Return(nil)
	mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mp.On("Close").Return(nil)

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.NoError(t, err)
	fir.AssertExpectations(t)
	mp.AssertExpectations(t)
}

// InviteMembers: 正常系 - 既存更新
func TestFamilyUsecase_InviteMembers_UpdateExistingSuccess(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
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
	fir.On("UpdateInvitationTokenAndExpires", mock.Anything, familyID, inviterID, mock.AnythingOfType("string"), mock.Anything).Return(errors.New("update error"))

	err := u.InviteMembers(ctx, InviteMembersInput{FamilyID: familyID, InviterUserID: inviterID, Emails: []string{"a@example.com"}})
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_Success(t *testing.T) {
	ctx, _, fmr, fir, tjr, _, _, _, mp, u := newTestEnv()
	userID := uuid.New()
	token := "tok-123"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InvitationToken: token, ExpiresAt: time.Now().Add(24 * time.Hour)}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	tjr.On("FindPendingRequest", mock.Anything, familyID, userID).Return(nil, nil)
	tjr.On("CreateJoinRequest", mock.Anything, mock.AnythingOfType("*domain.FamilyJoinRequest")).Return(nil)
	mp.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mp.On("Close").Return(nil)

	err := u.ApplyToFamily(ctx, token, userID)
	require.NoError(t, err)
	fir.AssertExpectations(t)
	fmr.AssertExpectations(t)
}

func TestFamilyUsecase_ApplyToFamily_NotFound(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-404"

	fir.On("FindInvitationByToken", mock.Anything, token).Return(nil, nil)

	err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "invitation not found")
}

func TestFamilyUsecase_ApplyToFamily_FindError(t *testing.T) {
	ctx, _, _, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-err"

	fir.On("FindInvitationByToken", mock.Anything, token).Return(nil, errors.New("db error"))

	err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_AlreadyMember(t *testing.T) {
	ctx, _, fmr, fir, _, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-exist"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InvitationToken: token, ExpiresAt: time.Now().Add(24 * time.Hour)}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(true, nil)

	err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "already a member")
}

func TestFamilyUsecase_ApplyToFamily_AddMemberError(t *testing.T) {
	ctx, _, fmr, fir, tjr, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-add-err"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InvitationToken: token, ExpiresAt: time.Now().Add(24 * time.Hour)}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	tjr.On("FindPendingRequest", mock.Anything, familyID, userID).Return(nil, nil)
	tjr.On("CreateJoinRequest", mock.Anything, mock.AnythingOfType("*domain.FamilyJoinRequest")).Return(errors.New("create error"))

	err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
}

func TestFamilyUsecase_ApplyToFamily_AlreadyPending(t *testing.T) {
	ctx, _, fmr, fir, tjr, _, _, _, _, u := newTestEnv()
	userID := uuid.New()
	token := "tok-pending"
	familyID := uuid.New()

	inv := &domain.FamilyInvitation{ID: uuid.New(), FamilyID: familyID, InvitationToken: token, ExpiresAt: time.Now().Add(24 * time.Hour)}
	fir.On("FindInvitationByToken", mock.Anything, token).Return(inv, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
	existing := &domain.FamilyJoinRequest{ID: uuid.New(), FamilyID: familyID, UserID: userID, Status: domain.JoinRequestStatusPending}
	tjr.On("FindPendingRequest", mock.Anything, familyID, userID).Return(existing, nil)

	err := u.ApplyToFamily(ctx, token, userID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "pending")
}

func TestFamilyUsecase_RespondToJoinRequest_SuccessApproved(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tjr := new(MockFamilyJoinRequestRepository)
	tm := new(MockTxManager)
	tg := new(MockTokenGen)
	wsp := new(MockWSPblisher)
	mp := new(MockMailPublisher)
	now := time.Now()
	u := NewFamilyUsecase(fr, fmr, fir, tjr, tm, &clock.Fixed{Time: now}, tg, wsp, mp)

	ctx := context.Background()
	requestID := uuid.New()
	requestUserID := uuid.New()
	responderID := uuid.New()
	familyID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:       requestID,
		FamilyID: familyID,
		UserID:   requestUserID,
		Status:   domain.JoinRequestStatusPending,
	}

	tjr.On("FindByID", mock.Anything, requestID).Return(jr, nil)
	tjr.On("UpdateStatusByID", mock.Anything, requestID, mock.MatchedBy(func(updates map[string]interface{}) bool {
		s, ok := updates["status"].(int)
		if !ok || s != int(domain.JoinRequestStatusApproved) {
			return false
		}
		if _, ok := updates["responded_user_id"].(uuid.UUID); !ok {
			return false
		}
		if _, ok := updates["responded_at"].(time.Time); !ok {
			return false
		}
		return true
	})).Return(nil)

	wsp.On("Publish", mock.Anything, requestUserID, mock.Anything).Return(nil)
	wsp.On("CloseUserConnections", requestUserID).Return()

	err := u.RespondToJoinRequest(ctx, requestID, domain.JoinRequestStatusApproved, responderID)
	require.NoError(t, err)

	tjr.AssertExpectations(t)
	fmr.AssertExpectations(t)
	tm.AssertExpectations(t)
}

func TestFamilyUsecase_RespondToJoinRequest_SuccessRejected_NoAddMember(t *testing.T) {
	fr := new(MockFamilyRepo)
	fmr := new(MockFamilyMemberRepo)
	fir := new(MockFamilyInvitationRepository)
	tjr := new(MockFamilyJoinRequestRepository)
	tm := new(MockTxManager)
	now := time.Now()
	tg := new(MockTokenGen)
	wsp := new(MockWSPblisher)
	mp := new(MockMailPublisher)
	u := NewFamilyUsecase(fr, fmr, fir, tjr, tm, &clock.Fixed{Time: now}, tg, wsp, mp)

	ctx := context.Background()
	requestID := uuid.New()
	responderID := uuid.New()
	requestUserID := uuid.New()
	familyID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:       requestID,
		FamilyID: familyID,
		UserID:   requestUserID,
		Status:   domain.JoinRequestStatusPending,
	}

	tjr.On("FindByID", mock.Anything, requestID).Return(jr, nil)
	tjr.On("UpdateStatusByID", mock.Anything, requestID, mock.MatchedBy(func(updates map[string]interface{}) bool {
		s, ok := updates["status"].(int)
		if !ok || s != int(domain.JoinRequestStatusRejected) {
			return false
		}
		return true
	})).Return(nil)
	wsp.On("Publish", mock.Anything, requestUserID, mock.Anything).Return(nil)
	wsp.On("CloseUserConnections", requestUserID).Return()

	err := u.RespondToJoinRequest(ctx, requestID, domain.JoinRequestStatusRejected, responderID)
	require.NoError(t, err)

	tjr.AssertExpectations(t)
}

func TestFamilyUsecase_RespondToJoinRequest_NotFound(t *testing.T) {
	ctx, _, _, _, tjr, _, _, _, _, u := newTestEnv()
	requestID := uuid.New()
	responderID := uuid.New()

	tjr.On("FindByID", mock.Anything, requestID).Return(nil, nil)

	err := u.RespondToJoinRequest(ctx, requestID, domain.JoinRequestStatusApproved, responderID)
	require.Error(t, err)
	var nf *apperrors.NotFoundError
	require.ErrorAs(t, err, &nf)
}

func TestFamilyUsecase_RespondToJoinRequest_BadRequest_WhenNotPending(t *testing.T) {
	ctx, _, _, _, tjr, _, _, _, _, u := newTestEnv()
	requestID := uuid.New()
	responderID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:     requestID,
		Status: domain.JoinRequestStatusApproved,
	}

	tjr.On("FindByID", mock.Anything, requestID).Return(jr, nil)

	err := u.RespondToJoinRequest(ctx, requestID, domain.JoinRequestStatusRejected, responderID)
	require.Error(t, err)
	var br *apperrors.BadRequestError
	require.ErrorAs(t, err, &br)
}

func TestFamilyUsecase_RespondToJoinRequest_UpdateError_Rollback(t *testing.T) {
	ctx, _, _, _, tjr, tm, _, wsp, mp, u := newTestEnv()
	requestID := uuid.New()
	responderID := uuid.New()

	jr := &domain.FamilyJoinRequest{
		ID:     requestID,
		Status: domain.JoinRequestStatusPending,
	}

	tjr.On("FindByID", mock.Anything, requestID).Return(jr, nil)
	tjr.On("UpdateStatusByID", mock.Anything, requestID, mock.Anything).Return(errors.New("update failed"))

	err := u.RespondToJoinRequest(ctx, requestID, domain.JoinRequestStatusRejected, responderID)
	require.Error(t, err)
	tm.AssertExpectations(t)
	wsp.AssertExpectations(t)
	mp.AssertExpectations(t)
}
