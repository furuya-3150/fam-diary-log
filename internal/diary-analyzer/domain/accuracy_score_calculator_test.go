package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAccuracyScoreCalculatorCalculate(t *testing.T) {
	tests := []struct {
		name            string
		suggestionCount int
		expectedScore   int
	}{
		{
			name:            "Zero suggestions - maximum score",
			suggestionCount: 0,
			expectedScore:   100,
		},
		{
			name:            "Five suggestions - middle equivalence class",
			suggestionCount: 5,
			expectedScore:   50,
		},
		{
			name:            "Ten suggestions - boundary at zero",
			suggestionCount: 10,
			expectedScore:   0,
		},
		{
			name:            "More than ten - clamped to zero",
			suggestionCount: 15,
			expectedScore:   0,
		},
	}

	calculator := NewAccuracyScoreCalculator()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculator.Calculate(tt.suggestionCount)
			assert.Equal(t, tt.expectedScore, result)
		})
	}
}
