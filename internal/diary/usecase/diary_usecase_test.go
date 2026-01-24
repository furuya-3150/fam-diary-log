package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/events"
	"github.com/furuya-3150/fam-diary-log/pkg/pagination"
)

type MockDiaryRepository struct {
	mock.Mock
}

func (m *MockDiaryRepository) Create(ctx context.Context, diary *domain.Diary) (*domain.Diary, error) {
	args := m.Called(ctx, diary)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Diary), args.Error(1)
}

func (m *MockDiaryRepository) List(ctx context.Context, criteria *domain.DiarySearchCriteria, pag *pagination.Pagination) ([]*domain.Diary, error) {
	args := m.Called(ctx, criteria, pag)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Diary), args.Error(1)
}

func (m *MockDiaryRepository) GetCount(ctx context.Context, criteria *domain.DiaryCountCriteria) (int, error) {
	args := m.Called(ctx, criteria)
	return args.Int(0), args.Error(1)
}

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) BeginTx(ctx context.Context) (context.Context, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return ctx, args.Error(1)
	}
	return args.Get(0).(context.Context), args.Error(1)
}

func (m *MockTransactionManager) CommitTx(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransactionManager) RollbackTx(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransactionManager) WithTx(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

type MockPublisher struct {
	mock.Mock
}

func (m *MockPublisher) Publish(ctx context.Context, event events.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPublisher) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockStreakRepository struct {
	mock.Mock
}

func (m *MockStreakRepository) CreateOrUpdate(ctx context.Context, streak *domain.Streak) (*domain.Streak, error) {
	args := m.Called(ctx, streak)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Streak), args.Error(1)
}

func (m *MockStreakRepository) Get(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error) {
	args := m.Called(ctx, userID, familyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Streak), args.Error(1)
}

// Helper function to create a valid diary for testing
func newValidDiary() *domain.Diary {
	return &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}
}

// TestDiaryUsecaseCreateValidationError tests various validation errors
func TestDiaryUsecaseCreateValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		diary   *domain.Diary
		wantErr bool
	}{
		{
			name: "empty title",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "",
				Content:  "valid content",
			},
			wantErr: true,
		},
		{
			name: "title too long",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    string(make([]byte, 256)),
				Content:  "valid content",
			},
			wantErr: true,
		},
		{
			name: "empty content",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "valid title",
				Content:  "",
			},
			wantErr: true,
		},
		{
			name: "whitespace only title",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "   ",
				Content:  "valid content",
			},
			wantErr: true,
		},
		{
			name: "whitespace only content",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "valid title",
				Content:  "   ",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDiaryRepository)
			mockTm := new(MockTransactionManager)
			mockPub := new(MockPublisher)
			mockStreakRepo := new(MockStreakRepository)

			usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

			_, err := usecase.Create(context.Background(), tt.diary)

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}

			if tt.wantErr {
				if _, ok := err.(*pkgerrors.ValidationError); !ok {
					t.Errorf("expected ValidationError, got %T", err)
				}
			}

			mockRepo.AssertNotCalled(t, "Create")
		})
	}
}

// TestDiaryUsecaseCreateRepositoryError tests repository error handling
func TestDiaryUsecaseCreateRepositoryError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	diary := newValidDiary()
	expectedErr := &pkgerrors.InternalError{Message: "database connection failed"}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, diary).Return(nil, expectedErr)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	_, err := usecase.Create(context.Background(), diary)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if _, ok := err.(*pkgerrors.InternalError); !ok {
		t.Errorf("expected InternalError, got %T", err)
	}
}

// TestDiaryUsecaseCreateSuccess tests successful diary creation
func TestDiaryUsecase_Create_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, input).Return(expected, nil)
	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(event interface{}) bool {
		if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
			return diaryEvent.DiaryID == diaryID && diaryEvent.UserID == userID && diaryEvent.FamilyID == familyID
		}
		return false
	})).Return(nil)
	mockPub.On("Close").Return(nil)
	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(&domain.Streak{}, nil)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Title, result.Title)
	assert.Equal(t, expected.Content, result.Content)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

