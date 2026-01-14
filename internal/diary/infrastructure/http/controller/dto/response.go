package dto

import (
	"time"

	"github.com/google/uuid"
)

type DiaryResponse struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	FamilyID  uuid.UUID
	Title     string
	Content   string
	CreatedAt time.Time
}
