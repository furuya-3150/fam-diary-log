package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
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

func (m *MockFamilyController) ApplyToFamily(ctx context.Context, req *controller_dto.ApplyRequest, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}

func (m *MockFamilyController) RespondToJoinRequest(ctx context.Context, req *controller_dto.RespondJoinRequestRequest, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}

func (m *MockFamilyController) JoinFamily(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
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
func (m *MockFamilyControllerForHandler) ApplyToFamily(ctx context.Context, req *dto.ApplyRequest, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}
func (m *MockFamilyControllerForHandler) RespondToJoinRequest(ctx context.Context, req *dto.RespondJoinRequestRequest, userID uuid.UUID) error {
	args := m.Called(ctx, req, userID)
	return args.Error(0)
}
func (m *MockFamilyControllerForHandler) JoinFamily(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

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

func (m *MockFamilyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) error {
	args := m.Called(ctx, token, userID)
	return args.Error(0)
}

func (m *MockFamilyUsecase) RespondToJoinRequest(ctx context.Context, requestID uuid.UUID, status domain.JoinRequestStatus, responderUserID uuid.UUID) error {
    args := m.Called(ctx, requestID, status, responderUserID)
    return args.Error(0)
}

func (m *MockFamilyUsecase) JoinFamilyIfApproved(ctx context.Context, userID uuid.UUID) (string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
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

	// expected := &dto.FamilyResponse{
	// 	ID:   uuid.New(),
	// 	Name: reqBody.Name,
	// }
	// mockController.On("CreateFamily", mock.Anything, reqBody.Name, userId).Return(expected, nil)
	mockUsecase.On("CreateFamily", mock.Anything, mock.Anything, mock.Anything).Return("token", nil)

	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)
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

func TestFamilyHandler_InviteMembers_BadRequest(t *testing.T) {
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

func TestFamilyHandler_ApplyToFamily_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
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

	mockController.On("ApplyToFamily", mock.Anything, mock.AnythingOfType("*dto.ApplyRequest"), mock.AnythingOfType("uuid.UUID")).Return(nil)

	_ = h.ApplyToFamily(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
	mockController.AssertExpectations(t)
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
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
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

	mockController.On("ApplyToFamily", mock.Anything, mock.AnythingOfType("*dto.ApplyRequest"), mock.AnythingOfType("uuid.UUID")).Return(assert.AnError)

	_ = h.ApplyToFamily(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	mockController.AssertExpectations(t)
}

func TestFamilyHandler_RespondToJoinRequest_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	userID := uuid.New()
	reqBody := &controller_dto.RespondJoinRequestRequest{
		ID:     uuid.New(),
		Status: int(domain.JoinRequestStatusApproved),
	}
	b, _ := json.Marshal(reqBody)
	url := "/families/respond"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mockController.On("RespondToJoinRequest", mock.Anything, mock.AnythingOfType("*dto.RespondJoinRequestRequest"), mock.AnythingOfType("uuid.UUID")).Return(nil)

	_ = h.RespondToJoinRequest(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
	mockController.AssertExpectations(t)
}

func TestFamilyHandler_RespondToJoinRequest_BadRequest_NoUser(t *testing.T) {
	// e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	reqBody := &controller_dto.RespondJoinRequestRequest{
		ID:     uuid.New(),
		Status: int(domain.JoinRequestStatusApproved),
	}
	b, _ := json.Marshal(reqBody)
	url := "/families/respond"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// no user_id in context
	_ = h.RespondToJoinRequest(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_RespondToJoinRequest_ControllerError(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyControllerForHandler)
	h := &familyHandler{
		fc: mockController,
		fu: new(MockFamilyUsecase),
	}

	userID := uuid.New()
	reqBody := &controller_dto.RespondJoinRequestRequest{
		ID:     uuid.New(),
		Status: int(domain.JoinRequestStatusRejected),
	}
	b, _ := json.Marshal(reqBody)
	url := "/families/respond"
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mockController.On("RespondToJoinRequest", mock.Anything, mock.AnythingOfType("*dto.RespondJoinRequestRequest"), mock.AnythingOfType("uuid.UUID")).Return(assert.AnError)

	_ = h.RespondToJoinRequest(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	mockController.AssertExpectations(t)
}

func TestFamilyHandler_JoinFamily_SetsCookie_Success(t *testing.T) {
	// e := echo.New()
	mc := new(MockFamilyControllerForHandler)
	h := NewFamilyHandler(mc, new(MockFamilyUsecase))

	userID := uuid.New()
	token := "signed-token"

	req := httptest.NewRequest(http.MethodPost, "/families/join", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mc.On("JoinFamily", mock.Anything, userID).Return(token, nil)

	err := h.JoinFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, rec.Code)
	cookies := rec.Result().Cookies()
	require.NotEmpty(t, cookies)
	found := false
	for _, ck := range cookies {
		if ck.Name == auth.FamilyCookieName {
			found = true
			require.Equal(t, token, ck.Value)
			require.Equal(t, "/", ck.Path)
		}
	}
	require.True(t, found)
}

func TestFamilyHandler_JoinFamily_BadRequest_NoUser(t *testing.T) {
	// e := echo.New()
	mc := new(MockFamilyControllerForHandler)
	h := NewFamilyHandler(mc, new(MockFamilyUsecase))

	req := httptest.NewRequest(http.MethodPost, "/families/join", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	// no user_id in context
	err := h.JoinFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestFamilyHandler_JoinFamily_UsecaseError(t *testing.T) {
	// e := echo.New()
	mc := new(MockFamilyControllerForHandler)
	h := NewFamilyHandler(mc, new(MockFamilyUsecase))

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/families/join", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	ctx := context.WithValue(context.Background(), auth.ContextKeyUserID, userID)
	c.SetRequest(req.WithContext(ctx))

	mc.On("JoinFamily", mock.Anything, userID).Return("", errors.New("failed"))

	err := h.JoinFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
