package domain

import (
	"time"

	"github.com/google/uuid"
)

type Family struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type Role int

const (
	RoleUnknown Role = iota
	RoleAdmin
	RoleMember
)

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleMember:
		return "member"
	default:
		return "unknown"
	}
}

type FamilyMember struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	FamilyID  uuid.UUID `gorm:"type:uuid;not null;index"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Role      Role      `gorm:"type:int;not null"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type FamilyInvitation struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey"`
	FamilyID        uuid.UUID  `gorm:"type:uuid;not null;index"`
	InviterUserID   uuid.UUID  `gorm:"type:uuid;not null"`
	InvitationToken string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	ExpiresAt       time.Time  `gorm:"not null"`
	CreatedAt       time.Time  `gorm:"not null"`
	UpdatedAt       time.Time  `gorm:"not null"`
}
