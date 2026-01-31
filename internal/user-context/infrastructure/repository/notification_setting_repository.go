package repository

import (
	"context"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type NotificationSettingRepository interface {
	GetByUserAndFamily(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error)
	Upsert(ctx context.Context, setting *domain.NotificationSetting) (*domain.NotificationSetting, error)
}

type notificationSettingRepository struct {
	dm *db.DBManager
}

func NewNotificationSettingRepository(dm *db.DBManager) NotificationSettingRepository {
	return &notificationSettingRepository{dm: dm}
}

func (r *notificationSettingRepository) GetByUserAndFamily(ctx context.Context, userID, familyID uuid.UUID) (*domain.NotificationSetting, error) {
	dbConn := r.dm.DB(ctx)
	var s domain.NotificationSetting
	err := dbConn.Where("user_id = ? AND family_id = ?", userID, familyID).First(&s).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *notificationSettingRepository) Upsert(ctx context.Context, setting *domain.NotificationSetting) (*domain.NotificationSetting, error) {
	dbConn := r.dm.DB(ctx)
	// ensure updated_at is set when updating via upsert
	now := time.Now()
	err := dbConn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}, {Name: "family_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"post_created_enabled": setting.PostCreatedEnabled,
			"updated_at":           now,
		}),
	}).Create(setting).Error

	if err != nil {
		return nil, err
	}
	// refresh from DB to return current state
	got, err := r.GetByUserAndFamily(ctx, setting.UserID, setting.FamilyID)
	if err != nil {
		return nil, err
	}
	return got, nil
}
