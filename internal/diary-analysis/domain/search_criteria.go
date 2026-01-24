package domain

import (
	"time"

	"github.com/google/uuid"
)

// DiaryAnalysisSearchCriteria represents the criteria for searching diary analyses
type DiaryAnalysisSearchCriteria struct {
	UserID    uuid.UUID
	WeekStart time.Time
	WeekEnd   time.Time
	Columns   []string // columns to select in the query
}
