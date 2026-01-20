package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	infctx "github.com/furuya-3150/fam-diary-log/internal/diary-analysis/infrastructure/context"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDiaryAnalysisUsecase struct {
	mock.Mock
}

func (m *MockDiaryAnalysisUsecase) GetWeekCharCount(ctx context.Context, userID uuid.UUID, dateStr string) (int, error) {
	args := m.Called(ctx, userID, dateStr)
	return args.Int(0), args.Error(1)
}

func (m *MockDiaryAnalysisUsecase) GetCharCountByDate(ctx context.Context, userID uuid.UUID, dateStr string) (map[string]interface{}, error) {
	args := m.Called(ctx, userID, dateStr)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// GetWeekCharCount with valid user ID and date - success
func TestDiaryAnalysisHandler_GetWeekCharCount_Success(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryAnalysisUsecase)
	handler := NewDiaryAnalysisHandler(mockUsecase)

	userID := uuid.New()
	date := "2026-01-20"
	expectedCount := map[string]interface{}{
		"2026-01-20": 100,
		"2026-01-21": 150,
	}

	mockUsecase.On("GetCharCountByDate", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx.Value(infctx.UserIDKey) == userID
	}), mock.MatchedBy(func(id uuid.UUID) bool {
		return id == userID
	}), mock.MatchedBy(func(argDate string) bool {
		return argDate == date
	})).Return(expectedCount, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/week-char-count/"+date, nil)

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("date")
	c.SetParamValues(date)

	// Call handler
	err := handler.GetWeekCharCount(c)
	if err != nil {
		t.Fatalf("GetWeekCharCount failed: %v", err)
	}

	// Verify response status code
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Verify response body
	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err, "failed to unmarshal response")
	assert.NotNil(t, response["data"], "expected data in response")
}

// GetWeekCharCount with missing user ID
func TestDiaryAnalysisHandler_GetWeekCharCount_MissingUserID(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryAnalysisUsecase)
	handler := NewDiaryAnalysisHandler(mockUsecase)

	date := "2026-01-20"

	// Create request without userID in context
	req := httptest.NewRequest(http.MethodGet, "/week-char-count/"+date, nil)
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("date")
	c.SetParamValues(date)

	// Call handler
	err := handler.GetWeekCharCount(c)
	// Handler returns nil, Echoが処理している

	// Verify response status code
	assert.NotEqual(t, http.StatusOK, rec.Code, "expected error status code")

	// Verify response body contains error message
	var errorResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.NoError(t, err, "failed to unmarshal error response")
	assert.Equal(t, "userIDを指定してください", errorResponse["message"])
	assert.Equal(t, "LOGIC_ERROR", errorResponse["code"])
	assert.NotNil(t, errorResponse["message"], "expected message ")
	assert.NotNil(t, errorResponse["code"], "expected code in response")
}

// GetWeekCharCount with invalid date format
func TestDiaryAnalysisHandler_GetWeekCharCount_InvalidDate(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryAnalysisUsecase)
	handler := NewDiaryAnalysisHandler(mockUsecase)

	userID := uuid.New()
	invalidDate := "invalid-date"

	mockUsecase.On("GetCharCountByDate", mock.Anything, userID, invalidDate).Return(map[string]interface{}{}, &errors.ValidationError{Message: "invalid date format"})

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/week-char-count/"+invalidDate, nil)

	// Set context values
	ctx := context.WithValue(req.Context(), infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	// Create response writer
	rec := httptest.NewRecorder()

	// Create Echo context
	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetParamNames("date")
	c.SetParamValues(invalidDate)

	// Call handler
	err := handler.GetWeekCharCount(c)
	// Handler returns nil, Echoが処理している

	// Verify response status code
	assert.NotEqual(t, http.StatusOK, rec.Code, "expected error status code")

	// Verify response body contains error message
	var errorResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Equal(t, nil, err)
	assert.Equal(t, "VALIDATION_ERROR", errorResponse["code"], "expected error in response")
	assert.Equal(t, "VALIDATION_ERROR", errorResponse["code"], "expected error in response")

	// Verify mock was called
	mockUsecase.AssertCalled(t, "GetCharCountByDate", mock.Anything, userID, invalidDate)
}