// diary creation with repository error
func TestDiaryUsecase_Create_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, input).Return(nil, &pkgerrors.InternalError{Message: "database connection failed"})
	mockTm.On("RollbackTx", mock.Anything).Return(nil)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &pkgerrors.InternalError{}, err)
	mockRepo.AssertExpectations(t)
}

// diary creation with cancelled context
func TestDiaryUsecaseCreateContextCancelled(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockTm.On("BeginTx", mock.Anything).Return(ctx, nil)
	mockRepo.On("Create", mock.Anything, input).Return(nil, context.Canceled)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestDiaryUsecaseListRepositoryError tests diary list with repository error
func TestDiaryUsecaseListRepositoryError(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockDiaryRepository)
	mockTxManager := new(MockTransactionManager)

	// テスト用の固定時刻を指定（2026-01-15, Thursday）
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	clk := &clock.Fixed{Time: fixedTime}

	// Clock を注入
	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTxManager, mockRepo, mockStreakRepo, nil, clk)

	familyID := uuid.New()

	repositoryErr := &pkgerrors.InternalError{Message: "database error"}
	expectedStartDate := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	expectedEndDate := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(criteria *domain.DiarySearchCriteria) bool {
		return criteria.FamilyID == familyID &&
			criteria.StartDate.Equal(expectedStartDate) &&
			criteria.EndDate.Equal(expectedEndDate)
	}), mock.Anything).Return(nil, repositoryErr)

	// Call usecase
	result, err := usecase.List(context.Background(), familyID)

	// Verify error
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repositoryErr, err)

	mockRepo.AssertExpectations(t)
}

// ============================================
// Event Publishing Tests
// ============================================

// TestDiaryUsecase_Create_PublishEvent tests that diary created event is published
func TestDiaryUsecase_Create_PublishEvent(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("Create", mock.Anything, input).Return(expected, nil)

	// Verify that Publish is called with correct event
	var capturedEvent *domain.DiaryCreatedEvent
	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(event interface{}) bool {
		if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
			capturedEvent = diaryEvent
			return diaryEvent.DiaryID == diaryID && diaryEvent.UserID == userID && diaryEvent.FamilyID == familyID && diaryEvent.Content == "This is a test diary content"
		}
		return false
	})).Return(nil)

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockPub.On("Close").Return(nil)

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(&domain.Streak{}, nil)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedEvent)
	assert.Equal(t, diaryID, capturedEvent.DiaryID)
	assert.Equal(t, userID, capturedEvent.UserID)
	assert.Equal(t, familyID, capturedEvent.FamilyID)
	assert.Equal(t, "This is a test diary content", capturedEvent.Content)
	mockPub.AssertExpectations(t)
}

// TestDiaryUsecase_Create_PublishEventError tests successful diary creation even if event publishing fails
func TestDiaryUsecase_Create_PublishEventError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("Create", mock.Anything, input).Return(expected, nil)

	// Simulate publisher error
	publishErr := &pkgerrors.InternalError{Message: "failed to publish event"}
	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(event interface{}) bool {
		if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
			return diaryEvent.DiaryID == diaryID
		}
		return false
	})).Return(publishErr)

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(&domain.Streak{}, nil)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	// Should still succeed even if publishing fails (error is only logged)
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

// TestDiaryUsecase_Create_NilPublisher tests create fails when publisher is nil
func TestDiaryUsecase_Create_NilPublisher(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	// Create usecase with nil publisher
	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	if logicErr, ok := err.(*pkgerrors.LogicError); ok {
		assert.Equal(t, "publisher is not set", logicErr.Message)
	} else {
		t.Errorf("expected LogicError, got %T", err)
	}
	mockRepo.AssertNotCalled(t, "Create")
}

// ============================================
// GetCount Tests
// ============================================

// TestDiaryUsecase_GetCount_Success tests successful count retrieval
func TestDiaryUsecase_GetCount_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	familyID := uuid.New()
	criteria := &domain.DiaryCountCriteria{
		FamilyID:  familyID,
		YearMonth: "2026-01",
	}

	mockRepo.On("GetCount", mock.Anything, criteria).Return(5, nil)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	count, err := usecase.GetCount(context.Background(), familyID, "2026", "01")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	mockRepo.AssertExpectations(t)
}

