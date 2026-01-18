package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	diarydomain "github.com/furuya-3150/fam-diary-log/internal/diary/domain"
)

type MockDiaryAnalysisUsecase struct {
	mock.Mock
}

func (m *MockDiaryAnalysisUsecase) Analyze(ctx context.Context, event *domain.DiaryCreatedEvent) (*domain.DiaryAnalysis, error) {
	args := m.Called(ctx, event)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.DiaryAnalysis), args.Error(1)
}

// TestDiaryEventHandlerHandleSuccess tests successful event handling
func TestDiaryEventHandlerHandleSuccess(t *testing.T) {
	// Arrange
	mockUsecase := new(MockDiaryAnalysisUsecase)
	log := slog.Default()

	diaryID := uuid.New()
	userID := uuid.New()
	familyID := uuid.New()
	content := "test diary content"

	event := &diarydomain.DiaryCreatedEvent{
		DiaryID:  diaryID,
		UserID:   userID,
		FamilyID: familyID,
		Content:  content,
	}

	expectedAnalysis := &domain.DiaryAnalysis{
		ID:            uuid.New(),
		DiaryID:       diaryID,
		UserID:        userID,
		FamilyID:      familyID,
		AccuracyScore: 85,
	}

	mockUsecase.On("Analyze", mock.Anything, mock.MatchedBy(func (event *domain.DiaryCreatedEvent) bool {
		return event.DiaryID == diaryID && event.UserID == userID && event.FamilyID == familyID && event.Content == content
	})).Return(expectedAnalysis, nil)

	handler := NewDiaryEventHandler(mockUsecase, log)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	// Act
	err = handler.Handle(context.Background(), "diary.created", eventBytes)

	t.Log("error", err)

	// Assert
	assert.NoError(t, err)
	mockUsecase.AssertExpectations(t)
}

// TestDiaryEventHandlerHandleInvalidRoutingKey tests invalid routing key
func TestDiaryEventHandlerHandleInvalidRoutingKey(t *testing.T) {
	// Arrange
	mockUsecase := new(MockDiaryAnalysisUsecase)
	log := slog.Default()

	event := &diarydomain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}

	handler := NewDiaryEventHandler(mockUsecase, log)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	// Act
	err = handler.Handle(context.Background(), "unknown.event", eventBytes)

	// Assert
	assert.Error(t, err)
	mockUsecase.AssertNotCalled(t, "Analyze")
}

// TestDiaryEventHandlerHandleInvalidEventType tests invalid event type
func TestDiaryEventHandlerHandleInvalidEventType(t *testing.T) {
	// Arrange
	mockUsecase := new(MockDiaryAnalysisUsecase)
	log := slog.Default()

	handler := NewDiaryEventHandler(mockUsecase, log)

	// Act - passing wrong event type
	err := handler.Handle(context.Background(), "diary.created", []byte("invalid event"))

	// Assert
	assert.Error(t, err)
	mockUsecase.AssertNotCalled(t, "Analyze")
}

// TestDiaryEventHandlerHandleUsecaseError tests usecase error handling
func TestDiaryEventHandlerHandleUsecaseError(t *testing.T) {
	// Arrange
	mockUsecase := new(MockDiaryAnalysisUsecase)
	log := slog.Default()

	event := &diarydomain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}
	
	mockUsecase.On("Analyze", mock.Anything, mock.MatchedBy(func (e *domain.DiaryCreatedEvent) bool {
		return e.DiaryID == event.DiaryID && e.UserID == event.UserID && e.FamilyID == event.FamilyID && e.Content == event.Content
	})).Return(nil, assert.AnError)

	handler := NewDiaryEventHandler(mockUsecase, log)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	// Act
	err = handler.Handle(context.Background(), "diary.created", eventBytes)

	// Assert
	assert.Error(t, err)
	mockUsecase.AssertExpectations(t)
}

// TestDiaryEventHandlerHandleContextCancelled tests cancelled context
func TestDiaryEventHandlerHandleContextCancelled(t *testing.T) {
	// Arrange
	mockUsecase := new(MockDiaryAnalysisUsecase)
	log := slog.Default()

	event := &diarydomain.DiaryCreatedEvent{
		DiaryID:  uuid.New(),
		UserID:   uuid.New(),
		FamilyID: uuid.New(),
		Content:  "test content",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	mockUsecase.On("Analyze", mock.Anything, mock.MatchedBy(func (e *domain.DiaryCreatedEvent) bool {
		return e.DiaryID == event.DiaryID && e.UserID == event.UserID && e.FamilyID == event.FamilyID && e.Content == event.Content
	})).Return(nil, context.Canceled)

	handler := NewDiaryEventHandler(mockUsecase, log)

	eventBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}
	// Act
	err = handler.Handle(ctx, "diary.created", eventBytes)

	// Assert
	assert.Error(t, err)
	mockUsecase.AssertExpectations(t)
}