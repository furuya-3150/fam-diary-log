package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/oauth"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository は UserRepository のモック
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByProviderID(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	args := m.Called(ctx, provider, providerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetAdminUsersByFamilyID(ctx context.Context, familyID uuid.UUID) ([]*domain.User, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

// MockOAuthProvider は OAuthProvider のモック
type MockOAuthProvider struct {
	mock.Mock
}

func (m *MockOAuthProvider) GetProviderName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockOAuthProvider) GetAuthURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func (m *MockOAuthProvider) ExchangeCode(ctx context.Context, code string) (*oauth.OAuthUserInfo, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oauth.OAuthUserInfo), args.Error(1)
}

func setupAuthUsecase(
	t *testing.T,
	mockRepo *MockUserRepository,
	mockFamilyMemberRepo *MockFamilyMemberRepo,
	mockProvider *MockOAuthProvider,
	mockTokenGenerator *MockTokenGen,
) *authUsecase {
	// テスト用のJWT設定
	return &authUsecase{
		authRepo:       mockRepo,
		familyMemberRepo: mockFamilyMemberRepo,
		googleProvider: mockProvider,
		tokenGenerator: mockTokenGenerator,
	}
}

func TestAuthUsecase_InitiateGoogleLogin(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockOAuthProvider)
		wantAuthURL bool
		wantState   bool
		wantErr     bool
		wantErrType error
	}{
		{
			name: "正常にAuthURLとStateを生成",
			setupMock: func(m *MockOAuthProvider) {
				m.On("GetAuthURL", mock.AnythingOfType("string")).
					Return("https://accounts.google.com/o/oauth2/auth?client_id=test")
			},
			wantAuthURL: true,
			wantState:   true,
			wantErr:     false,
		},
		{
			name: "OAuth設定が不正でエラー",
			setupMock: func(m *MockOAuthProvider) {
				m.On("GetAuthURL", mock.AnythingOfType("string")).
					Return("")
			},
			wantAuthURL: false,
			wantState:   false, // エラーの場合はstateも生成されない
			wantErr:     true,
			wantErrType: &pkgerrors.InternalError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockProvider := new(MockOAuthProvider)
			tt.setupMock(mockProvider)

			usecase := setupAuthUsecase(t, mockRepo, nil, mockProvider, nil)

			authURL, state, err := usecase.InitiateGoogleLogin()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrType != nil {
					assert.IsType(t, tt.wantErrType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.wantAuthURL {
				assert.NotEmpty(t, authURL)
				assert.Contains(t, authURL, "accounts.google.com")
			} else {
				assert.Empty(t, authURL)
			}

			if tt.wantState {
				assert.NotEmpty(t, state)
				// Base64エンコードされた文字列であることを確認
				assert.Greater(t, len(state), 20)
			}

			mockProvider.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_HandleGoogleCallback_NewUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockFamilyMemberRepo := new(MockFamilyMemberRepo)
	mockProvider := new(MockOAuthProvider)
	mockTokenGenerator := new(MockTokenGen)
	usecase := setupAuthUsecase(t, mockRepo, mockFamilyMemberRepo, mockProvider, mockTokenGenerator)

	ctx := context.Background()
	code := "test-auth-code"

	// モックの設定
	userInfo := &oauth.OAuthUserInfo{
		ProviderID: "google-12345",
		Email:      "newuser@example.com",
		Name:       "New User",
		Picture:    "https://example.com/photo.jpg",
	}

	mockProvider.On("ExchangeCode", ctx, code).Return(userInfo, nil)

	// 既存ユーザーなし（Provider IDで検索）
	mockRepo.On("GetUserByProviderID", ctx, domain.AuthProviderGoogle, "google-12345").
		Return(nil, nil)

	// 既存ユーザーなし（Emailで検索）
	mockRepo.On("GetUserByEmail", ctx, "newuser@example.com").
		Return(nil, nil)

	// 新規ユーザー作成
	mockRepo.On("CreateUser", ctx, mock.MatchedBy(func(user *domain.User) bool {
		return user.Email == "newuser@example.com" &&
			user.Name == "New User" &&
			user.Provider == domain.AuthProviderGoogle &&
			user.ProviderID == "google-12345"
	})).Return(&domain.User{
		ID:         uuid.New(),
		Email:      "newuser@example.com",
		Name:       "New User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-12345",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}, nil)
	mockFamilyMemberRepo.On("GetFamilyMemberByUserID", ctx, mock.AnythingOfType("uuid.UUID")).Return(nil, nil)
	mockTokenGenerator.On("GenerateToken", ctx, mock.AnythingOfType("uuid.UUID"), uuid.Nil, domain.RoleUnknown).Return("jwt-access-token", nil)

	// テスト実行
	isJoined, token, err := usecase.HandleGoogleCallback(ctx, code)

	// 検証
	require.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.False(t, isJoined)

	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockFamilyMemberRepo.AssertExpectations(t)
	mockTokenGenerator.AssertExpectations(t)
}

func TestAuthUsecase_HandleGoogleCallback_ExistingUser(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockOAuthProvider)
	mockFamilyMemberRepo := new(MockFamilyMemberRepo)
	mockTokenGenerator := new(MockTokenGen)
	usecase := setupAuthUsecase(t, mockRepo, mockFamilyMemberRepo, mockProvider, mockTokenGenerator)

	ctx := context.Background()
	code := "test-auth-code"

	existingUser := &domain.User{
		ID:         uuid.New(),
		Email:      "existing@example.com",
		Name:       "Existing User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-12345",
		CreatedAt:  time.Now().Add(-24 * time.Hour),
		UpdatedAt:  time.Now().Add(-24 * time.Hour),
	}

	userInfo := &oauth.OAuthUserInfo{
		ProviderID: "google-12345",
		Email:      "existing@example.com",
		Name:       "Existing User",
		Picture:    "https://example.com/photo.jpg",
	}

	mockProvider.On("ExchangeCode", ctx, code).Return(userInfo, nil)
	mockRepo.On("GetUserByProviderID", ctx, domain.AuthProviderGoogle, "google-12345").
		Return(existingUser, nil)
	mockFamilyMemberRepo.On("GetFamilyMemberByUserID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&domain.FamilyMember{}, nil)
	mockTokenGenerator.On("GenerateToken", ctx, mock.AnythingOfType("uuid.UUID"), uuid.Nil, domain.RoleUnknown).Return("jwt-access-token", nil)

	// テスト実行
	isJoined, token, err := usecase.HandleGoogleCallback(ctx, code)

	// 検証
	require.NoError(t, err)
	require.NotEmpty(t, token)
	assert.True(t, isJoined) // 既存ユーザーなので家族に参加済み
	// CreateUserが呼ばれていないことを確認
	mockRepo.AssertNotCalled(t, "CreateUser")
	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockFamilyMemberRepo.AssertExpectations(t)
	mockTokenGenerator.AssertExpectations(t)
}

func TestAuthUsecase_HandleGoogleCallback_EmailAlreadyExists(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockOAuthProvider)
	mockFamilyMemberRepo := new(MockFamilyMemberRepo)
	mockTokenGenerator := new(MockTokenGen)
	usecase := setupAuthUsecase(t, mockRepo, mockFamilyMemberRepo, mockProvider, mockTokenGenerator)

	ctx := context.Background()
	code := "test-auth-code"

	// 異なるプロバイダーで既に登録されているユーザー
	existingUser := &domain.User{
		ID:         uuid.New(),
		Email:      "duplicate@example.com",
		Name:       "Existing User",
		Provider:   domain.AuthProviderGoogle, // 既にGoogle登録済み
		ProviderID: "different-provider-id",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	userInfo := &oauth.OAuthUserInfo{
		ProviderID: "new-google-id",
		Email:      "duplicate@example.com",
		Name:       "New User",
		Picture:    "https://example.com/photo.jpg",
	}

	mockProvider.On("ExchangeCode", ctx, code).Return(userInfo, nil)

	// Provider IDで見つからない
	mockRepo.On("GetUserByProviderID", ctx, domain.AuthProviderGoogle, "new-google-id").
		Return(nil, nil)

	// Emailで既存ユーザーが見つかる
	mockRepo.On("GetUserByEmail", ctx, "duplicate@example.com").
		Return(existingUser, nil)

	// テスト実行
	_, _, err := usecase.HandleGoogleCallback(ctx, code)

	// 検証
	require.Error(t, err)

	var validationErr *pkgerrors.ValidationError
	assert.ErrorAs(t, err, &validationErr)
	assert.Contains(t, err.Error(), "already registered")

	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestAuthUsecase_HandleGoogleCallback_ExchangeCodeError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockOAuthProvider)
	mockFamilyMemberRepo := new(MockFamilyMemberRepo)
	mockTokenGenerator := new(MockTokenGen)
	usecase := setupAuthUsecase(t, mockRepo, mockFamilyMemberRepo, mockProvider, mockTokenGenerator)

	ctx := context.Background()
	code := "invalid-code"

	mockProvider.On("ExchangeCode", ctx, code).
		Return(nil, errors.New("invalid authorization code"))

	// テスト実行
	_, _, err := usecase.HandleGoogleCallback(ctx, code)

	// 検証
	require.Error(t, err)

	var validationErr *pkgerrors.ValidationError
	assert.ErrorAs(t, err, &validationErr)

	mockProvider.AssertExpectations(t)
	mockRepo.AssertNotCalled(t, "GetUserByProviderID")
	mockRepo.AssertNotCalled(t, "CreateUser")
}

func TestAuthUsecase_HandleGoogleCallback_CreateUserError(t *testing.T) {
	mockRepo := new(MockUserRepository)
	mockProvider := new(MockOAuthProvider)
	mockFamilyMemberRepo := new(MockFamilyMemberRepo)
	mockTokenGenerator := new(MockTokenGen)
	usecase := setupAuthUsecase(t, mockRepo, mockFamilyMemberRepo, mockProvider, mockTokenGenerator)

	ctx := context.Background()
	code := "test-auth-code"

	userInfo := &oauth.OAuthUserInfo{
		ProviderID: "google-12345",
		Email:      "newuser@example.com",
		Name:       "New User",
		Picture:    "https://example.com/photo.jpg",
	}

	mockProvider.On("ExchangeCode", ctx, code).Return(userInfo, nil)
	mockRepo.On("GetUserByProviderID", ctx, domain.AuthProviderGoogle, "google-12345").
		Return(nil, nil)
	mockRepo.On("GetUserByEmail", ctx, "newuser@example.com").
		Return(nil, nil)
	mockRepo.On("CreateUser", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil, errors.New("database error"))

	// テスト実行
	_, _, err := usecase.HandleGoogleCallback(ctx, code)

	// 検証
	require.Error(t, err)

	var internalErr *pkgerrors.InternalError
	assert.ErrorAs(t, err, &internalErr)

	mockProvider.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
	mockFamilyMemberRepo.AssertExpectations(t)
}