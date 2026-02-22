package domain

import (
	"time"

	"github.com/google/uuid"
)

// DiaryCreatedEvent represents an event when a diary is created
type DiaryCreatedEvent struct {
	ID                 string    `json:"id"`
	DiaryID            uuid.UUID `json:"diary_id"`
	UserID             uuid.UUID `json:"user_id"`
	FamilyID           uuid.UUID `json:"family_id"`
	Title              string    `json:"title"`
	Content            string    `json:"content"`
	WritingTimeSeconds int       `json:"writing_time_seconds"`
	Timestamp          time.Time `json:"timestamp"`
}

func (e *DiaryCreatedEvent) EventType() string {
	return "diary.created"
}

// NewDiaryCreatedEvent creates a new DiaryCreatedEvent
func NewDiaryCreatedEvent(diaryID, userID, familyID uuid.UUID, title, content string, writingTimeSeconds int) *DiaryCreatedEvent {
	return &DiaryCreatedEvent{
		ID:                 uuid.New().String(),
		DiaryID:            diaryID,
		UserID:             userID,
		FamilyID:           familyID,
		Title:              title,
		Content:            content,
		WritingTimeSeconds: writingTimeSeconds,
		Timestamp:          time.Now(),
	}
}
