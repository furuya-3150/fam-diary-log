package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return ctx, args.Error(1)
	}
	return args.Get(0).(context.Context), args.Error(1)
}
func (m *MockTransactionManager) CommitTx(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *MockTransactionManager) RollbackTx(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupUserUsecase(repo *MockUserRepository, tx *MockTransactionManager) *userUsecase {
	return &userUsecase{repo: repo, tm: tx}
}

func TestEditUser_Success(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	existing := &domain.User{ID: id, Email: "old@example.com", Name: "Old", CreatedAt: now, UpdatedAt: now}
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	repo.On("GetUserByID", mock.Anything, id).Return(existing, nil)
	repo.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.ID == id && u.Name == "New" && u.Email == "new@example.com"
	})).Return(&domain.User{ID: id, Name: "New", Email: "new@example.com", CreatedAt: now, UpdatedAt: now}, nil)
	uc := setupUserUsecase(repo, tx)
	in := &EditUserInput{ID: id.String(), Name: "New", Email: "new@example.com"}
	got, err := uc.EditUser(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, "New", got.Name)
	require.Equal(t, "new@example.com", got.Email)
	repo.AssertExpectations(t)
}

func TestEditUser_InvalidUUID(t *testing.T) {
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	uc := setupUserUsecase(repo, tx)
	in := &EditUserInput{ID: "invalid-uuid", Name: "X", Email: "x@example.com"}
	_, err := uc.EditUser(context.Background(), in)
	require.Error(t, err)
	var verr *pkgerrors.ValidationError
	require.ErrorAs(t, err, &verr)
}

func TestEditUser_UserNotFound(t *testing.T) {
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	repo.On("GetUserByID", mock.Anything, mock.Anything).Return(nil, nil)
	uc := setupUserUsecase(repo, tx)
	id := uuid.New()
	in := &EditUserInput{ID: id.String(), Name: "X", Email: "x@example.com"}
	_, err := uc.EditUser(context.Background(), in)
	require.Error(t, err)
	var verr *pkgerrors.ValidationError
	require.ErrorAs(t, err, &verr)
	repo.AssertExpectations(t)
}

func TestEditUser_UpdateError(t *testing.T) {
	id := uuid.New()
	now := time.Now()
	existing := &domain.User{ID: id, Email: "old@example.com", Name: "Old", CreatedAt: now, UpdatedAt: now}
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	repo.On("GetUserByID", mock.Anything, id).Return(existing, nil)
	repo.On("UpdateUser", mock.Anything, mock.Anything).Return(nil, errors.New("db fail"))
	uc := setupUserUsecase(repo, tx)
	in := &EditUserInput{ID: id.String(), Name: "New", Email: "new@example.com"}
	_, err := uc.EditUser(context.Background(), in)
	require.Error(t, err)
	var ierr *pkgerrors.InternalError
	require.ErrorAs(t, err, &ierr)
	repo.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	id := uuid.New()
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	user := &domain.User{ID: id, Email: "test@example.com"}
	repo.On("GetUserByID", mock.Anything, id).Return(user, nil)
	uc := setupUserUsecase(repo, tx)
	got, err := uc.GetUser(context.Background(), id.String())
	require.NoError(t, err)
	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Email, got.Email)
	repo.AssertExpectations(t)
}

func TestGetUser_InvalidUUID(t *testing.T) {
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	uc := setupUserUsecase(repo, tx)
	_, err := uc.GetUser(context.Background(), "invalid-uuid")
	require.Error(t, err)
	var verr *pkgerrors.ValidationError
	require.ErrorAs(t, err, &verr)
}

func TestGetUser_NotFound(t *testing.T) {
	id := uuid.New()
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	repo.On("GetUserByID", mock.Anything, id).Return(nil, nil)
	uc := setupUserUsecase(repo, tx)
	_, err := uc.GetUser(context.Background(), id.String())
	require.Error(t, err)
	var verr *pkgerrors.ValidationError
	require.ErrorAs(t, err, &verr)
	repo.AssertExpectations(t)
}

func TestGetUser_RepoError(t *testing.T) {
	id := uuid.New()
	repo := new(MockUserRepository)
	tx := new(MockTransactionManager)
	repo.On("GetUserByID", mock.Anything, id).Return(nil, errors.New("db fail"))
	uc := setupUserUsecase(repo, tx)
	_, err := uc.GetUser(context.Background(), id.String())
	require.Error(t, err)
	var ierr *pkgerrors.InternalError
	require.ErrorAs(t, err, &ierr)
	repo.AssertExpectations(t)
}
