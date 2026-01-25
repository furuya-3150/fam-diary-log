package controller

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthUsecase は AuthUsecase のモック
type MockAuthUsecase struct {
	mock.Mock
}

func (m *MockAuthUsecase) InitiateGoogleLogin() (authURL string, state string, err error) {
	args := m.Called()
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAuthUsecase) HandleGoogleCallback(ctx context.Context, code string) (*domain.AuthResponse, error) {
	args := m.Called(ctx, code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResponse), args.Error(1)
}

func TestAuthController_InitiateGoogleLogin(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func(*MockAuthUsecase)
		wantAuthURL string
		wantState   string
		wantErr     bool
	}{
		{
			name: "正常にAuthURLとStateを返す",
			setupMock: func(m *MockAuthUsecase) {
				m.On("InitiateGoogleLogin").
					Return("https://accounts.google.com/o/oauth2/auth?client_id=test", "random-state-123", nil)
			},
			wantAuthURL: "https://accounts.google.com/o/oauth2/auth?client_id=test",
			wantState:   "random-state-123",
			wantErr:     false,
		},
		{
			name: "UsecaseでエラーがOccur",
			setupMock: func(m *MockAuthUsecase) {
				m.On("InitiateGoogleLogin").
					Return("", "", errors.New("failed to generate state"))
			},
			wantAuthURL: "",
			wantState:   "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAuthUsecase)
			tt.setupMock(mockUsecase)

			controller := NewAuthController(mockUsecase)

			authURL, state, err := controller.InitiateGoogleLogin()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantAuthURL, authURL)
				assert.Equal(t, tt.wantState, state)
			}

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAuthController_HandleGoogleCallback(t *testing.T) {
	tests := []struct {
		name      string
		code      string
		setupMock func(*MockAuthUsecase, context.Context, string)
		wantErr   bool
		validate  func(*testing.T, *MockAuthUsecase, error)
	}{
		{
			name: "正常に認証レスポンスを返す",
			code: "valid-auth-code",
			setupMock: func(m *MockAuthUsecase, ctx context.Context, code string) {
				authResp := &domain.AuthResponse{
					User: &domain.User{
						ID:         uuid.New(),
						Email:      "test@example.com",
						Name:       "Test User",
						Provider:   domain.AuthProviderGoogle,
						ProviderID: "google-12345",
						CreatedAt:  time.Now(),
						UpdatedAt:  time.Now(),
					},
					AccessToken: "jwt-access-token",
					ExpiresIn:   3600,
				}
				m.On("HandleGoogleCallback", ctx, code).Return(authResp, nil)
			},
			wantErr: false,
			validate: func(t *testing.T, m *MockAuthUsecase, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "無効なコードでエラー",
			code: "invalid-code",
			setupMock: func(m *MockAuthUsecase, ctx context.Context, code string) {
				m.On("HandleGoogleCallback", ctx, code).
					Return(nil, errors.New("invalid authorization code"))
			},
			wantErr: true,
			validate: func(t *testing.T, m *MockAuthUsecase, err error) {
				assert.Error(t, err)
			},
		},
		{
			name: "空のコードでエラー",
			code: "",
			setupMock: func(m *MockAuthUsecase, ctx context.Context, code string) {
				m.On("HandleGoogleCallback", ctx, code).
					Return(nil, errors.New("code cannot be empty"))
			},
			wantErr: true,
			validate: func(t *testing.T, m *MockAuthUsecase, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockAuthUsecase)
			ctx := context.Background()
			tt.setupMock(mockUsecase, ctx, tt.code)

			controller := NewAuthController(mockUsecase)

			resp, err := controller.HandleGoogleCallback(ctx, tt.code)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, resp)
				assert.NotNil(t, resp.User)
				assert.NotEmpty(t, resp.AccessToken)
				assert.Equal(t, int64(3600), resp.ExpiresIn)
			}

			tt.validate(t, mockUsecase, err)
			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestAuthController_toAuthResponse(t *testing.T) {
	mockUsecase := new(MockAuthUsecase)
	controller := NewAuthController(mockUsecase).(*authController)

	userID := uuid.New()
	now := time.Now()

	domainResp := &domain.AuthResponse{
		User: &domain.User{
			ID:         userID,
			Email:      "test@example.com",
			Name:       "Test User",
			Provider:   domain.AuthProviderGoogle,
			ProviderID: "google-12345",
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		AccessToken: "test-access-token",
		ExpiresIn:   3600,
	}

	dtoResp := controller.toAuthResponse(domainResp)

	assert.NotNil(t, dtoResp)
	assert.NotNil(t, dtoResp.User)
	assert.Equal(t, userID, dtoResp.User.ID)
	assert.Equal(t, "test@example.com", dtoResp.User.Email)
	assert.Equal(t, "Test User", dtoResp.User.Name)
	assert.Equal(t, "google", dtoResp.User.Provider)
	assert.Equal(t, now, dtoResp.User.CreatedAt)
	assert.Equal(t, "test-access-token", dtoResp.AccessToken)
	assert.Equal(t, int64(3600), dtoResp.ExpiresIn)
}
