package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	infctx "github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/context"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// TODO: 前テストのctxの値をチェック

// create diary successfully
func TestE2E_CreateDiary_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test")
	}
	t.Parallel()

	godotenv.Load("../../../../cmd/diary-api/.env")
	e := NewRouter()

	familyID := uuid.New()
	userID := uuid.New()

	createRequest := map[string]string{
		"title":   "My First Diary",
		"content": "This is my first diary entry",
	}

	body, _ := json.Marshal(createRequest)
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add auth context
	ctx := req.Context()
	ctx = context.WithValue(ctx, infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call the Echo server
	e.ServeHTTP(rec, req)

	// Verify response status
	if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
		t.Errorf("expected status %d or %d, got %d", http.StatusOK, http.StatusCreated, rec.Code)
		t.Logf("response body: %s", rec.Body.String())
		return
	}

	// Verify response body structure
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if _, ok := response["data"]; !ok {
		t.Error("response should contain 'data' field")
	}
}

// create diary with validation error
func TestE2E_CreateDiary_ValidationError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test")
	}

	t.Parallel()

	godotenv.Load("../../../../cmd/diary-api/.env")
	e := NewRouter()

	familyID := uuid.New()
	userID := uuid.New()

	// Empty title should fail validation
	createRequest := map[string]string{
		"title":   "",
		"content": "This is a diary with empty title",
	}

	body, _ := json.Marshal(createRequest)
	req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := req.Context()
	ctx = context.WithValue(ctx, infctx.FamilyIDKey, familyID)
	ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	// Should return error status (400 or 422)
	if rec.Code != http.StatusBadRequest && rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d or %d, got %d", http.StatusBadRequest, http.StatusUnprocessableEntity, rec.Code)
		t.Logf("response: %s", rec.Body.String())
	}
}

// create multiple diaries successfully
func TestE2E_CreateDiary_MultipleDiaries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test")
	}

	t.Parallel()

	godotenv.Load("../../../../cmd/diary-api/.env")
	e := NewRouter()

	familyID := uuid.New()
	userID := uuid.New()

	diaries := []map[string]string{
		{
			"title":   "First Diary",
			"content": "First entry content",
		},
		{
			"title":   "Second Diary",
			"content": "Second entry content",
		},
		{
			"title":   "Third Diary",
			"content": "Third entry content",
		},
	}

	for _, diary := range diaries {
		body, _ := json.Marshal(diary)
		req := httptest.NewRequest(http.MethodPost, "/diaries", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		ctx := req.Context()
		ctx = context.WithValue(ctx, infctx.FamilyIDKey, familyID)
		ctx = context.WithValue(ctx, infctx.UserIDKey, userID)
		req = req.WithContext(ctx)

		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK && rec.Code != http.StatusCreated {
			t.Errorf("failed to create diary: status %d, body: %s", rec.Code, rec.Body.String())
		}
	}
}

// health check succee
func TestE2E_Healthz(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test")
	}
	
	t.Parallel()

	godotenv.Load("../../../../cmd/diary-api/.env")
	e := NewRouter()

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	expectedBody := "Hello, World!"
	if rec.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}