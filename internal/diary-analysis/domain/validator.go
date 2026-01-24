package domain

import (
	"fmt"
	"time"
)

// ValidateYYYYMMDDFormat validates date in YYYY-MM-DD format
func ValidateYYYYMMDDFormat(dateStr string) (time.Time, error) {
	if dateStr == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format, expected YYYY-MM-DD: %v", err)
	}

	return parsedDate, nil
}
