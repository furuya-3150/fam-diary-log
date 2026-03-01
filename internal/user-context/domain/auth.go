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
	ID         uuid.UUID    `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid();"`
	Email      string       `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Name       string       `json:"name" gorm:"type:varchar(255);not null"`
	Provider   AuthProvider `json:"provider" gorm:"type:varchar(50);not null;uniqueIndex:idx_provider_id"`
	ProviderID string       `json:"provider_id" gorm:"type:varchar(255);not null;uniqueIndex:idx_provider_id"` // ID from OAuth provider
	CreatedAt  time.Time    `json:"created_at" gorm:"autoCreateTime;"`
	UpdatedAt  time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}

// RefreshToken represents a refresh token stored in the DB
type RefreshToken struct {
	ID        uuid.UUID `json:"id"         gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id"    gorm:"type:uuid;not null;index"`
	Token     string    `json:"token"      gorm:"type:varchar(512);uniqueIndex;not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	Revoked   bool      `json:"revoked"    gorm:"default:false"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// TableName specifies the table name for RefreshToken model
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}