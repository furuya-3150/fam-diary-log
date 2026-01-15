package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
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

type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) Begin(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransactionManager) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransactionManager) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTransactionManager) DiaryRepository() repository.DiaryRepository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(repository.DiaryRepository)
}

func (m *MockTransactionManager) ExecuteTransaction(ctx context.Context, fn func(context.Context) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
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

			usecase := NewDiaryUsecase(mockTm, mockRepo)

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

	diary := newValidDiary()
	expectedErr := &pkgerrors.InternalError{Message: "database connection failed"}

	mockRepo.On("Create", mock.Anything, diary).Return(nil, expectedErr)
	usecase := NewDiaryUsecase(mockTm, mockRepo)

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
	usecase := NewDiaryUsecase(mockTm, mockRepo)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Title, result.Title)
	assert.Equal(t, expected.Content, result.Content)
	mockRepo.AssertExpectations(t)
}

// diary creation with repository error
func TestDiaryUsecase_Create_RepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryRepository)
	mockTm := new(MockTransactionManager)

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	mockRepo.On("Create", mock.Anything, input).Return(nil, &pkgerrors.InternalError{Message: "database connection failed"})

	usecase := NewDiaryUsecase(mockTm, mockRepo)

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

	input := &domain.Diary{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Title:    "Test Diary",
		Content:  "This is a test diary",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockRepo.On("Create", ctx, input).Return(nil, context.Canceled)

	usecase := NewDiaryUsecase(mockTm, mockRepo)

	// Act
	result, err := usecase.Create(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

func TestDiaryUsecaseCreateValidateCreateDiaryRequest(t *testing.T) {
	tests := []struct {
		name      string
		diary     *domain.Diary
		wantError bool
	}{
		{
			name: "valid diary with single character title",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "a",
				Content:  "valid content",
			},
			wantError: false,
		},
		{
			name: "valid diary with max title length",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    string(make([]byte, 255)),
				Content:  "valid content",
			},
			wantError: false,
		},
		{
			name: "invalid - whitespace only title",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "   ",
				Content:  "valid content",
			},
			wantError: true,
		},
		{
			name: "invalid - whitespace only content",
			diary: &domain.Diary{
				ID:       uuid.New(),
				UserID:   uuid.New(),
				FamilyID: uuid.New(),
				Title:    "valid title",
				Content:  "   ",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockDiaryRepository)
			mockTm := new(MockTransactionManager)

			usecase := NewDiaryUsecase(mockTm, mockRepo)

			if !tt.wantError {
				mockRepo.On("Create", mock.Anything, tt.diary).Return(tt.diary, nil)
			}

			_, err := usecase.Create(context.Background(), tt.diary)

			if tt.wantError {
				assert.Error(t, err)
				mockRepo.AssertNotCalled(t, "Create")
			} else {
				assert.NoError(t, err)
				mockRepo.AssertExpectations(t)
			}
		})
	}
}

// ------------
// List Diaries
// ------------

// TestDiaryUsecaseListSuccess tests successful diary list retrieval
func TestDiaryUsecaseListSuccess(t *testing.T) {
	t.Parallel()

	mockRepo := new(MockDiaryRepository)
	mockTxManager := new(MockTransactionManager)

	// テスト用の固定時刻を指定（2026-01-15, Thursday）
	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	clk := &clock.Fixed{Time: fixedTime}

	// Clock を注入
	usecase := NewDiaryUsecaseWithClock(mockTxManager, mockRepo, clk)

	familyID := uuid.New()
	diaryID1 := uuid.New()
	diaryID2 := uuid.New()
	createdAt := time.Now()

	expectedDiaries := []*domain.Diary{
		{
			ID:        diaryID1,
			FamilyID:  familyID,
			Title:     "Test Diary 1",
			Content:   "Content 1",
			CreatedAt: createdAt,
		},
		{
			ID:        diaryID2,
			FamilyID:  familyID,
			Title:     "Test Diary 2",
			Content:   "Content 2",
			CreatedAt: createdAt.Add(-24 * time.Hour),
		},
	}

	// 2026-01-15は木曜日なので、週は2026-01-12（月）から2026-01-18（日）
	expectedStartDate := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	expectedEndDate := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)
	mockRepo.On("List", mock.Anything, mock.MatchedBy(func(criteria *domain.DiarySearchCriteria) bool {
		return criteria.FamilyID == familyID &&
			criteria.StartDate.Equal(expectedStartDate) &&
			criteria.EndDate.Equal(expectedEndDate)
	}), mock.Anything).Return(expectedDiaries, nil)

	// Call usecase
	result, err := usecase.List(context.Background(), familyID)

	// Verify result
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, len(expectedDiaries), len(result))
	assert.Equal(t, diaryID1, result[0].ID)
	assert.Equal(t, diaryID2, result[1].ID)

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
	usecase := NewDiaryUsecaseWithClock(mockTxManager, mockRepo, clk)

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
