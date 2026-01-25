package repository

import (
	"context"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/helper"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// IntegrationTestDeps holds dependencies for integration tests
type IntegrationTestDeps struct {
	DB   *gorm.DB
	Repo AuthRepository
}

func TestAuthRepositoryIntegration_CreateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	userID := uuid.New()

	user := &domain.User{
		ID:         userID,
		Email:      "test@example.com",
		Name:       "Test User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-12345",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Act
	result, err := deps.Repo.CreateUser(ctx, user)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "test@example.com", result.Email)
	assert.Equal(t, "Test User", result.Name)
	assert.Equal(t, domain.AuthProviderGoogle, result.Provider)
	assert.Equal(t, "google-12345", result.ProviderID)

	// Verify user was persisted
	var persistedUser domain.User
	err = deps.DB.First(&persistedUser, "id = ?", userID).Error
	require.NoError(t, err)
	assert.Equal(t, userID, persistedUser.ID)
	assert.Equal(t, "test@example.com", persistedUser.Email)
}

func TestAuthRepositoryIntegration_CreateUser_DuplicateEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()

	// Create first user
	user1 := &domain.User{
		ID:         uuid.New(),
		Email:      "duplicate@example.com",
		Name:       "First User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-11111",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := deps.Repo.CreateUser(ctx, user1)
	require.NoError(t, err)

	// Try to create user with duplicate email
	user2 := &domain.User{
		ID:         uuid.New(),
		Email:      "duplicate@example.com", // Same email
		Name:       "Second User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-22222",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Act
	result, err := deps.Repo.CreateUser(ctx, user2)

	// Assert - should fail due to unique constraint
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthRepositoryIntegration_CreateUser_DuplicateProviderID(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()

	// Create first user
	user1 := &domain.User{
		ID:         uuid.New(),
		Email:      "user1@example.com",
		Name:       "First User",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-same-id",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := deps.Repo.CreateUser(ctx, user1)
	require.NoError(t, err)

	// Try to create user with same provider and provider_id
	user2 := &domain.User{
		ID:         uuid.New(),
		Email:      "user2@example.com",
		Name:       "Second User",
		Provider:   domain.AuthProviderGoogle, // Same provider
		ProviderID: "google-same-id",          // Same provider ID
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Act
	result, err := deps.Repo.CreateUser(ctx, user2)

	// Assert - should fail due to unique constraint on (provider, provider_id)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthRepositoryIntegration_GetUserByProviderID_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	userID := uuid.New()

	// Create user
	user := &domain.User{
		ID:         userID,
		Email:      "findme@example.com",
		Name:       "Find Me",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-find-me",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := deps.Repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Act
	result, err := deps.Repo.GetUserByProviderID(ctx, domain.AuthProviderGoogle, "google-find-me")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "findme@example.com", result.Email)
	assert.Equal(t, "Find Me", result.Name)
	assert.Equal(t, domain.AuthProviderGoogle, result.Provider)
	assert.Equal(t, "google-find-me", result.ProviderID)
}

func TestAuthRepositoryIntegration_GetUserByProviderID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()

	// Act - search for non-existent user
	result, err := deps.Repo.GetUserByProviderID(ctx, domain.AuthProviderGoogle, "non-existent-id")

	// Assert - データが存在しない場合、userもerrもnilであること
	assert.NoError(t, err, "err should be nil when user not found")
	assert.Nil(t, result, "user should be nil when not found")
}

func TestAuthRepositoryIntegration_GetUserByEmail_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	userID := uuid.New()

	// Create user
	user := &domain.User{
		ID:         userID,
		Email:      "findbyemail@example.com",
		Name:       "Find By Email",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-findbyemail",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := deps.Repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Act
	result, err := deps.Repo.GetUserByEmail(ctx, "findbyemail@example.com")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "findbyemail@example.com", result.Email)
	assert.Equal(t, "Find By Email", result.Name)
	assert.Equal(t, domain.AuthProviderGoogle, result.Provider)
	assert.Equal(t, "google-findbyemail", result.ProviderID)
}

func TestAuthRepositoryIntegration_GetUserByEmail_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()

	// Act - search for non-existent email
	result, err := deps.Repo.GetUserByEmail(ctx, "nonexistent@example.com")

	// Assert - データが存在しない場合、userもerrもnilであること
	assert.NoError(t, err, "err should be nil when user not found")
	assert.Nil(t, result, "user should be nil when not found")
}

func TestAuthRepositoryIntegration_GetUserByID_Found(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	userID := uuid.New()

	// Create user
	user := &domain.User{
		ID:         userID,
		Email:      "getbyid@example.com",
		Name:       "Get By ID",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-getbyid",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err := deps.Repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Act
	result, err := deps.Repo.GetUserByID(ctx, userID)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "getbyid@example.com", result.Email)
	assert.Equal(t, "Get By ID", result.Name)
}

func TestAuthRepositoryIntegration_GetUserByID_NotFound(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	nonExistentID := uuid.New()

	// Act
	result, err := deps.Repo.GetUserByID(ctx, nonExistentID)

	// Assert - データが存在しない場合、userもerrもnilであること
	assert.NoError(t, err, "err should be nil when user not found")
	assert.Nil(t, result, "user should be nil when not found")
}

func TestAuthRepositoryIntegration_UpdateUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Integration test - requires database setup")
	}

	deps := setupIntegrationTestDeps(t)
	defer teardownIntegrationTest(t, deps)

	ctx := context.Background()
	userID := uuid.New()

	// Create user
	user := &domain.User{
		ID:         userID,
		Email:      "update@example.com",
		Name:       "Original Name",
		Provider:   domain.AuthProviderGoogle,
		ProviderID: "google-update",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	createdUser, err := deps.Repo.CreateUser(ctx, user)
	require.NoError(t, err)

	// Update user
	createdUser.Name = "Updated Name"
	createdUser.Email = "updated@example.com"
	createdUser.UpdatedAt = time.Now()

	// Act
	result, err := deps.Repo.UpdateUser(ctx, createdUser)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, userID, result.ID)
	assert.Equal(t, "updated@example.com", result.Email)
	assert.Equal(t, "Updated Name", result.Name)

	// Verify update was persisted
	var persistedUser domain.User
	err = deps.DB.First(&persistedUser, "id = ?", userID).Error
	require.NoError(t, err)
	assert.Equal(t, "updated@example.com", persistedUser.Email)
	assert.Equal(t, "Updated Name", persistedUser.Name)
}


// setupIntegrationTestDeps sets up test database and dependencies
func setupIntegrationTestDeps(t *testing.T) *IntegrationTestDeps {
	t.Helper()

	godotenv.Load("../../../cmd/user-context/.env")

	// Setup test database
	dbManager := helper.SetupTestDB(t)

	return &IntegrationTestDeps{
		DB:   dbManager.GetGorm(),
		Repo: NewAuthRepository(dbManager),
	}
}

// teardownIntegrationTest cleans up after integration tests
func teardownIntegrationTest(t *testing.T, deps *IntegrationTestDeps) {
	t.Helper()
	if deps == nil {
		return
	}

	// Clean up database tables
	if deps.DB != nil {
		helper.TeardownTestDB(t, deps.DB)

	}
}
