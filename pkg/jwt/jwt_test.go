package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestValidateToken(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New()
	familyID := uuid.New()

	// Generate a valid token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"email":     "test@example.com",
		"name":      "Test User",
		"provider":  "email",
		"family_id": familyID.String(),
		"iat":       now.Unix(),
		"exp":       now.Add(1 * time.Hour).Unix(),
		"iss":       "fam-diary-log",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Test validation
	parsedClaims, err := ValidateToken(tokenString, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Verify claims
	if parsedClaims.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, parsedClaims.UserID)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	secret := "test-secret-key"
	userID := uuid.New()
	familyID := uuid.New()

	// Generate an expired token
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"email":     "test@example.com",
		"name":      "Test User",
		"provider":  "email",
		"family_id": familyID.String(),
		"iat":       now.Add(-2 * time.Hour).Unix(),
		"exp":       now.Add(-1 * time.Hour).Unix(), // Expired 1 hour ago
		"iss":       "fam-diary-log",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Test validation should fail
	_, err = ValidateToken(tokenString, secret)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateToken_InvalidSecret(t *testing.T) {
	secret := "test-secret-key"
	wrongSecret := "wrong-secret-key"
	userID := uuid.New()
	familyID := uuid.New()

	// Generate a token with correct secret
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":       userID.String(),
		"email":     "test@example.com",
		"name":      "Test User",
		"provider":  "email",
		"family_id": familyID.String(),
		"iat":       now.Unix(),
		"exp":       now.Add(1 * time.Hour).Unix(),
		"iss":       "fam-diary-log",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Test validation with wrong secret should fail
	_, err = ValidateToken(tokenString, wrongSecret)
	if err == nil {
		t.Error("Expected error for invalid secret, got nil")
	}
}
