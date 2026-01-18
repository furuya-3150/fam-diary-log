package gateway

import (
	"context"
)

// TextAnalysisResult represents the result of text analysis
type TextAnalysisResult struct {
	AccuracyScore float64
	Tokens        []Token
}

// Token represents a morphological token
type Token struct {
	Surface string
	ID      int
	Reading string
	POS     string
}

// NLPGateway defines the interface for external NLP services
type NLPGateway interface {
	// CheckAccuracy checks grammar and spelling accuracy
	CheckAccuracy(ctx context.Context, text string) (int, error)
}
