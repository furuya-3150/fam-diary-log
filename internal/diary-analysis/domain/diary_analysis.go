package domain

import (
	"time"

	"github.com/google/uuid"
)

type DiaryAnalysis struct {
	ID            uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	DiaryID       uuid.UUID `gorm:"type:uuid;not null" json:"diary_id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	FamilyID      uuid.UUID `gorm:"type:uuid;not null" json:"family_id"`
	CharCount     int       `gorm:"not null;default:0" json:"char_count"`
	SentenceCount int       `gorm:"not null;default:0" json:"sentence_count"`
	AccuracyScore int       `gorm:"default:0" json:"accuracy_score"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name
func (DiaryAnalysis) TableName() string {
	return "diary_analyses"
}

// GetWeekCharCountRequest represents a request to get character count for a week
type GetWeekCharCountRequest struct {
	Date   string    `json:"date"` // YYYY-MM-DD format
	UserID uuid.UUID `json:"user_id"`
}
