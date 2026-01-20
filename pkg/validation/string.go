package validation

import (
	"fmt"
	"strings"
)

// NotEmpty validates that a string value is not empty
func NotEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// MaxLength validates that a string value does not exceed the maximum length
func MaxLength(value string, max int, fieldName string) error {
	if len(value) > max {
		return fmt.Errorf("%s は%d文字以内でなければなりません", fieldName, max)
	}
	return nil
}

// MinLength validates that a string value meets the minimum length
func MinLength(value string, min int, fieldName string) error {
	if len(value) < min {
		return fmt.Errorf("%s は%d文字以上でなければなりません", fieldName, min)
	}
	return nil
}

// Length validates that a string value is within a specific length range
func Length(value string, min, max int, fieldName string) error {
	if err := MinLength(value, min, fieldName); err != nil {
		return err
	}
	if err := MaxLength(value, max, fieldName); err != nil {
		return err
	}
	return nil
}

// NotEmptyAndMaxLength is a convenience function combining NotEmpty and MaxLength
func NotEmptyAndMaxLength(value string, max int, fieldName string) error {
	if err := NotEmpty(value, fieldName); err != nil {
		return err
	}
	if err := MaxLength(value, max, fieldName); err != nil {
		return err
	}
	return nil
}

// OneOf validates that a string value is one of the allowed options
func OneOf(value string, allowedValues []string, fieldName string) error {
	for _, allowed := range allowedValues {
		if value == allowed {
			return nil
		}
	}
	return fmt.Errorf("%s は許可された値のいずれかでなければなりません", fieldName)
}

// ValidateYearMonth validates year and month strings and returns them as integers
func ValidateYearMonth(year, month string) (int, int, error) {
	var y, m int
	_, err := fmt.Sscanf(year+"-"+month, "%d-%d", &y, &m)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid year or month format: expected YYYY-MM")
	}

	if y < 1 || y > 9999 {
		return 0, 0, fmt.Errorf("year must be between 1 and 9999")
	}

	if m < 1 || m > 12 {
		return 0, 0, fmt.Errorf("month must be between 1 and 12")
	}

	return y, m, nil
}
