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
	controller_dto "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockNotificationUsecase struct{ mock.Mock }

func (m *MockNotificationUsecase) GetNotificationSetting(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error) {
	args := m.Called(ctx, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.NotificationSetting), args.Error(1)
}

func (m *MockNotificationUsecase) UpdateNotificationSetting(ctx context.Context, setting *domain.NotificationSetting) error {
	args := m.Called(ctx, setting)
	return args.Error(0)
}

func TestNotificationHandler_GetNotificationSetting_Success(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()
	familyID := uuid.New()
	setting := &domain.NotificationSetting{UserID: userID, FamilyID: familyID, PostCreatedEnabled: true}

	req := httptest.NewRequest(http.MethodGet, "/settings/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	mu.On("GetNotificationSetting", mock.Anything, userID, familyID).Return(setting, nil)

	err := h.GetNotificationSetting(c)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, rec.Code)
	mu.AssertCalled(t, "GetNotificationSetting", mock.Anything, userID, familyID)
	body := rec.Body.Bytes()
	require.NotEmpty(t, body)
}

func TestNotificationHandler_GetNotificationSetting_BadRequest_NoUser(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	familyID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/settings/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.GetNotificationSetting(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandler_GetNotificationSetting_BadRequest_NoFamily(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/settings/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.GetNotificationSetting(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandler_GetNotificationSetting_UsecaseError(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()
	familyID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/settings/notifications", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	mu.On("GetNotificationSetting", mock.Anything, userID, familyID).Return(nil, errors.New("db err"))

	_ = h.GetNotificationSetting(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestNotificationHandler_UpdateNotificationSetting_Success(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()
	familyID := uuid.New()
	reqBody := &controller_dto.NotificationSettingRequest{PostCreatedEnabled: false}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/settings/notifications", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	mu.On("UpdateNotificationSetting", mock.Anything, mock.MatchedBy(func(ns *domain.NotificationSetting) bool {
		return ns.UserID == userID && ns.FamilyID == familyID && ns.PostCreatedEnabled == false
	})).Return(nil)

	_ = h.UpdateNotificationSetting(c)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestNotificationHandler_UpdateNotificationSetting_BadRequest_InvalidBody(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()
	familyID := uuid.New()
	// invalid JSON
	req := httptest.NewRequest(http.MethodPut, "/settings/notifications", bytes.NewReader([]byte("{")))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.UpdateNotificationSetting(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandler_UpdateNotificationSetting_BadRequest_NoUser(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	familyID := uuid.New()
	reqBody := &controller_dto.NotificationSettingRequest{PostCreatedEnabled: false}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/settings/notifications", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	_ = h.UpdateNotificationSetting(c)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestNotificationHandler_UpdateNotificationSetting_UsecaseError(t *testing.T) {
	e := echo.New()
	mu := new(MockNotificationUsecase)
	h := NewNotificationHandler(mu)

	userID := uuid.New()
	familyID := uuid.New()
	reqBody := &controller_dto.NotificationSettingRequest{PostCreatedEnabled: true}
	b, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/settings/notifications", bytes.NewReader(b))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	ctx := context.WithValue(context.Background(), "user_id", userID)
	ctx = context.WithValue(ctx, "family_id", familyID)
	c.SetRequest(req.WithContext(ctx))

	mu.On("UpdateNotificationSetting", mock.Anything, mock.Anything).Return(errors.New("upsert err"))

	_ = h.UpdateNotificationSetting(c)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
