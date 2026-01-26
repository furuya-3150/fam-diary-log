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

func TestUserHandler_EditProfile_Success(t *testing.T) {
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
	require.Error(t, err)
	// コントローラーは呼ばれない
	mockController.AssertNotCalled(t, "EditProfile", mock.Anything, mock.Anything)
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

	mockController.On("EditProfile", mock.Anything, mock.Anything).Return(nil, errors.New("controller error"))

	err := h.EditProfile(c)
	require.Error(t, err)
	require.Equal(t, http.StatusInternalServerError, err.(*echo.HTTPError).Code)
	mockController.AssertExpectations(t)
}
