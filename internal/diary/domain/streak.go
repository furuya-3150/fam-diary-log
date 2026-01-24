package domain

import (
	"time"

	"github.com/google/uuid"
)

type Streak struct {
	UserID       uuid.UUID `gorm:"primaryKey;type:uuid;not null" json:"user_id"`
	FamilyID     uuid.UUID `gorm:"primaryKey;type:uuid;not null" json:"family_id"`
	CurrentStreak int       `gorm:"not null;default:0" json:"current_streak"`
	LastPostDate  *time.Time `gorm:"type:date" json:"last_post_date"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName specifies the table name
func (Streak) TableName() string {
	return "streaks"
}
