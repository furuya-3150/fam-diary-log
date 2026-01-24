package dto

import (
	"time"

	"github.com/google/uuid"
)

type DiaryResponse struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
}

type StreakResponse struct {
	UserID        uuid.UUID  `json:"user_id"`
	FamilyID      uuid.UUID  `json:"family_id"`
	CurrentStreak int        `json:"current_streak"`
	LastPostDate  *time.Time `json:"last_post_date"`
}
