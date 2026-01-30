package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateFamilyRequest struct {
	Name string `json:"name"`
}

type FamilyResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type InviteMembersRequest struct {
	FamilyID uuid.UUID `json:"family_id"`
	UserID   uuid.UUID `json:"user_id"`
	Emails   []string  `json:"emails"`
}

type InvitationInfo struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InviteMembersResponse struct {
	Invitations []InvitationInfo `json:"invitations"`
}

type ApplyRequest struct {
	Token string `json:"token"`
}
