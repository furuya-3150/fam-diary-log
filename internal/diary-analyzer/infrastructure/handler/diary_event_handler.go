package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/domain"
	"github.com/furuya-3150/fam-diary-log/internal/diary-analyzer/usecase"
)

// DiaryEventHandler handles diary-specific events
type DiaryEventHandler struct {
	analyzerService usecase.DiaryAnalysisUsecase
	l               *slog.Logger
}

// NewDiaryEventHandler creates a new DiaryEventHandler
func NewDiaryEventHandler(analyzerService usecase.DiaryAnalysisUsecase, l *slog.Logger) *DiaryEventHandler {
	return &DiaryEventHandler{
		analyzerService: analyzerService,
		l:               l,
	}
}

// Handle handles events based on routing key
func (h *DiaryEventHandler) Handle(ctx context.Context, routingKey string, content []byte) error {
	switch routingKey {
	case "diary.created":
		return h.handleDiaryCreated(ctx, content)
	default:
		return fmt.Errorf("unknown routing key: %s", routingKey)
	}
}

// handleDiaryCreated handles diary.created events
func (h *DiaryEventHandler) handleDiaryCreated(ctx context.Context, content []byte) error {
	log.Println("content", "data", string(content))
	// Unmarshal to generic event
	var diaryCreatedEvent domain.DiaryCreatedEvent
	if err := json.Unmarshal(content, &diaryCreatedEvent); err != nil {
		return fmt.Errorf("invalid event type for diary.created %v", err)
	}


	// Call analyzer service to analyze the diary
	if _, err := h.analyzerService.Analyze(ctx, &diaryCreatedEvent); err != nil {
		h.l.Error("failed to analyze diary", "diary_id", diaryCreatedEvent.DiaryID, "error", err.Error())
		return err
	}

	return nil
}
