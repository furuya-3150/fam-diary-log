package usecase

import (
	"context"
	"log/slog"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/gateway"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
)

const (
	CheckAccuracyDefaultScore = 100
)

type DiaryAnalysisUsecase interface {
	Analyze(ctx context.Context, event *domain.DiaryCreatedEvent) (*domain.DiaryAnalysis, error)
}

type diaryAnalysisUsecase struct {
	ar         repository.DiaryAnalysisRepository
	nlpGateway gateway.NLPGateway
}

func NewDiaryAnalysisUsecase(
	ar repository.DiaryAnalysisRepository,
	nlpGateway gateway.NLPGateway,
) DiaryAnalysisUsecase {
	return &diaryAnalysisUsecase{
		ar:         ar,
		nlpGateway: nlpGateway,
	}
}

func NewDiaryAnalysisUsecaseWithNLPGateway(ar repository.DiaryAnalysisRepository, nlpGateway gateway.NLPGateway) DiaryAnalysisUsecase {
	return &diaryAnalysisUsecase{
		ar:         ar,
		nlpGateway: nlpGateway,
	}
}

// Analyze performs diary content analysis
func (u *diaryAnalysisUsecase) Analyze(ctx context.Context, event *domain.DiaryCreatedEvent) (*domain.DiaryAnalysis, error) {
	if event.DiaryID == uuid.Nil || event.UserID == uuid.Nil || event.FamilyID == uuid.Nil {
		return nil, &errors.ValidationError{Message: "diary_id, user_id, and family_id are required"}
	}

	if event.Content == "" {
		return nil, &errors.ValidationError{Message: "content is required"}
	}

	if u.nlpGateway == nil {
		return nil, &errors.LogicError{Message: "NLP gateway not configured"}
	}

	if event.DiaryID == uuid.Nil || event.UserID == uuid.Nil || event.FamilyID == uuid.Nil {
		return nil, &errors.ValidationError{Message: "diary_id, user_id, and family_id are required"}
	}

	// Perform analysis
	analysis := &domain.DiaryAnalysis{
		ID:            uuid.New(),
		DiaryID:       event.DiaryID,
		UserID:        event.UserID,
		FamilyID:      event.FamilyID,
		CharCount:     len([]rune(event.Content)),
		SentenceCount: u.countSentences(event.Content),
	}

	// Check accuracy (get suggestion count from gateway)
	suggestionCount, err := u.nlpGateway.CheckAccuracy(ctx, event.Content)
	if err == nil {
		// Calculate accuracy score using domain logic
		calculator := domain.NewAccuracyScoreCalculator()
		analysis.AccuracyScore = calculator.Calculate(suggestionCount)
	} else {
		analysis.AccuracyScore = CheckAccuracyDefaultScore
		slog.Error("Failed to check accuracy", "error", err)
	}

	// Store result
	_, err = u.ar.Create(ctx, analysis)
	if err != nil {
		return nil, err
	}

	return analysis, nil
}

// countSentences counts sentences in content
func (u *diaryAnalysisUsecase) countSentences(content string) int {
	count := 0
	for _, r := range content {
		if r == '。' || r == '！' || r == '？' {
			count++
		}
	}
	return count
}
