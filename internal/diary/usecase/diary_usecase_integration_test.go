package usecase

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/helper"
	"github.com/furuya-3150/fam-diary-log/internal/diary/infrastructure/repository"
	pubpkg "github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/events"
	"gorm.io/gorm"
)

// IntegrationTestDeps holds dependencies for integration tests
type IntegrationTestDeps struct {
	DB        *gorm.DB
	TM        db.TransactionManager
	DR        repository.DiaryRepository
	SR        repository.StreakRepository
	Publisher pubpkg.Publisher
	Clock     clock.Clock
}

// EventCapture holds captured events for testing
type EventCapture struct {
	Events []events.Event
	mu     sync.Mutex
}

// MockPublisherWithCapture is a publisher that captures published events
type MockPublisherWithCapture struct {
	capture *EventCapture
}

func (m *MockPublisherWithCapture) Publish(ctx context.Context, event events.Event) error {
	m.capture.mu.Lock()
	defer m.capture.mu.Unlock()
	m.capture.Events = append(m.capture.Events, event)
	return nil
}

func (m *MockPublisherWithCapture) Close() error {
	return nil
}

// Note: These integration tests require a test database setup.

// TestDiaryUsecaseIntegration_CreateWithTransaction tests end-to-end diary creation with real transaction
func TestDiaryUsecaseIntegration_CreateWithTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	// Arrange
	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	userID := uuid.New()
	familyID := uuid.New()
	input := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Integration Test Diary",
		Content:  "This is an integration test diary",
	}

	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, input.Title, result.Title)
	assert.Equal(t, input.Content, result.Content)

	// Verify diary was persisted
	retrievedDiary := &domain.Diary{}
	dbErr := deps.DB.First(retrievedDiary, "id = ?", result.ID).Error
	assert.NoError(t, dbErr)
	assert.Equal(t, result.ID, retrievedDiary.ID)

	// Verify streak was created
	retrievedStreak := &domain.Streak{}
	streakErr := deps.DB.First(retrievedStreak, "user_id = ? AND family_id = ?", userID, familyID).Error
	assert.NoError(t, streakErr)
	assert.Equal(t, domain.DefaultStreakValue, retrievedStreak.CurrentStreak)
}

// TestDiaryUsecaseIntegration_StreakCalculationFlow tests complete streak calculation flow
func TestDiaryUsecaseIntegration_StreakCalculationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	// Arrange
	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	userID := uuid.New()
	familyID := uuid.New()

	// Day 1: Create first diary
	day1Time := time.Date(2026, 1, 13, 10, 0, 0, 0, time.Local)
	log.Println("Day 1 Time:", day1Time)
	clk1 := &clock.Fixed{Time: day1Time}
	usecase1 := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk1)

	diary1 := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Day 1 Diary",
		Content:  "First day entry",
	}

	result1, err := usecase1.Create(context.Background(), diary1)
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Verify initial streak
	streak1, err := deps.SR.Get(context.Background(), userID, familyID)
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultStreakValue, streak1.CurrentStreak)

	// Day 2: Create second diary (consecutive)
	day2Time := time.Date(2026, 1, 14, 10, 0, 0, 0, time.Local)
	clk2 := &clock.Fixed{Time: day2Time}
	usecase2 := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk2)

	diary2 := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Day 2 Diary",
		Content:  "Second day entry",
	}

	result2, err := usecase2.Create(context.Background(), diary2)
	require.NoError(t, err)
	require.NotNil(t, result2)

	// Verify streak incremented
	streak2, err := deps.SR.Get(context.Background(), userID, familyID)
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultStreakValue+1, streak2.CurrentStreak)

	// Day 4 (Gap): Create third diary (non-consecutive)
	day4Time := time.Date(2026, 1, 16, 10, 0, 0, 0, time.Local)
	clk4 := &clock.Fixed{Time: day4Time}
	usecase4 := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk4)

	diary4 := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Day 4 Diary",
		Content:  "Third day entry (gap of 1 day)",
	}

	result4, err := usecase4.Create(context.Background(), diary4)
	require.NoError(t, err)
	require.NotNil(t, result4)

	// Verify streak reset
	streak4, err := deps.SR.Get(context.Background(), userID, familyID)
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultStreakValue, streak4.CurrentStreak)
}