// TestDiaryUsecase_GetCount_InvalidMonth tests invalid month validation
func TestDiaryUsecase_GetCount_InvalidMonth(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	familyID := uuid.New()
	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	count, err := usecase.GetCount(context.Background(), familyID, "2026", "13")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.IsType(t, &pkgerrors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "GetCount")
}

// TestDiaryUsecase_GetCount_InvalidYear tests invalid year validation
func TestDiaryUsecase_GetCount_InvalidYear(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	familyID := uuid.New()
	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	count, err := usecase.GetCount(context.Background(), familyID, "0", "01")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.IsType(t, &pkgerrors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "GetCount")
}

// TestDiaryUsecase_GetCount_ZeroCount tests count when no diaries exist
func TestDiaryUsecase_GetCount_ZeroCount(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	familyID := uuid.New()
	criteria := &domain.DiaryCountCriteria{
		FamilyID:  familyID,
		YearMonth: "2026-02",
	}

	mockRepo.On("GetCount", mock.Anything, criteria).Return(0, nil)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	count, err := usecase.GetCount(context.Background(), familyID, "2026", "02")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	mockRepo.AssertExpectations(t)
}

// TestDiaryUsecase_GetCount_RepositoryError tests error handling
func TestDiaryUsecase_GetCount_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	familyID := uuid.New()
	criteria := &domain.DiaryCountCriteria{
		FamilyID:  familyID,
		YearMonth: "2026-01",
	}

	expectedErr := &pkgerrors.InternalError{Message: "database error"}
	mockRepo.On("GetCount", mock.Anything, criteria).Return(0, expectedErr)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, nil, &clock.Real{})

	// Act
	count, err := usecase.GetCount(context.Background(), familyID, "2026", "01")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, expectedErr, err)
	mockRepo.AssertExpectations(t)
}

// ============================================
// Transaction Tests
// ============================================

// TestDiaryUsecase_Create_CommitOnSuccess tests that commit is called on successful creation
func TestDiaryUsecase_Create_CommitOnSuccess(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(d *domain.Diary) bool {
		return d.UserID == userID && d.FamilyID == familyID
	})).Return(expected, nil)

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.MatchedBy(func(s *domain.Streak) bool {
		return s.UserID == userID && s.FamilyID == familyID
	})).Return(&domain.Streak{}, nil)

	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(event interface{}) bool {
		if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
			return diaryEvent.DiaryID == expected.ID
		}
		return false
	})).Return(nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockPub.On("Close").Return(nil)

	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockTm.AssertCalled(t, "BeginTx", mock.Anything)
	mockTm.AssertCalled(t, "CommitTx", mock.Anything)
	mockTm.AssertNotCalled(t, "RollbackTx", mock.Anything)
}

// TestDiaryUsecase_Create_RollbackOnDiaryCreateError tests that rollback is called when diary creation fails
func TestDiaryUsecase_Create_RollbackOnDiaryCreateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	expectedErr := &pkgerrors.InternalError{Message: "database error"}
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, expectedErr)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)

	mockStreakRepo := new(MockStreakRepository)
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockTm.AssertCalled(t, "BeginTx", mock.Anything)
	mockTm.AssertCalled(t, "RollbackTx", mock.Anything)
	mockTm.AssertNotCalled(t, "CommitTx", mock.Anything)
}

// TestDiaryUsecase_Create_RollbackOnStreakUpdateError tests that rollback is called when streak update fails
func TestDiaryUsecase_Create_RollbackOnStreakUpdateError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	streakErr := &pkgerrors.InternalError{Message: "streak update failed"}
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(nil, streakErr)

	mockTm.On("RollbackTx", mock.Anything).Return(nil)
	// mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockPub.On("Publish", mock.Anything, mock.MatchedBy(func(event interface{}) bool {
		if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
			return diaryEvent.DiaryID == expected.ID
		}
		return false
	})).Return(nil)
	mockPub.On("Close").Return(nil)

	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err) // Streak error is not returned, only logged
	assert.Nil(t, result)
	mockTm.AssertCalled(t, "BeginTx", mock.Anything)
	mockTm.AssertCalled(t, "RollbackTx", mock.Anything)
	mockTm.AssertNotCalled(t, "CommitTx", mock.Anything)
}

