package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	infctx "github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/context"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

type MockDiaryController struct {
	mock.Mock
}

func (m *MockDiaryController) Create(ctx context.Context, d *domain.Diary) (*dto.DiaryResponse, error) {
	args := m.Called(ctx, d)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.DiaryResponse), args.Error(1)
}

func (m *MockDiaryController) List(ctx context.Context, familyID uuid.UUID) ([]dto.DiaryResponse, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.DiaryResponse), args.Error(1)
}

// create diary successfully
func TestDiaryHandler_Create_Success(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()
	diaryID := uuid.New()

	requestBody := &domain.Diary{
		Title:   "Test Diary",
		Content: "This is a test diary",
	}

	expectedResponse := &dto.DiaryResponse{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    requestBody.Title,
		Content:  requestBody.Content,
	}

	mockController.On("Create", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx.Value(infctx.FamilyIDKey) == familyID && ctx.Value(infctx.UserIDKey) == userID
	}), mock.MatchedBy(func(d *domain.Diary) bool {
		return d.Title == requestBody.Title && d.Content == requestBody.Content
	})).Return(expectedResponse, nil)

	// Create request
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Create(c)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify response status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify response body
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if _, ok := response["data"]; !ok {
		t.Error("response should contain 'data' field")
	}

	mockController.AssertExpectations(t)
}

// create diary with validation error
func TestDiaryHandler_Create_ValidationError(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	requestBody := &domain.Diary{
		Title:   "",
		Content: "Valid content",
	}

	validationErr := &errors.ValidationError{Message: "title cannot be empty"}
	mockController.On("Create", mock.Anything, mock.Anything).Return(nil, validationErr)

	// Create request
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Create(c)
	if err != nil {
		// Error is expected
		t.Logf("expected error: %v", err)
	}

	// Verify response status code is 400 Bad Request
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Verify response body contains error message
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	mockController.AssertExpectations(t)
}

// create diary with internal error
func TestDiaryHandler_Create_InternalError(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	requestBody := &domain.Diary{
		Title:   "Test Diary",
		Content: "This is a test diary",
	}

	internalErr := &errors.InternalError{Message: "database error"}
	mockController.On("Create", mock.Anything, mock.Anything).Return(nil, internalErr)

	// Create request
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Create(c)
	if err != nil {
		// Error is expected
		t.Logf("expected error: %v", err)
	}

	// Verify response status code is 500 Internal Server Error
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	// Verify response body contains error message
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	mockController.AssertExpectations(t)
}

// TestDiaryHandler_Create_BindError tests binding error handling
func TestDiaryHandler_Create_BindError(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.Create(c)
	if err != nil {
		// Error is expected
		t.Logf("expected error: %v", err)
	}

	// Verify response status code is 400 Bad Request (due to bind error)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Verify response body contains error message
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Controller should not be called on bind error
	mockController.AssertNotCalled(t, "Create")
}

// ------------
// List Diaries
// ------------

// list diaries successfully
func TestDiaryHandler_List_Success(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()
	diaryID1 := uuid.New()
	diaryID2 := uuid.New()

	expectedResponses := []dto.DiaryResponse{
		{
			ID:       diaryID1,
			UserID:   userID,
			FamilyID: familyID,
			Title:    "Test Diary 1",
			Content:  "Content 1",
		},
		{
			ID:       diaryID2,
			UserID:   userID,
			FamilyID: familyID,
			Title:    "Test Diary 2",
			Content:  "Content 2",
		},
	}

	mockController.On("List", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx.Value(infctx.FamilyIDKey) == familyID
	}), familyID).Return(expectedResponses, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/diaries", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.List(c)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	// Verify response status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify response body
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	mockController.AssertExpectations(t)
}

// list diaries with error
func TestDiaryHandler_List_Error(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	internalErr := &errors.InternalError{Message: "database error"}
	mockController.On("List", mock.Anything, familyID).Return(nil, internalErr)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/diaries", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.List(c)
	if err != nil {
		t.Logf("expected error: %v", err)
	}

	mockController.AssertExpectations(t)
}
