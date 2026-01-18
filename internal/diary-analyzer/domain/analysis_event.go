package domain

import (
	"time"

	"github.com/google/uuid"
)

type DiaryAnalysis struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	DiaryID            uuid.UUID `gorm:"column:diary_id;type:uuid;not null"`
	UserID             uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	FamilyID           uuid.UUID `gorm:"column:family_id;type:uuid;not null"`
	CharCount          int       `gorm:"column:char_count;type:integer"`
	SentenceCount      int       `gorm:"column:sentence_count;type:integer"`
	AccuracyScore      int       `gorm:"column:accuracy_score;type:integer"`
	WritingTimeSeconds int       `gorm:"column:writing_time_seconds;type:integer"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
