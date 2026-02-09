package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ...existing imports...

type MockUserController struct {
	mock.Mock
}

func (m *MockUserController) EditProfile(ctx context.Context, req *controller_dto.EditUserRequest) (*controller_dto.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller_dto.UserResponse), args.Error(1)
}

func (m *MockUserController) GetProfile(ctx context.Context, userID uuid.UUID) (*controller_dto.UserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller_dto.UserResponse), args.Error(1)
}

func TestUserHandler_EditProfile_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	reqBody := &controller_dto.EditUserRequest{
		ID:    uuid.Nil,
		Name:  "Alice",
		Email: "alice@example.com",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), auth.ContextKeyFamilyID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, uuid.New())
	c.SetRequest(req.WithContext(ctx))

	expected := &controller_dto.UserResponse{
		ID:    reqBody.ID,
		Email: reqBody.Email,
		Name:  reqBody.Name,
	}
	mockController.On("EditProfile", mock.Anything, mock.MatchedBy(func(r *controller_dto.EditUserRequest) bool {
		return r.Name == reqBody.Name && r.Email == reqBody.Email
	})).Return(expected, nil)

	err := h.EditProfile(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var got controller_dto.UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Equal(t, expected.Email, got.Email)
	require.Equal(t, expected.Name, got.Name)
	mockController.AssertExpectations(t)
}

func TestUserHandler_EditProfile_BadRequest(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader([]byte("invalid json")))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := h.EditProfile(c)
	require.NoError(t, err)
	// コントローラーは呼ばれない
	mockController.AssertNotCalled(t, "EditProfile", mock.Anything, mock.Anything)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, jsonErr)
	require.Contains(t, resp, "message")
	require.NotEmpty(t, resp["message"])
}

func TestUserHandler_EditProfile_ControllerError(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	reqBody := &controller_dto.EditUserRequest{
		ID:    controller_dto.EditUserRequest{}.ID,
		Name:  "Alice",
		Email: "alice@example.com",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/users/me", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), auth.ContextKeyFamilyID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, uuid.New())
	c.SetRequest(req.WithContext(ctx))

	mockController.On("EditProfile", mock.Anything, mock.Anything).Return(nil, errors.New("controller error"))

	err := h.EditProfile(c)
	require.NoError(t, err)
	mockController.AssertExpectations(t)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, jsonErr)
	require.Contains(t, resp, "message")
	require.NotEmpty(t, resp["message"])
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// contextにuser_idをセット
	c.SetRequest(req.WithContext(context.WithValue(req.Context(), auth.ContextKeyUserID, userID)))

	expected := &controller_dto.UserResponse{
		ID:    userID,
		Email: "test@example.com",
	}
	mockController.On("GetProfile", mock.Anything, userID).Return(expected, nil)

	err := h.GetProfile(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)

	var got controller_dto.UserResponse
	err = json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err)
	require.Equal(t, expected.Email, got.Email)
	mockController.AssertExpectations(t)
}

func TestUserHandler_GetProfile_BadRequest(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// contextにuser_idをセットしない（nil）

	err := h.GetProfile(c)
	require.NoError(t, err)
	mockController.AssertNotCalled(t, "GetProfile", mock.Anything, mock.Anything)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, jsonErr)
	require.Contains(t, resp, "message")
	require.NotEmpty(t, resp["message"])
}

func TestUserHandler_GetProfile_ControllerError(t *testing.T) {
	e := echo.New()
	mockController := new(MockUserController)
	h := &userHandler{userController: mockController}

	userID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetRequest(req.WithContext(context.WithValue(req.Context(), auth.ContextKeyUserID, userID)))

	mockController.On("GetProfile", mock.Anything, userID).Return(nil, errors.New("controller error"))

	err := h.GetProfile(c)
	require.NoError(t, err)
	mockController.AssertExpectations(t)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, jsonErr)
	require.Contains(t, resp, "message")
	require.NotEmpty(t, resp["message"])
}
