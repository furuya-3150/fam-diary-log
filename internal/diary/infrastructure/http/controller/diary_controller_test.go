package controller

import (
	"context"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockDiaryUsecase struct {
	mock.Mock
}

func (m *MockDiaryUsecase) Create(ctx context.Context, diary *domain.Diary) (*domain.Diary, error) {
	args := m.Called(ctx, diary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Diary), args.Error(1)
}

func (m *MockDiaryUsecase) List(ctx context.Context, familyID uuid.UUID) ([]*domain.Diary, error) {
	args := m.Called(ctx, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Diary), args.Error(1)
}

func (m *MockDiaryUsecase) GetCount(ctx context.Context, familyID uuid.UUID, year, month string) (int, error) {
	args := m.Called(ctx, familyID, year, month)
	return args.Int(0), args.Error(1)
}

func (m *MockDiaryUsecase) GetStreak(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error) {
	args := m.Called(ctx, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Streak), args.Error(1)
}

// diary created successfully
func TestDiaryController_Create_Success(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	diaryID := uuid.New()
	familyID := uuid.New()
	userID := uuid.New()

	inputDiary := &domain.Diary{
		FamilyID: familyID,
		UserID:   userID,
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	expectedDiary := &domain.Diary{
		ID:        diaryID,
		FamilyID:  familyID,
		UserID:    userID,
		Title:     "Test Diary",
		Content:   "This is a test diary",
		CreatedAt: time.Now(),
	}

	mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(d *domain.Diary) bool {
		return d.Title == inputDiary.Title && d.Content == inputDiary.Content
	})).Return(expectedDiary, nil)

	// Call controller
	result, err := controller.Create(context.Background(), inputDiary)

	// Verify result
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.ID != diaryID {
		t.Errorf("expected ID %v, got %v", diaryID, result.ID)
	}

	if result.Title != inputDiary.Title {
		t.Errorf("expected title %q, got %q", inputDiary.Title, result.Title)
	}

	if result.Content != inputDiary.Content {
		t.Errorf("expected content %q, got %q", inputDiary.Content, result.Content)
	}

	mockUsecase.AssertExpectations(t)
}

