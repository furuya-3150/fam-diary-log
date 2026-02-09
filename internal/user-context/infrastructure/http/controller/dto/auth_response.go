package dto

import (
	"github.com/google/uuid"
)

type AuthResponse struct {
	User        *UserResponse `json:"user"`
	AccessToken string        `json:"access_token"`
	ExpiresIn   int64         `json:"expires_in"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
}
