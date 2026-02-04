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

// DiaryListQuery represents query parameters for listing diaries.
// target_date is required and must be in YYYY-MM-DD format.
type DiaryListQuery struct {
	TargetDate string `query:"target_date" validate:"required,datetime=2006-01-02"`
}
