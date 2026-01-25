package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserContextResponse struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	FamilyID  uuid.UUID `json:"family_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
