package domain

import (
	"time"

	"github.com/google/uuid"
)

type DiarySearchCriteria struct {
	FamilyID  uuid.UUID
	UserID    uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

// DiaryCountCriteria represents the criteria for counting diaries
type DiaryCountCriteria struct {
	FamilyID  uuid.UUID
	YearMonth string
}
