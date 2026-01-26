package dto

import "github.com/google/uuid"

type EditUserRequest struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
}