// TestDiaryUsecase_Create_RollbackOnPublishError tests that rollback is called when publish fails
func TestDiaryUsecase_Create_RollbackOnPublishError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Test Diary",
		Content:  "This is a test diary content",
	}

	expected := &domain.Diary{
		ID:        diaryID,
		UserID:    userID,
		FamilyID:  familyID,
		Title:     "Test Diary",
		Content:   "This is a test diary content",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(expected, nil)

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.Anything).Return(&domain.Streak{}, nil)

	publishErr := &pkgerrors.InternalError{Message: "publish failed"}
	mockPub.On("Publish", mock.Anything, mock.Anything).Return(publishErr)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)

	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, &clock.Real{})

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err) // Publish error is not returned, only logged
	assert.Nil(t, result)
	mockTm.AssertCalled(t, "BeginTx", mock.Anything)
	mockTm.AssertCalled(t, "RollbackTx", mock.Anything)
	mockTm.AssertNotCalled(t, "CommitTx", mock.Anything)
}

// TestDiaryUsecase_Create_StreakCalculationOnFirstEntry tests streak calculation for first diary entry
func TestDiaryUsecase_Create_StreakCalculationOnFirstEntry(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		Title:    "First Diary",
		Content:  "First diary content",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(input, nil)

	mockStreakRepo := new(MockStreakRepository)
	// First entry: no existing streak
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)

	// Should create streak with default value
	var capturedStreak *domain.Streak
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.MatchedBy(func(s *domain.Streak) bool {
		capturedStreak = s
		return s.UserID == userID && s.FamilyID == familyID
	})).Return(&domain.Streak{CurrentStreak: domain.DefaultStreakValue}, nil)

	mockPub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockPub.On("Close").Return(nil)

	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedStreak)
	assert.Equal(t, domain.DefaultStreakValue, capturedStreak.CurrentStreak)
	mockStreakRepo.AssertExpectations(t)
}

// TestDiaryUsecase_Create_StreakIncrementOnConsecutiveEntry tests streak increment for consecutive entries
func TestDiaryUsecase_Create_StreakIncrementOnConsecutiveEntry(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Second Diary",
		Content:  "Second diary content",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(input, nil)
	// mockRepo.On("Close").Return(nil)

	// Fixed time: 2026-01-15 (Thursday)
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	// Previous entry: yesterday 2026-01-14
	previousPostDate := time.Date(2026, 1, 14, 9, 0, 0, 0, time.UTC)
	existingStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 1,
		LastPostDate:  &previousPostDate,
	}

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(existingStreak, nil)

	// Should increment streak to 2
	var capturedStreak *domain.Streak
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.MatchedBy(func(s *domain.Streak) bool {
		capturedStreak = s
		return s.UserID == userID && s.FamilyID == familyID
	})).Return(&domain.Streak{CurrentStreak: 2}, nil)

	mockPub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockPub.On("Close").Return(nil)

	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedStreak)
	assert.Equal(t, 2, capturedStreak.CurrentStreak)
	mockStreakRepo.AssertExpectations(t)
}

// TestDiaryUsecase_Create_StreakResetOnNonConsecutiveEntry tests streak resets to default on non-consecutive entry
func TestDiaryUsecase_Create_StreakResetOnNonConsecutiveEntry(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		Title:    "After Gap",
		Content:  "Diary after gap",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockTm.On("CommitTx", mock.Anything).Return(nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(input, nil)

	// Fixed time: 2026-01-15 (Thursday)
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	// Previous entry: 3 days ago 2026-01-12
	previousPostDate := time.Date(2026, 1, 12, 9, 0, 0, 0, time.UTC)
	existingStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 5,
		LastPostDate:  &previousPostDate,
	}

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(existingStreak, nil)

	// Should reset streak to default
	var capturedStreak *domain.Streak
	mockStreakRepo.On("CreateOrUpdate", mock.Anything, mock.MatchedBy(func(s *domain.Streak) bool {
		capturedStreak = s
		return s.UserID == userID && s.FamilyID == familyID
	})).Return(&domain.Streak{CurrentStreak: domain.DefaultStreakValue}, nil)

	mockPub.On("Publish", mock.Anything, mock.Anything).Return(nil)
	mockPub.On("Close").Return(nil)

	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, capturedStreak)
	assert.Equal(t, domain.DefaultStreakValue, capturedStreak.CurrentStreak)
	mockStreakRepo.AssertExpectations(t)
}

