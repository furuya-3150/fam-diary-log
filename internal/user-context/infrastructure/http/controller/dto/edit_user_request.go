package dto

import "github.com/google/uuid"

type EditUserRequest struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name" validate:"required,min=1,max=100"`
	Email string    `json:"email" validate:"required,email,max=255"`
}
