package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateDiaryRequest represents a request to create a diary
type CreateDiaryRequest struct {
	Title              string `json:"title" validate:"required,min=1,max=255"`
	Content            string `json:"content" validate:"required,min=1"`
	WritingTimeSeconds int    `json:"writing_time_seconds" validate:"required,min=0"`
}

type DiaryResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FamilyID  uuid.UUID `json:"family_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
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
