package usecase

import (
    "context"
    "errors"
    "testing"

    "github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
    "github.com/furuya-3150/fam-diary-log/pkg/clock"
    "github.com/google/uuid"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

type MockTokenGen struct{ mock.Mock }

func (m *MockTokenGen) GenerateToken(ctx context.Context, userID uuid.UUID, familyID uuid.UUID, role domain.Role) (string, int64, error) {
    args := m.Called(ctx, userID, familyID, role)
    return args.String(0), int64(args.Int(1)), args.Error(2)
}

func TestJoinFamilyIfApproved_Success(t *testing.T) {
    fr := new(MockFamilyRepo)
    fmr := new(MockFamilyMemberRepo)
    fir := new(MockFamilyInvitationRepository)
    fjr := new(MockFamilyJoinRequestRepository)
    tm := new(MockTxManager)
    tg := new(MockTokenGen)

    now := clock.Fixed{}
    u := NewFamilyUsecaseWithToken(fr, fmr, fir, fjr, tm, &now, tg)

    ctx := context.Background()
    userID := uuid.New()
    famID := uuid.New()

    jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

    fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
    fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
    // tm.On("BeginTx", mock.Anything).Return(ctx, nil)
    fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
    // tm.On("CommitTx", mock.Anything).Return(nil)
    tg.On("GenerateToken", mock.Anything, userID, famID, domain.RoleMember).Return("signed", 3600, nil)

    token, expires, err := u.JoinFamilyIfApproved(ctx, userID)
    require.NoError(t, err)
    require.Equal(t, "signed", token)
    require.Equal(t, int64(3600), expires)

    fjr.AssertExpectations(t)
    fmr.AssertExpectations(t)
    tg.AssertExpectations(t)
}

func TestJoinFamilyIfApproved_NoApprovedRequest(t *testing.T) {
    fr := new(MockFamilyRepo)
    fmr := new(MockFamilyMemberRepo)
    fir := new(MockFamilyInvitationRepository)
    fjr := new(MockFamilyJoinRequestRepository)
    tm := new(MockTxManager)
    tg := new(MockTokenGen)

    u := NewFamilyUsecaseWithToken(fr, fmr, fir, fjr, tm, &clock.Fixed{}, tg)

    ctx := context.Background()
    userID := uuid.New()

    fjr.On("FindApprovedByUser", mock.Anything, userID).Return(nil, nil)

    _, _, err := u.JoinFamilyIfApproved(ctx, userID)
    require.Error(t, err)
}

func TestJoinFamilyIfApproved_AlreadyMember(t *testing.T) {
    fr := new(MockFamilyRepo)
    fmr := new(MockFamilyMemberRepo)
    fir := new(MockFamilyInvitationRepository)
    fjr := new(MockFamilyJoinRequestRepository)
    tm := new(MockTxManager)
    tg := new(MockTokenGen)

    u := NewFamilyUsecaseWithToken(fr, fmr, fir, fjr, tm, &clock.Fixed{}, tg)

    ctx := context.Background()
    userID := uuid.New()
    fjr.On("FindApprovedByUser", mock.Anything, userID).Return(&domain.FamilyJoinRequest{}, nil)
    fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(true, nil)

    _, _, err := u.JoinFamilyIfApproved(ctx, userID)
    require.Error(t, err)
}

func TestJoinFamilyIfApproved_AddMemberError(t *testing.T) {
    fr := new(MockFamilyRepo)
    fmr := new(MockFamilyMemberRepo)
    fir := new(MockFamilyInvitationRepository)
    fjr := new(MockFamilyJoinRequestRepository)
    tm := new(MockTxManager)
    tg := new(MockTokenGen)

    now := clock.Fixed{}
    u := NewFamilyUsecaseWithToken(fr, fmr, fir, fjr, tm, &now, tg)

    ctx := context.Background()
    userID := uuid.New()
    famID := uuid.New()
    jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

    fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
    fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
    // tm.On("BeginTx", mock.Anything).Return(ctx, nil)
    fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(errors.New("add err"))
    // tm.On("RollbackTx", mock.Anything).Return(nil)

    _, _, err := u.JoinFamilyIfApproved(ctx, userID)
    require.Error(t, err)
}

func TestJoinFamilyIfApproved_TokenGenError(t *testing.T) {
    fr := new(MockFamilyRepo)
    fmr := new(MockFamilyMemberRepo)
    fir := new(MockFamilyInvitationRepository)
    fjr := new(MockFamilyJoinRequestRepository)
    tm := new(MockTxManager)
    tg := new(MockTokenGen)

    now := clock.Fixed{}
    u := NewFamilyUsecaseWithToken(fr, fmr, fir, fjr, tm, &now, tg)

    ctx := context.Background()
    userID := uuid.New()
    famID := uuid.New()
    jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

    fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
    fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)
    // tm.On("BeginTx", mock.Anything).Return(ctx, nil)
    fmr.On("AddFamilyMember", mock.Anything, mock.AnythingOfType("*domain.FamilyMember")).Return(nil)
    // tm.On("CommitTx", mock.Anything).Return(nil)
    tg.On("GenerateToken", mock.Anything, userID, famID, domain.RoleMember).Return("", 0, errors.New("tg err"))

    _, _, err := u.JoinFamilyIfApproved(ctx, userID)
    require.Error(t, err)
}