// diary creation with validation error
func TestDiaryController_Create_ValidationError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	inputDiary := &domain.Diary{
		Title:   "",
		Content: "Valid content",
	}

	validationErr := &errors.ValidationError{Message: "title cannot be empty"}
	mockUsecase.On("Create", mock.Anything, mock.Anything).Return(nil, validationErr)

	// Call controller
	result, err := controller.Create(context.Background(), inputDiary)

	// Verify result
	if err == nil {
		t.Fatal("expected validation error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}

// diary creation with internal error
func TestDiaryController_Create_InternalError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	inputDiary := &domain.Diary{
		Title:   "Test Diary",
		Content: "Valid content",
	}

	internalErr := &errors.InternalError{Message: "database error"}
	mockUsecase.On("Create", mock.Anything, mock.Anything).Return(nil, internalErr)

	// Call controller
	result, err := controller.Create(context.Background(), inputDiary)

	// Verify result
	if err == nil {
		t.Fatal("expected internal error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}

// dto conversion test
func TestDiaryController_Create_DTOConversion(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	diaryID := uuid.New()
	familyID := uuid.New()
	userID := uuid.New()
	createdAt := time.Now()

	inputDiary := &domain.Diary{
		ID:       diaryID,
		FamilyID: familyID,
		UserID:   userID,
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	usecaseDiary := &domain.Diary{
		ID:        diaryID,
		FamilyID:  familyID,
		UserID:    userID,
		Title:     "Test Diary",
		Content:   "This is a test diary",
		CreatedAt: createdAt,
	}

	mockUsecase.On("Create", mock.Anything, mock.Anything).Return(usecaseDiary, nil)

	// Call controller
	result, err := controller.Create(context.Background(), inputDiary)

	// Verify result
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify DTO conversion
	expectedDTO := &dto.DiaryResponse{
		ID:        diaryID,
		FamilyID:  familyID,
		UserID:    userID,
		Title:     "Test Diary",
		Content:   "This is a test diary",
		CreatedAt: createdAt,
	}

	if result.ID != expectedDTO.ID {
		t.Errorf("expected ID %v, got %v", expectedDTO.ID, result.ID)
	}

	if result.FamilyID != expectedDTO.FamilyID {
		t.Errorf("expected FamilyID %v, got %v", expectedDTO.FamilyID, result.FamilyID)
	}

	if result.UserID != expectedDTO.UserID {
		t.Errorf("expected UserID %v, got %v", expectedDTO.UserID, result.UserID)
	}

	if result.Title != expectedDTO.Title {
		t.Errorf("expected Title %q, got %q", expectedDTO.Title, result.Title)
	}

	if result.Content != expectedDTO.Content {
		t.Errorf("expected Content %q, got %q", expectedDTO.Content, result.Content)
	}

	if result.CreatedAt != expectedDTO.CreatedAt {
		t.Errorf("expected CreatedAt %v, got %v", expectedDTO.CreatedAt, result.CreatedAt)
	}

	mockUsecase.AssertExpectations(t)
}

// diary creation with cancelled context
func TestDiaryController_Create_ContextCancelled(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	inputDiary := &domain.Diary{
		Title:   "Test Diary",
		Content: "Valid content",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ctxErr := &errors.InternalError{Message: "context cancelled"}
	mockUsecase.On("Create", mock.MatchedBy(func(c context.Context) bool {
		return c.Err() != nil
	}), mock.Anything).Return(nil, ctxErr)

	// Call controller
	result, err := controller.Create(ctx, inputDiary)

	// Verify result
	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	t.Logf("error: %v", err)

	mockUsecase.AssertExpectations(t)
}

// diary creation with multiple calls independent
func TestDiaryController_Create_MultipleCallsIndependent(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	diary1ID := uuid.New()
	diary2ID := uuid.New()

	diary1 := &domain.Diary{
		ID:        diary1ID,
		Title:     "First Diary",
		Content:   "First content",
		CreatedAt: time.Now(),
	}

	diary2 := &domain.Diary{
		ID:        diary2ID,
		Title:     "Second Diary",
		Content:   "Second content",
		CreatedAt: time.Now(),
	}

	// Setup expectations for both calls
	mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(d *domain.Diary) bool {
		return d.Title == "First Diary"
	})).Return(diary1, nil).Once()

	mockUsecase.On("Create", mock.Anything, mock.MatchedBy(func(d *domain.Diary) bool {
		return d.Title == "Second Diary"
	})).Return(diary2, nil).Once()

	// Call controller twice
	result1, err1 := controller.Create(context.Background(), &domain.Diary{Title: "First Diary", Content: "First content"})
	result2, err2 := controller.Create(context.Background(), &domain.Diary{Title: "Second Diary", Content: "Second content"})

	// Verify results
	if err1 != nil {
		t.Fatalf("First Create failed: %v", err1)
	}

	if err2 != nil {
		t.Fatalf("Second Create failed: %v", err2)
	}

	if result1.ID != diary1ID {
		t.Errorf("expected first ID %v, got %v", diary1ID, result1.ID)
	}

	if result2.ID != diary2ID {
		t.Errorf("expected second ID %v, got %v", diary2ID, result2.ID)
	}

	mockUsecase.AssertExpectations(t)
}

// ------------
// List Diaries
// ------------

// list diaries successfully
func TestDiaryController_List_Success(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()
	userID := uuid.New()
	diaryID1 := uuid.New()
	diaryID2 := uuid.New()
	createdAt := time.Now()

	expectedDiaries := []*domain.Diary{
		{
			ID:        diaryID1,
			FamilyID:  familyID,
			UserID:    userID,
			Title:     "Test Diary 1",
			Content:   "Content 1",
			CreatedAt: createdAt,
		},
		{
			ID:        diaryID2,
			FamilyID:  familyID,
			UserID:    userID,
			Title:     "Test Diary 2",
			Content:   "Content 2",
			CreatedAt: createdAt.Add(-24 * time.Hour),
		},
	}

	mockUsecase.On("List", mock.Anything, familyID).Return(expectedDiaries, nil)

	// Call controller
	result, err := controller.List(context.Background(), familyID)

	// Verify result
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result) != len(expectedDiaries) {
		t.Errorf("expected %d diaries, got %d", len(expectedDiaries), len(result))
	}

	if result[0].ID != diaryID1 {
		t.Errorf("expected first ID %v, got %v", diaryID1, result[0].ID)
	}

	if result[1].ID != diaryID2 {
		t.Errorf("expected second ID %v, got %v", diaryID2, result[1].ID)
	}

	mockUsecase.AssertExpectations(t)
}

// list diaries with validation error
func TestDiaryController_List_ValidationError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()

	validationErr := &errors.ValidationError{Message: "invalid date format"}
	mockUsecase.On("List", mock.Anything, familyID).Return(nil, validationErr)

	// Call controller
	result, err := controller.List(context.Background(), familyID)

	// Verify result
	if err == nil {
		t.Fatal("expected validation error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}

// list diaries with internal error
func TestDiaryController_List_InternalError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()

	internalErr := &errors.InternalError{Message: "database error"}
	mockUsecase.On("List", mock.Anything, familyID).Return(nil, internalErr)

	// Call controller
	result, err := controller.List(context.Background(), familyID)

	// Verify result
	if err == nil {
		t.Fatal("expected internal error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetCount_Success tests successful count retrieval
func TestDiaryController_GetCount_Success(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()

	mockUsecase.On("GetCount", mock.Anything, familyID, "2026", "01").Return(5, nil)

	// Call controller
	count, err := controller.GetCount(context.Background(), familyID, "2026", "01")

	// Verify result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetCount_ZeroCount tests count when no diaries exist
func TestDiaryController_GetCount_ZeroCount(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()

	mockUsecase.On("GetCount", mock.Anything, familyID, "2026", "02").Return(0, nil)

	// Call controller
	count, err := controller.GetCount(context.Background(), familyID, "2026", "02")

	// Verify result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetCount_UsecaseError tests error handling
func TestDiaryController_GetCount_UsecaseError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	familyID := uuid.New()
	usecaseErr := &errors.ValidationError{Message: "invalid date format"}

	mockUsecase.On("GetCount", mock.Anything, familyID, "2026", "13").Return(0, usecaseErr)

	// Call controller
	count, err := controller.GetCount(context.Background(), familyID, "2026", "13")

	// Verify result
	if err == nil {
		t.Fatal("expected error")
	}

	if count != 0 {
		t.Errorf("expected count 0, got %d", count)
	}

	mockUsecase.AssertExpectations(t)
}

// ============================================
// GetStreak Tests
// ============================================

// TestDiaryController_GetStreak_Success tests successful streak retrieval
func TestDiaryController_GetStreak_Success(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	userID := uuid.New()
	familyID := uuid.New()
	lastPostDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	expectedStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 5,
		LastPostDate:  &lastPostDate,
	}

	mockUsecase.On("GetStreak", mock.Anything, userID, familyID).Return(expectedStreak, nil)

	// Call controller
	result, err := controller.GetStreak(context.Background(), userID, familyID)

	// Verify result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, result.UserID)
	}

	if result.FamilyID != familyID {
		t.Errorf("expected FamilyID %v, got %v", familyID, result.FamilyID)
	}

	if result.CurrentStreak != 5 {
		t.Errorf("expected CurrentStreak 5, got %d", result.CurrentStreak)
	}

	if result.LastPostDate == nil {
		t.Fatal("LastPostDate is nil")
	}

	if !result.LastPostDate.Equal(lastPostDate) {
		t.Errorf("expected LastPostDate %v, got %v", lastPostDate, result.LastPostDate)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetStreak_NotFound tests when streak doesn't exist
func TestDiaryController_GetStreak_NotFound(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	userID := uuid.New()
	familyID := uuid.New()

	mockUsecase.On("GetStreak", mock.Anything, userID, familyID).Return(nil, nil)

	// Call controller
	result, err := controller.GetStreak(context.Background(), userID, familyID)

	// Verify result
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil for non-existent streak")
	}

	// Should return default values
	if result.CurrentStreak != 0 {
		t.Errorf("expected CurrentStreak 0, got %d", result.CurrentStreak)
	}

	if result.LastPostDate != nil {
		t.Errorf("expected LastPostDate nil, got %v", result.LastPostDate)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetStreak_ValidationError tests validation error handling
func TestDiaryController_GetStreak_ValidationError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	userID := uuid.New()
	familyID := uuid.New()

	validationErr := &errors.ValidationError{Message: "invalid user ID"}
	mockUsecase.On("GetStreak", mock.Anything, userID, familyID).Return(nil, validationErr)

	// Call controller
	result, err := controller.GetStreak(context.Background(), userID, familyID)

	// Verify result
	if err == nil {
		t.Fatal("expected error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}

// TestDiaryController_GetStreak_InternalError tests internal error handling
func TestDiaryController_GetStreak_InternalError(t *testing.T) {
	t.Parallel()

	mockUsecase := new(MockDiaryUsecase)
	controller := NewDiaryController(mockUsecase)

	userID := uuid.New()
	familyID := uuid.New()

	internalErr := &errors.InternalError{Message: "database error"}
	mockUsecase.On("GetStreak", mock.Anything, userID, familyID).Return(nil, internalErr)

	// Call controller
	result, err := controller.GetStreak(context.Background(), userID, familyID)

	// Verify result
	if err == nil {
		t.Fatal("expected error")
	}

	if result != nil {
		t.Errorf("expected nil result on error, got %v", result)
	}

	mockUsecase.AssertExpectations(t)
}
