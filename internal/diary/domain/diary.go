package domain

import (
	"time"

	"github.com/google/uuid"
)

// domainがgorm（技術）にするが開発コストを下げるため容認
type Diary struct {
	ID        uuid.UUID `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"column:user_id;type:uuid;not null"`
	FamilyID  uuid.UUID `gorm:"column:family_id;type:uuid;not null"`
	Title     string    `gorm:"column:title;type:varchar(255)"`
	Content   string    `gorm:"column:content;type:text"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}
