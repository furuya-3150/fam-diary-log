package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
)

type MockDiaryAnalysisRepository struct {
	mock.Mock
}

func (m *MockDiaryAnalysisRepository) Create(ctx context.Context, analysis *domain.DiaryAnalysis) (*domain.DiaryAnalysis, error) {
	args := m.Called(ctx, analysis)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DiaryAnalysis), args.Error(1)
}

type MockNLPGateway struct {
	mock.Mock
}

func (m *MockNLPGateway) CheckAccuracy(ctx context.Context, content string) (int, error) {
	args := m.Called(ctx, content)
	return args.Get(0).(int), args.Error(1)
}

// TestDiaryAnalysisUsecaseAnalyzeSuccess tests successful analysis with NLP gateway
func TestDiaryAnalysisUsecaseAnalyzeSuccess(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	content := "これは日記のテスト内容です。"

	event := &domain.DiaryCreatedEvent{
		DiaryID:  diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Content:  content,
	}

	mockGateway.On("CheckAccuracy", mock.Anything, mock.MatchedBy(func(content string) bool {
		return content == event.Content
	})).Return(2, nil)

	expectedAnalysis := &domain.DiaryAnalysis{
		DiaryID:       diaryID,
		UserID:        userID,
		FamilyID:      familyID,
		AccuracyScore: 80,
	}

	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(analysis *domain.DiaryAnalysis) bool {
		return analysis.DiaryID == diaryID &&
			analysis.UserID == userID &&
			analysis.FamilyID == familyID &&
			analysis.AccuracyScore == 80
	})).Return(expectedAnalysis, nil)

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, diaryID, result.DiaryID)
	assert.Equal(t, 80, result.AccuracyScore)
	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

// TestDiaryAnalysisUsecaseAnalyzeEmptyContent tests empty content validation
func TestDiaryAnalysisUsecaseAnalyzeEmptyContent(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "",
	}

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &errors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "Create")
	mockGateway.AssertNotCalled(t, "CheckAccuracy")
}

// TestDiaryAnalysisUsecaseAnalyzeNLPGatewayNotConfigured tests NLP gateway not configured
func TestDiaryAnalysisUsecaseAnalyzeNLPGatewayNotConfigured(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}

	// Create usecase without NLPGateway (nil)
	usecase := NewDiaryAnalysisUsecase(mockRepo)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &errors.LogicError{}, err)
	mockRepo.AssertNotCalled(t, "Create")
}

// TestDiaryAnalysisUsecaseAnalyzeNLPGatewayError tests NLP gateway error with default score
func TestDiaryAnalysisUsecaseAnalyzeNLPGatewayError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	content := "test content"

	event := &domain.DiaryCreatedEvent{
		DiaryID:  diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Content:  content,
	}

	mockGateway.On("CheckAccuracy", mock.Anything, mock.MatchedBy(func(content string) bool {
		return content == event.Content
	})).Return(0, assert.AnError)
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(analysis *domain.DiaryAnalysis) bool {
		return analysis.DiaryID == diaryID &&
			analysis.AccuracyScore == CheckAccuracyDefaultScore
	})).Return(&domain.DiaryAnalysis{}, nil)

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, CheckAccuracyDefaultScore, result.AccuracyScore)
	mockRepo.AssertExpectations(t)
	mockGateway.AssertExpectations(t)
}

// TestDiaryAnalysisUsecaseAnalyzeRepositoryError tests repository error handling
func TestDiaryAnalysisUsecaseAnalyzeRepositoryError(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}

	mockGateway.On("CheckAccuracy", mock.Anything, mock.Anything).Return(85, nil)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, assert.AnError)

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRepo.AssertExpectations(t)
}

// TestDiaryAnalysisUsecaseAnalyzeContextCancelled tests cancelled context
func TestDiaryAnalysisUsecaseAnalyzeContextCancelled(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockGateway.On("CheckAccuracy", mock.Anything, mock.Anything).Return(0, context.Canceled)
	mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil, context.Canceled)

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(ctx, event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDiaryAnalysisUsecaseCountSentences tests sentence counting
func TestDiaryAnalysisUsecaseCountSentences(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway).(*diaryAnalysisUsecase)

	tests := []struct {
		name          string
		content       string
		expectedCount int
	}{
		{
			name:          "single sentence",
			content:       "これは一つの文です。",
			expectedCount: 1,
		},
		{
			name:          "multiple sentences",
			content:       "最初の文。次の文。三番目の文。",
			expectedCount: 3,
		},
		{
			name:          "with exclamation",
			content:       "すごい！素晴らしい？",
			expectedCount: 2,
		},
		{
			name:          "no sentences",
			content:       "句点がない",
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			count := usecase.countSentences(tt.content)

			// Assert
			assert.Equal(t, tt.expectedCount, count)
		})
	}
}

// TestDiaryAnalysisUsecaseAnalyzeNilDiaryID tests diary_id validation
func TestDiaryAnalysisUsecaseAnalyzeNilDiaryID(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.Nil,
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "これは日記のテスト内容です。",
	}

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &errors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "Create")
	mockGateway.AssertNotCalled(t, "CheckAccuracy")
}

// TestDiaryAnalysisUsecaseAnalyzeNilUserID tests user_id validation
func TestDiaryAnalysisUsecaseAnalyzeNilUserID(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.Nil,
		FamilyID: uuid.New(),
		Content:  "これは日記のテスト内容です。",
	}

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &errors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "Create")
	mockGateway.AssertNotCalled(t, "CheckAccuracy")
}

// TestDiaryAnalysisUsecaseAnalyzeNilFamilyID tests family_id validation
func TestDiaryAnalysisUsecaseAnalyzeNilFamilyID(t *testing.T) {
	// Arrange
	mockRepo := new(MockDiaryAnalysisRepository)
	mockGateway := new(MockNLPGateway)

	event := &domain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.Nil,
		Content:  "これは日記のテスト内容です。",
	}

	usecase := NewDiaryAnalysisUsecaseWithNLPGateway(mockRepo, mockGateway)

	// Act
	result, err := usecase.Analyze(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.IsType(t, &errors.ValidationError{}, err)
	mockRepo.AssertNotCalled(t, "Create")
	mockGateway.AssertNotCalled(t, "CheckAccuracy")
}
