package domain

import (
	"time"

	"github.com/google/uuid"
)

// AuthProvider represents the authentication provider type
type AuthProvider string

const (
	AuthProviderGoogle AuthProvider = "google"
)

// User represents a user in the system
type User struct {
	ID         uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey"`
	Email      string       `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Name       string       `json:"name" gorm:"type:varchar(255);not null"`
	Provider   AuthProvider `json:"provider" gorm:"type:varchar(50);not null;uniqueIndex:idx_provider_id"`
	ProviderID string       `json:"provider_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_id"` // ID from OAuth provider
	CreatedAt  time.Time    `json:"created_at" gorm:"not null"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"not null"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
