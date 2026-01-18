package domain

// AccuracyScoreCalculator calculates accuracy score based on suggestions
type AccuracyScoreCalculator struct{}

// NewAccuracyScoreCalculator creates a new AccuracyScoreCalculator
func NewAccuracyScoreCalculator() *AccuracyScoreCalculator {
	return &AccuracyScoreCalculator{}
}

// Calculate calculates accuracy score from suggestion count
// Each suggestion reduces score by 10 points, minimum score is 0
func (c *AccuracyScoreCalculator) Calculate(suggestionCount int) int {
	score := 100 - (suggestionCount * 10)
	if score < 0 {
		score = 0
	}
	return score
}
