package jwt

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims structure
type Claims struct {
	UserID    uuid.UUID `json:"sub"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Provider  string    `json:"provider"`
	FamilyID  uuid.UUID `json:"family_id"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"iat"`
	ExpiresAt time.Time `json:"exp"`
	Issuer    string    `json:"iss"`
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString, secret string) (*Claims, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			slog.Debug("Unexpected signing method", "alg", token.Header["alg"])
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		slog.Debug("Failed to parse token", "error", err)
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Verify token is valid
	if !token.Valid {
		slog.Debug("Invalid token")
		return nil, fmt.Errorf("invalid token")
	}

	// Extract and validate claims
	claims, err := ParseClaims(token)
	if err != nil {
		slog.Debug("Failed to parse token", "error", err)
		return nil, err
	}

	// Check expiration
	if claims.ExpiresAt.Before(time.Now()) {
		slog.Debug("Token expired", "exp", claims.ExpiresAt)
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}

// ParseClaims extracts claims from a parsed JWT token
func ParseClaims(token *jwt.Token) (*Claims, error) {
	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Debug("Invalid claims format")
		return nil, fmt.Errorf("invalid claims format")
	}
	slog.Debug("Parsed claims", "claims", mapClaims)

	// Parse UserID
	subStr, ok := mapClaims["sub"].(string)
	if !ok {
		slog.Debug("Invalid or missing 'sub' claim")
		return nil, fmt.Errorf("invalid or missing 'sub' claim")
	}
	userID, err := uuid.Parse(subStr)
	if err != nil {
		slog.Debug("Invalid user ID format", "error", err)
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Parse FamilyID
	familyIDStr, ok := mapClaims["family_id"].(string)
	familyID := uuid.Nil
	if ok {
		familyID, err = uuid.Parse(familyIDStr)
		if err != nil {
			slog.Debug("Invalid family ID format", "error", err)
			return nil, fmt.Errorf("invalid family ID format: %w", err)
		}
	}

	// Parse role (optional)
	role, _ := mapClaims["role"].(string)

	// Parse issuer (optional)
	issuer, _ := mapClaims["iss"].(string)

	// Parse timestamps
	iat, err := parseTimestamp(mapClaims["iat"])
	if err != nil {
		slog.Debug("Invalid 'iat' claim", "error", err)
		return nil, fmt.Errorf("invalid 'iat' claim: %w", err)
	}

	exp, err := parseTimestamp(mapClaims["exp"])
	if err != nil {
		slog.Debug("Invalid 'exp' claim", "error", err)
		return nil, fmt.Errorf("invalid 'exp' claim: %w", err)
	}

	return &Claims{
		UserID: userID,
		// Email:     email,
		// Name:      name,
		// Provider:  provider,
		FamilyID:  familyID,
		Role:      role,
		IssuedAt:  iat,
		ExpiresAt: exp,
		Issuer:    issuer,
	}, nil
}

// parseTimestamp converts a JWT timestamp (float64 or int) to time.Time
func parseTimestamp(claim interface{}) (time.Time, error) {
	switch v := claim.(type) {
	case float64:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	default:
		return time.Time{}, fmt.Errorf("invalid timestamp type")
	}
}

// GetUserID extracts the user ID from a token string (without full validation)
// Use ValidateToken for secure validation
func GetUserID(tokenString, secret string) (uuid.UUID, error) {
	claims, err := ValidateToken(tokenString, secret)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// ValidateAndGetClaims validates a JWT token and returns the full claims
// This is the primary function to use for authentication
func ValidateAndGetClaims(tokenString, secret string) (*Claims, error) {
	return ValidateToken(tokenString, secret)
}
