package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockFamilyController struct {
	mock.Mock
}

func (m *MockFamilyController) CreateFamily(ctx context.Context, req *controller_dto.CreateFamilyRequest, userID uuid.UUID) (*controller_dto.FamilyResponse, error) {
	args := m.Called(ctx, req, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*controller_dto.FamilyResponse), args.Error(1)
}

func (m *MockFamilyController) InviteMembers(ctx context.Context, req *controller_dto.InviteMembersRequest) (error) {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func TestFamilyHandler_CreateFamily_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyController)
	h := &familyHandler{familyController: mockController}

	reqBody := &controller_dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), "user_id", uuid.New())
	c.SetRequest(req.WithContext(ctx))

	expected := &controller_dto.FamilyResponse{
		ID:   uuid.New(),
		Name: reqBody.Name,
	}
	mockController.On("CreateFamily", mock.Anything, reqBody, mock.AnythingOfType("uuid.UUID")).Return(expected, nil)

	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestFamilyHandler_CreateFamily_Error(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyController)
	h := &familyHandler{familyController: mockController}

	reqBody := &controller_dto.CreateFamilyRequest{
		Name: "TestFamily",
	}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/families", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctx := context.WithValue(context.Background(), "user_id", uuid.New())
	c.SetRequest(req.WithContext(ctx))

	mockController.On("CreateFamily", mock.Anything, reqBody, mock.AnythingOfType("uuid.UUID")).Return(nil, assert.AnError)

	err := h.CreateFamily(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestFamilyHandler_InviteMembers_Success(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyController)
	h := &familyHandler{familyController: mockController}

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
	
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	mockController.On("InviteMembers", mock.Anything, mock.AnythingOfType("*dto.InviteMembersRequest")).Return(nil)

	_ = h.InviteMembers(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
	mockController.AssertExpectations(t)
}

func TestFamilyHandler_InviteMembers_BadRequest(t *testing.T) {
	e := echo.New()
	mockController := new(MockFamilyController)
	h := &familyHandler{familyController: mockController}

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

	ctx := context.WithValue(context.Background(), "user_id", uuid.New())
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.InviteMembers(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
