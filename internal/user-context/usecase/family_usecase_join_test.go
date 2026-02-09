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

func TestActivateFamilyContext_Success(t *testing.T) {
	ctx, _, fmr, _, fjr, _, _, tg, _, _, u := newTestEnv()

	userID := uuid.New()
	famID := uuid.New()

	jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

	fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(true, nil) // すでにメンバー
	tg.On("GenerateToken", mock.Anything, userID, famID, domain.RoleMember).Return("signed", nil)

	token, err := u.ActivateFamilyContext(ctx, userID, famID)
	require.NoError(t, err)
	require.Equal(t, "signed", token)

	fjr.AssertExpectations(t)
	fmr.AssertExpectations(t)
	tg.AssertExpectations(t)
}

func TestActivateFamilyContext_NoApprovedRequest(t *testing.T) {
	_, _, _, _, fjr, _, _, _, _, _, u := newTestEnv()

	ctx := context.Background()
	userID := uuid.New()
	famID := uuid.New()

	fjr.On("FindApprovedByUser", mock.Anything, userID).Return(nil, nil)

	_, err := u.ActivateFamilyContext(ctx, userID, famID)
	require.Error(t, err)
}

func TestActivateFamilyContext_NotMember(t *testing.T) {
	_, _, fmr, _, fjr, _, _, _, _, _, u := newTestEnv()

	ctx := context.Background()
	userID := uuid.New()
	famID := uuid.New()
	jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}
	fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(false, nil)

	_, err := u.ActivateFamilyContext(ctx, userID, famID)
	require.Error(t, err)
}

func TestActivateFamilyContext_FamilyIDMismatch(t *testing.T) {
	_, _, _, _, fjr, _, _, _, _, _, u := newTestEnv()

	ctx := context.Background()
	userID := uuid.New()
	famID := uuid.New()
	differentFamID := uuid.New()
	jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

	fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)

	_, err := u.ActivateFamilyContext(ctx, userID, differentFamID)
	require.Error(t, err)
}

func TestActivateFamilyContext_TokenGenError(t *testing.T) {
	_, _, fmr, _, fjr, _, _, tg, _, _, u := newTestEnv()

	ctx := context.Background()
	userID := uuid.New()
	famID := uuid.New()
	jr := &domain.FamilyJoinRequest{FamilyID: famID, UserID: userID}

	fjr.On("FindApprovedByUser", mock.Anything, userID).Return(jr, nil)
	fmr.On("IsUserAlreadyMember", mock.Anything, userID).Return(true, nil)
	tg.On("GenerateToken", mock.Anything, userID, famID, domain.RoleMember).Return("", errors.New("tg err"))

	_, err := u.ActivateFamilyContext(ctx, userID, famID)
	require.Error(t, err)
}
