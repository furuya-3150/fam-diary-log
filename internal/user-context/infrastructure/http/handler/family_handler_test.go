package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFamilyController struct {
	mock.Mock
}

func (m *MockFamilyController) CreateFamily(ctx context.Context, req *controller_dto.CreateFamilyRequest, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockFamilyController) InviteMembers(ctx context.Context, req *controller_dto.InviteMembersRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockFamilyController) ApplyToFamily(ctx context.Context, req *controller_dto.ApplyRequest, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

type MockFamilyControllerForHandler struct{ mock.Mock }

func (m *MockFamilyControllerForHandler) CreateFamily(ctx context.Context, req *dto.CreateFamilyRequest, userID uuid.UUID) (*dto.FamilyResponse, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.FamilyResponse), args.Error(1)
}
func (m *MockFamilyControllerForHandler) InviteMembers(ctx context.Context, req *dto.InviteMembersRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

// func (m *MockFamilyControllerForHandler) ApplyToFamily(ctx context.Context, req *dto.ApplyRequest, userID uuid.UUID) (string, error) {
// 	args := m.Called(ctx, req, userID)
// 	if args.Get(0) == nil {
// 		return "", args.Error(1)
// 	}
// 	return args.String(0), args.Error(1)
// }

type MockFamilyUsecase struct {
	mock.Mock
}

func (m *MockFamilyUsecase) CreateFamily(ctx context.Context, name string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, name, userID)
	if args.Get(0) == "" {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func (m *MockFamilyUsecase) InviteMembers(ctx context.Context, in usecase.InviteMembersInput) error {
	args := m.Called(ctx, in)
	return args.Error(0)
}

func (m *MockFamilyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, token, userID)
	if args.Get(0) == "" {
		return "", args.Error(1)
	}
	return args.Get(0).(string), args.Error(1)
}

func TestFamilyHandler_CreateFamily_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{fc: mockController, fu: mockUsecase}

	reqBody := &dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	userId := uuid.New()
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userId)
	c.SetRequest(req.WithContext(ctx))

	mockUsecase.On("CreateFamily", mock.Anything, mock.Anything, mock.Anything).Return("test-token-123", nil)

	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// Cookie検証
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	found := false
	for _, ck := range cookies {
		if ck.Name == auth.AuthCookieName {
			found = true
			require.Equal(t, "test-token-123", ck.Value)
			require.Equal(t, "/", ck.Path)
			require.True(t, ck.HttpOnly)
		}
	}
	require.True(t, found, "Auth cookie not found")
}

func TestFamilyHandler_CreateFamily_Error(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: mockController,
		fu: mockUsecase,
	}

	reqBody := &controller_dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, uuid.New())
	c.SetRequest(req.WithContext(ctx))

	mockUsecase.On("CreateFamily", mock.Anything, mock.Anything, mock.Anything).Return("", errors.New("failed to create family"))

	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestFamilyHandler_CreateFamily_BadRequest_NoUser(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: mockController,
		fu: mockUsecase,
	}

	reqBody := &controller_dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// UserIDなし
	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_InviteMembers_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: mockController,
		fu: mockUsecase,
	}

	familyID := uuid.New()
	userID := uuid.New()
	reqBody := &controller_dto.InviteMembersRequest{
		Emails: []string{"test1@example.com", "test2@example.com"},
	}
	b, _ := json.Marshal(reqBody)
	url := "/families/invitations"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	ctx = context.WithValue(ctx, auth.ContextKeyFamilyID, familyID)
	c.SetRequest(req.WithContext(ctx))

	mockController.On("InviteMembers", mock.Anything, mock.AnythingOfType("*dto.InviteMembersRequest")).Return(nil)

	_ = h.InviteMembers(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
	mockController.AssertExpectations(t)
}

func TestFamilyHandler_InviteMembers_BadRequest_InvalidEmail(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyController)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	familyID := uuid.New()
	// 不正なメール形式
	reqBody := &controller_dto.InviteMembersRequest{
		Emails: []string{"badmail"},
	}
	b, _ := json.Marshal(reqBody)
	url := "/families/invitations"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, uuid.New())
	ctx = context.WithValue(ctx, auth.ContextKeyFamilyID, familyID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.InviteMembers(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_InviteMembers_BadRequest_NoFamilyID(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	reqBody := &controller_dto.InviteMembersRequest{
		Emails: []string{"test@example.com"},
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families/invitations", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// FamilyIDなし、UserIDのみ
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, uuid.New())
	c.SetRequest(req.WithContext(ctx))

	err := h.InviteMembers(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_InviteMembers_BadRequest_NoUserID(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	reqBody := &controller_dto.InviteMembersRequest{
		Emails: []string{"test@example.com"},
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families/invitations", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// UserIDなし、FamilyIDのみ
	ctx := context.WithValue(context.Background(), auth.ContextKeyFamilyID, uuid.New())
	c.SetRequest(req.WithContext(ctx))

	err := h.InviteMembers(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_ApplyToFamily_Success(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: new(MockFamilyControllerForHandler),
		fu: mockUsecase,
	}

	userID := uuid.New()
	reqBody := &controller_dto.ApplyRequest{Token: "tok-123"}
	b, _ := json.Marshal(reqBody)
	url := "/invitations/apply"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mockUsecase.On("ApplyToFamily", mock.Anything, "tok-123", userID).Return("test-token", nil)

	_ = h.ApplyToFamily(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
	mockUsecase.AssertExpectations(t)
}

func TestFamilyHandler_ApplyToFamily_BadRequest_NoToken(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	userID := uuid.New()
	reqBody := &controller_dto.ApplyRequest{Token: ""}
	b, _ := json.Marshal(reqBody)
	url := "/invitations/apply"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.ApplyToFamily(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_ApplyToFamily_ControllerError(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: new(MockFamilyControllerForHandler),
		fu: mockUsecase,
	}

	userID := uuid.New()
	reqBody := &controller_dto.ApplyRequest{Token: "tok-err"}
	b, _ := json.Marshal(reqBody)
	url := "/invitations/apply"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mockUsecase.On("ApplyToFamily", mock.Anything, "tok-err", userID).Return("", assert.AnError)

	_ = h.ApplyToFamily(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	mockUsecase.AssertExpectations(t)
}

func TestFamilyHandler_ApplyToFamily_BadRequest_NoUserID(t *testing.T) {
	e := echo.New()
	h := &familyHandler{
		fc: new(MockFamilyControllerForHandler),
		fu: new(MockFamilyUsecase),
	}

	reqBody := &controller_dto.ApplyRequest{Token: "tok-123"}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/invitations/apply", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// UserIDなし
	err := h.ApplyToFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_ApplyToFamily_Success_WithCookie(t *testing.T) {
	e := echo.New()
	mockUsecase := new(MockFamilyUsecase)
	h := &familyHandler{
		fc: new(MockFamilyControllerForHandler),
		fu: mockUsecase,
	}

	userID := uuid.New()
	reqBody := &controller_dto.ApplyRequest{Token: "tok-123"}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/invitations/apply", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mockUsecase.On("ApplyToFamily", mock.Anything, "tok-123", userID).Return("family-token-456", nil)

	err := h.ApplyToFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// Cookie検証
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	found := false
	for _, ck := range cookies {
		if ck.Name == auth.AuthCookieName {
			found = true
			require.Equal(t, "family-token-456", ck.Value)
			require.Equal(t, "/", ck.Path)
			require.True(t, ck.HttpOnly)
		}
	}
	require.True(t, found, "Auth cookie not found")
	mockUsecase.AssertExpectations(t)
}
