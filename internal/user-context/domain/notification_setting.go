package domain

import (
	"time"

	"github.com/google/uuid"
)

type NotificationSetting struct {
	ID                 uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();"`
	UserID             uuid.UUID `gorm:"type:uuid;not null"`
	FamilyID           uuid.UUID `gorm:"type:uuid;not null"`
	PostCreatedEnabled bool      `gorm:"not null;default:true"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}
