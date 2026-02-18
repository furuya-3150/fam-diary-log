package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandlerGetFamilyMembersSuccess(t *testing.T) {
	e := echo.New()

	// モックusecaseの準備
	mockUsecase := &mockUserUsecase{
		users: []*domain.User{
			{
				ID:       uuid.New(),
				Email:    "user1@example.com",
				Name:     "User One",
				Provider: domain.AuthProviderGoogle,
			},
			{
				ID:       uuid.New(),
				Email:    "user2@example.com",
				Name:     "User Two",
				Provider: domain.AuthProviderGoogle,
			},
		},
	}

	handler := &userHandler{
		userUsecase: mockUsecase,
	}

	familyID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/families/me/members", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// コンテキストにfamilyIDをセット
	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, auth.ContextKeyFamilyID, familyID)
	c.SetRequest(c.Request().WithContext(ctx))

	// 実行
	err := handler.GetFamilyMembers(c)
	require.NoError(t, err)

	// レスポンスの検証
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	data, ok := response["data"].([]interface{})
	require.True(t, ok)
	assert.Len(t, data, 2)
}

func TestUserHandlerGetFamilyMembersWithFieldSelection(t *testing.T) {
	e := echo.New()

	mockUsecase := &mockUserUsecase{
		users: []*domain.User{
			{
				ID:   uuid.New(),
				Name: "User One",
			},
		},
		capturedFields: nil,
	}

	handler := &userHandler{
		userUsecase: mockUsecase,
	}

	familyID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/families/me/members?fields=id,name", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := c.Request().Context()
	ctx = context.WithValue(ctx, auth.ContextKeyFamilyID, familyID)
	c.SetRequest(c.Request().WithContext(ctx))

	// 実行
	err := handler.GetFamilyMembers(c)
	require.NoError(t, err)

	// fieldsパラメータが正しくパースされたことを確認
	assert.Equal(t, []string{"id", "name"}, mockUsecase.capturedFields)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestUserHandlerGetFamilyMembersNoFamilyID(t *testing.T) {
	e := echo.New()

	mockUsecase := &mockUserUsecase{}
	handler := &userHandler{
		userUsecase: mockUsecase,
	}

	req := httptest.NewRequest(http.MethodGet, "/families/me/members", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// familyIDをセットしない
	err := handler.GetFamilyMembers(c)

	// エラーが返されるべき
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// モックusecase
type mockUserUsecase struct {
	users          []*domain.User
	capturedFields []string
	shouldError    bool
}

func (m *mockUserUsecase) GetFamilyMembers(ctx context.Context, familyID uuid.UUID, fields []string) ([]*domain.User, error) {
	m.capturedFields = fields
	if m.shouldError {
		return nil, assert.AnError
	}
	return m.users, nil
}

func (m *mockUserUsecase) EditUser(ctx context.Context, input *usecase.EditUserInput) (*domain.User, error) {
	return nil, nil
}

func (m *mockUserUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	return nil, nil
}