// TestDiaryUsecaseIntegration_PublishEventToBroker tests that diary creation publishes event
// Uses mock broker - focuses on DB integration
func TestDiaryUsecaseIntegration_PublishEventToBroker(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	// Arrange
	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	userID := uuid.New()
	familyID := uuid.New()
	input := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Event Test Diary",
		Content:  "This diary should trigger an event",
	}

	fixedTime := time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC)
	clk := &clock.Fixed{Time: fixedTime}
	usecase := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk)

	// Act
	result, err := usecase.Create(context.Background(), input)

	// Assert - diary created
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.ID)

	// Verify event was published to broker
	// If using MockPublisherWithCapture, check captured events
	if capturePub, ok := deps.Publisher.(*MockPublisherWithCapture); ok {
		assert.NotEmpty(t, capturePub.capture.Events, "Event should be published to broker")

		// Verify the event content
		for _, event := range capturePub.capture.Events {
			if diaryEvent, ok := event.(*domain.DiaryCreatedEvent); ok {
				assert.Equal(t, result.ID, diaryEvent.DiaryID)
				assert.Equal(t, userID, diaryEvent.UserID)
				assert.Equal(t, familyID, diaryEvent.FamilyID)
				assert.Equal(t, input.Content, diaryEvent.Content)
				return
			}
		}
		t.Fatal("DiaryCreatedEvent not found in published events")
	}
}

// TestDiaryUsecaseIntegration_StreakEventPublished tests that streak updates trigger events
// Uses mock broker - focuses on DB integration
func TestDiaryUsecaseIntegration_StreakEventPublished(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	// Arrange
	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	userID := uuid.New()
	familyID := uuid.New()

	// Create first diary
	diary1 := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Day 1",
		Content:  "First entry",
	}

	fixedTime1 := time.Date(2026, 1, 13, 10, 0, 0, 0, time.UTC)
	clk1 := &clock.Fixed{Time: fixedTime1}
	usecase1 := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk1)

	result1, err := usecase1.Create(context.Background(), diary1)
	require.NoError(t, err)
	require.NotNil(t, result1)

	// Verify streak was created with first diary
	streak, err := deps.SR.Get(context.Background(), userID, familyID)
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultStreakValue, streak.CurrentStreak)

	// Create second diary (consecutive day)
	diary2 := &domain.Diary{
		UserID:   userID,
		FamilyID: familyID,
		Title:    "Day 2",
		Content:  "Second entry",
	}

	fixedTime2 := time.Date(2026, 1, 14, 10, 0, 0, 0, time.UTC)
	clk2 := &clock.Fixed{Time: fixedTime2}
	usecase2 := NewDiaryUsecase(deps.TM, deps.DR, deps.SR, deps.Publisher, clk2)

	result2, err := usecase2.Create(context.Background(), diary2)
	require.NoError(t, err)
	require.NotNil(t, result2)

	// Verify streak was incremented
	updatedStreak, err := deps.SR.Get(context.Background(), userID, familyID)
	require.NoError(t, err)
	assert.Equal(t, domain.DefaultStreakValue+1, updatedStreak.CurrentStreak)

	// Verify both diary events were published
	if capturePub, ok := deps.Publisher.(*MockPublisherWithCapture); ok {
		assert.GreaterOrEqual(t, len(capturePub.capture.Events), 2, "At least 2 events should be published")

		// Verify both DiaryCreatedEvents exist
		eventCount := 0
		for _, event := range capturePub.capture.Events {
			if _, ok := event.(*domain.DiaryCreatedEvent); ok {
				eventCount++
			}
		}
		assert.Equal(t, 2, eventCount, "Should have 2 DiaryCreatedEvents")
	}
}

// setupIntegrationTestDeps sets up test database and dependencies
// Broker is mocked to focus on DB integration testing
func setupIntegrationTestDeps(t *testing.T) *IntegrationTestDeps {
	t.Helper()

	godotenv.Load("../../../cmd/diary-api/.env")
	// Setup test database
	dbManager := helper.SetupTestDB(t)

	// Use MockPublisherWithCapture for broker - prevents message accumulation in RabbitMQ
	capturePub := &MockPublisherWithCapture{
		capture: &EventCapture{},
	}

	return &IntegrationTestDeps{
		DB:        dbManager.GetGorm(),
		TM:        db.NewTransaction(dbManager),
		DR:        repository.NewDiaryRepository(dbManager),
		SR:        repository.NewStreakRepository(dbManager),
		Publisher: capturePub,
		Clock:     &clock.Real{},
	}
}

// teardownIntegrationTest cleans up after integration tests
func teardownIntegrationTest(t *testing.T, deps *IntegrationTestDeps) {
	t.Helper()
	if deps == nil {
		return
	}

	// Close publisher
	if deps.Publisher != nil {
		if err := deps.Publisher.Close(); err != nil {
			t.Logf("Warning: Error closing publisher: %s", err.Error())
		}
	}

	// Clean up database tables
	if deps.DB != nil {
		helper.TeardownTestDB(t, deps.DB)
	}
}
