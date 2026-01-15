package datetime

import (
	"testing"
	"time"
)

func TestGetWeekRange_Monday(t *testing.T) {
	// Monday, January 13, 2026
	date := time.Date(2026, 1, 13, 15, 30, 45, 0, time.UTC)

	monday, sunday := GetWeekRange(date)

	// Expected Monday
	expectedMonday := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	if !monday.Equal(expectedMonday) {
		t.Errorf("expected monday %v, got %v", expectedMonday, monday)
	}

	// Expected Sunday
	expectedSunday := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)
	if !sunday.Equal(expectedSunday) {
		t.Errorf("expected sunday %v, got %v", expectedSunday, sunday)
	}
}

// TestGetWeekRange_Wednesday tests GetWeekRange with a Wednesday date
func TestGetWeekRange_Wednesday(t *testing.T) {
	// Wednesday, January 14, 2026
	date := time.Date(2026, 1, 14, 10, 30, 0, 0, time.UTC)

	monday, sunday := GetWeekRange(date)

	// Expected Monday
	expectedMonday := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	if !monday.Equal(expectedMonday) {
		t.Errorf("expected monday %v, got %v", expectedMonday, monday)
	}

	// Expected Sunday
	expectedSunday := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)
	if !sunday.Equal(expectedSunday) {
		t.Errorf("expected sunday %v, got %v", expectedSunday, sunday)
	}
}

// TestGetWeekRange_Sunday tests GetWeekRange with a Sunday date
func TestGetWeekRange_Sunday(t *testing.T) {
	// Sunday, January 18, 2026
	date := time.Date(2026, 1, 18, 20, 0, 0, 0, time.UTC)

	monday, sunday := GetWeekRange(date)

	// Expected Monday
	expectedMonday := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	if !monday.Equal(expectedMonday) {
		t.Errorf("expected monday %v, got %v", expectedMonday, monday)
	}

	// Expected Sunday
	expectedSunday := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)
	if !sunday.Equal(expectedSunday) {
		t.Errorf("expected sunday %v, got %v", expectedSunday, sunday)
	}
}

// TestGetWeekRange_MondayTime tests GetWeekRange returns correct time for Monday
func TestGetWeekRange_MondayTime(t *testing.T) {
	// Tuesday, January 15, 2026
	date := time.Date(2026, 1, 15, 14, 20, 10, 0, time.UTC)

	monday, _ := GetWeekRange(date)

	// Monday should be at 00:00:00
	if monday.Hour() != 0 || monday.Minute() != 0 || monday.Second() != 0 {
		t.Errorf("monday should be at 00:00:00, got %02d:%02d:%02d", monday.Hour(), monday.Minute(), monday.Second())
	}
}

// TestGetWeekRange_SundayTime tests GetWeekRange returns correct time for Sunday
func TestGetWeekRange_SundayTime(t *testing.T) {
	// Tuesday, January 15, 2026
	date := time.Date(2026, 1, 15, 14, 20, 10, 0, time.UTC)

	_, sunday := GetWeekRange(date)

	// Sunday should be at 23:59:59
	if sunday.Hour() != 23 || sunday.Minute() != 59 || sunday.Second() != 59 {
		t.Errorf("sunday should be at 23:59:59, got %02d:%02d:%02d", sunday.Hour(), sunday.Minute(), sunday.Second())
	}
}

// TestGetWeekRange_DifferentDatesInSameWeek tests that different dates in the same week return same range
func TestGetWeekRange_DifferentDatesInSameWeek(t *testing.T) {
	// Multiple dates in the same week (Jan 12-18, 2026)
	dates := []time.Time{
		time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC), // Monday
		time.Date(2026, 1, 13, 10, 0, 0, 0, time.UTC), // Tuesday
		time.Date(2026, 1, 14, 10, 0, 0, 0, time.UTC), // Wednesday
		time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), // Thursday
		time.Date(2026, 1, 16, 10, 0, 0, 0, time.UTC), // Friday
		time.Date(2026, 1, 17, 10, 0, 0, 0, time.UTC), // Saturday
		time.Date(2026, 1, 18, 10, 0, 0, 0, time.UTC), // Sunday
	}

	expectedMonday := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	expectedSunday := time.Date(2026, 1, 18, 23, 59, 59, 999999999, time.UTC)

	for _, date := range dates {
		monday, sunday := GetWeekRange(date)
		if !monday.Equal(expectedMonday) {
			t.Errorf("for date %v, expected monday %v, got %v", date, expectedMonday, monday)
		}
		if !sunday.Equal(expectedSunday) {
			t.Errorf("for date %v, expected sunday %v, got %v", date, expectedSunday, sunday)
		}
	}
}

// TestGetWeekRange_DifferentWeeks tests different weeks return different ranges
func TestGetWeekRange_DifferentWeeks(t *testing.T) {
	// Monday of week 1
	date1 := time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC)
	// Monday of week 2
	date2 := time.Date(2026, 1, 19, 10, 0, 0, 0, time.UTC)

	monday1, sunday1 := GetWeekRange(date1)
	monday2, sunday2 := GetWeekRange(date2)

	// Different weeks should return different ranges
	if monday1.Equal(monday2) || sunday1.Equal(sunday2) {
		t.Error("different weeks should return different ranges")
	}

	// Verify that date1 is in week 1 and date2 is in week 2
	if monday1.Day() != 12 || monday2.Day() != 19 {
		t.Errorf("week 1 monday should be 12, week 2 monday should be 19, got %d and %d", monday1.Day(), monday2.Day())
	}
}
