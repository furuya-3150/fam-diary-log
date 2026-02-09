package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
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

func (m *MockDiaryController) List(ctx context.Context, familyID uuid.UUID, targetDate string) ([]dto.DiaryResponse, error) {
	args := m.Called(ctx, familyID, targetDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.DiaryResponse), args.Error(1)
}

func (m *MockDiaryController) GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error) {
	args := m.Called(ctx, familyID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockDiaryController) GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*dto.StreakResponse, error) {
	args := m.Called(ctx, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.StreakResponse), args.Error(1)
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
		return ctx.Value(auth.ContextKeyFamilyID) == familyID && ctx.Value(auth.ContextKeyUserID) == userID
	}), mock.MatchedBy(func(d *domain.Diary) bool {
		return d.Title == requestBody.Title && d.Content == requestBody.Content
	})).Return(expectedResponse, nil)

	// Create request
	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/families/me/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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
	req := httptest.NewRequest(http.MethodPost, "/families/me/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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
	req := httptest.NewRequest(http.MethodPost, "/families/me/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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
	req := httptest.NewRequest(http.MethodPost, "/families/me/diaries", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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
		return ctx.Value(auth.ContextKeyFamilyID) == familyID
	}), familyID, mock.Anything).Return(expectedResponses, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/families/me/diaries?target_date=2026-01-01", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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
	mockController.On("List", mock.Anything, familyID, mock.Anything).Return(nil, internalErr)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/families/me/diaries?target_date=2026-01-01", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
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

// missing target_date should return 400 and controller should not be called
func TestDiaryHandler_List_MissingTargetDate(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()

	// No expectation for List: it should not be called

	// Create request without target_date
	req := httptest.NewRequest(http.MethodGet, "/families/me/diaries", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
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

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Ensure controller.List was not called
	mockController.AssertNotCalled(t, "List")
}

// invalid target_date format should return 400 and controller should not be called
func TestDiaryHandler_List_InvalidTargetDate(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()

	// Create request with invalid target_date
	req := httptest.NewRequest(http.MethodGet, "/families/me/diaries?target_date=invalid-date", nil)

	// Set context values
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
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

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Ensure controller.List was not called
	mockController.AssertNotCalled(t, "List")
}

// TestDiaryHandler_GetCount_Success tests successful count retrieval
func TestDiaryHandler_GetCount_Success(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()

	mockController.On("GetCount", mock.Anything, familyID, "2026", "01").Return(5, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/families/me/diaries/count?year=2026&month=01", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.GetCount(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	mockController.AssertExpectations(t)
}

// TestDiaryHandler_GetCount_InvalidParams tests invalid parameters
func TestDiaryHandler_GetCount_InvalidParams(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()

	mockController.On("GetCount", mock.Anything, familyID, "2026", "13").Return(0, &errors.ValidationError{Message: "invalid month"})

	// Create HTTP request
	req := httptest.NewRequest("GET", "/families/me/diaries/count?year=2026&month=13", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.GetCount(c)
	if err != nil {
		t.Logf("expected error: %v", err)
	}

	mockController.AssertExpectations(t)
}

// TestDiaryHandler_GetCount_MissingParams tests missing query parameters
func TestDiaryHandler_GetCount_MissingParams(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()

	testCases := []struct {
		name string
		url  string
	}{
		{"missing year", "/families/me/diaries/count?month=01"},
		{"missing month", "/families/me/diaries/count?year=2026"},
		{"missing both", "/families/me/diaries/count"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest("GET", tc.url, nil)
			ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
			req = req.WithContext(ctx)

			// Create response writer
			rec := httptest.NewRecorder()

			// Create Echo context
			e := echo.New()
			c := e.NewContext(req, rec)

			// Call handler
			err := handler.GetCount(c)
			if err != nil {
				t.Logf("expected error: %v", err)
			}

			// Should return 400 Bad Request
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
			}
		})
	}

	// Controller should never be called
	mockController.AssertNotCalled(t, "GetCount")
}

// ============================================
// GetStreak Tests
// ============================================

// TestDiaryHandler_GetStreak_Success tests successful streak retrieval
func TestDiaryHandler_GetStreak_Success(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	expectedResponse := &dto.StreakResponse{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 5,
		LastPostDate:  nil,
	}

	mockController.On("GetStreak", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx.Value(auth.ContextKeyFamilyID) == familyID && ctx.Value(auth.ContextKeyUserID) == userID
	}), userID, familyID).Return(expectedResponse, nil)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/families/me/diaries/streak", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.GetStreak(c)
	if err != nil {
		t.Fatalf("GetStreak failed: %v", err)
	}

	// Verify response status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify response body
	var response struct {
		Data dto.StreakResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify streak data
	assert.Equal(t, expectedResponse, &response.Data)

	mockController.AssertExpectations(t)
}

// TestDiaryHandler_GetStreak_ControllerError tests error from controller
func TestDiaryHandler_GetStreak_ControllerError(t *testing.T) {
	t.Parallel()

	mockController := new(MockDiaryController)
	handler := NewDiaryHandler(mockController)

	familyID := uuid.New()
	userID := uuid.New()

	controllerErr := &errors.ValidationError{Message: "invalid user ID"}
	mockController.On("GetStreak", mock.Anything, userID, familyID).Return(nil, controllerErr)

	// Create HTTP request
	req := httptest.NewRequest("GET", "/families/me/diaries/streak", nil)
	ctx := context.WithValue(req.Context(), auth.ContextKeyFamilyID, familyID)
	ctx = context.WithValue(ctx, auth.ContextKeyUserID, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)

	// Call handler
	err := handler.GetStreak(c)
	if err != nil {
		t.Logf("expected error: %v", err)
	}

	mockController.AssertExpectations(t)
}
