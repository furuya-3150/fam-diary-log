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