// TestDiaryUsecase_Create_DuplicatePostError tests error when posting duplicate diary on same day
func TestDiaryUsecase_Create_DuplicatePostError(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)

	userID := uuid.New()
	familyID := uuid.New()

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Duplicate Post",
		Content:  "Duplicate diary content",
	}

	mockTm.On("BeginTx", mock.Anything).Return(context.Background(), nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(input, nil)
	mockTm.On("RollbackTx", mock.Anything).Return(nil)

	// Fixed time: 2026-01-15 (Thursday)
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)

	// Previous entry: same day 2026-01-15
	previousPostDate := time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)
	existingStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 3,
		LastPostDate:  &previousPostDate,
	}

	mockStreakRepo := new(MockStreakRepository)
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(existingStreak, nil)

	mockPub.On("Publish", mock.Anything, mock.Anything).Return(nil)

	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.Error(t, err) // Error is logged but not returned
	assert.Nil(t, result)
	mockTm.AssertCalled(t, "RollbackTx", mock.Anything)
}

// ============================================
// GetStreak Tests
// ============================================

// TestDiaryUsecase_GetStreak_Success tests successful streak retrieval
func TestDiaryUsecase_GetStreak_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	userID := uuid.New()
	familyID := uuid.New()
	lastPostDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	expectedStreak := &domain.Streak{
		UserID:        userID,
		FamilyID:      familyID,
		CurrentStreak: 5,
		LastPostDate:  &lastPostDate,
	}

	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(expectedStreak, nil)

	clk := &clock.Real{}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.GetStreak(context.Background(), userID, familyID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, familyID, result.FamilyID)
	assert.Equal(t, 5, result.CurrentStreak)
	assert.Equal(t, lastPostDate, *result.LastPostDate)

	mockStreakRepo.AssertExpectations(t)
}

// TestDiaryUsecase_GetStreak_NotFound tests when streak doesn't exist
func TestDiaryUsecase_GetStreak_NotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	userID := uuid.New()
	familyID := uuid.New()

	// Repository returns (nil, nil) when record not found
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, nil)

	clk := &clock.Real{}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.GetStreak(context.Background(), userID, familyID)

	// Assert
	assert.NoError(t, err)
	assert.Nil(t, result)

	mockStreakRepo.AssertExpectations(t)
}

// TestDiaryUsecase_GetStreak_InvalidUserID tests validation error for invalid user ID
func TestDiaryUsecase_GetStreak_InvalidUserID(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	familyID := uuid.New()

	clk := &clock.Real{}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.GetStreak(context.Background(), uuid.Nil, familyID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &pkgerrors.ValidationError{}, err)
	assert.Equal(t, "invalid user ID", err.(*pkgerrors.ValidationError).Message)
}

// TestDiaryUsecase_GetStreak_InvalidFamilyID tests validation error for invalid family ID
func TestDiaryUsecase_GetStreak_InvalidFamilyID(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	userID := uuid.New()

	clk := &clock.Real{}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.GetStreak(context.Background(), userID, uuid.Nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &pkgerrors.ValidationError{}, err)
	assert.Equal(t, "invalid family ID", err.(*pkgerrors.ValidationError).Message)
}

// TestDiaryUsecase_GetStreak_RepositoryError tests repository error handling
func TestDiaryUsecase_GetStreak_RepositoryError(t *testing.T) {
	t.Parallel()

	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)
	mockPub := new(MockPublisher)
	mockStreakRepo := new(MockStreakRepository)

	userID := uuid.New()
	familyID := uuid.New()

	repositoryErr := &pkgerrors.InternalError{Message: "database error"}
	mockStreakRepo.On("Get", mock.Anything, userID, familyID).Return(nil, repositoryErr)

	clk := &clock.Real{}
	usecase := NewDiaryUsecase(mockTm, mockRepo, mockStreakRepo, mockPub, clk)

	// Act
	result, err := usecase.GetStreak(context.Background(), userID, familyID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, repositoryErr, err)

	mockStreakRepo.AssertExpectations(t)
}
