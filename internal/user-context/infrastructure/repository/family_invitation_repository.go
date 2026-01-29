package repository

import (
	"context"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyInvitationRepository interface {
	CreateInvitation(ctx context.Context, invitation *domain.FamilyInvitation) error
	UpdateInvitationTokenAndExpires(ctx context.Context, familyID, inviterUserID uuid.UUID, token string, expiresAt time.Time) error
	FindInvitationByFamilyID(ctx context.Context, familyID uuid.UUID) (*domain.FamilyInvitation, error)
}

type familyInvitationRepository struct {
	dm *db.DBManager
}

func NewFamilyInvitationRepository(dm *db.DBManager) FamilyInvitationRepository {
	return &familyInvitationRepository{dm: dm}
}

func (r *familyInvitationRepository) CreateInvitation(ctx context.Context, invitation *domain.FamilyInvitation) error {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Create(invitation).Error; err != nil {
		return err
	}
	return nil
}

func (r *familyInvitationRepository) UpdateInvitationTokenAndExpires(ctx context.Context, familyID, inviterUserID uuid.UUID, token string, expiresAt time.Time) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Model(&domain.FamilyInvitation{}).
		Where("family_id = ? AND inviter_user_id = ?", familyID, inviterUserID).
		Updates(map[string]interface{}{"invitation_token": token, "expires_at": expiresAt, "updated_at": time.Now()}).Error
}

func (r *familyInvitationRepository) FindInvitationByFamilyID(ctx context.Context, familyID uuid.UUID) (*domain.FamilyInvitation, error) {
	dbConn := r.dm.DB(ctx)
	var inv domain.FamilyInvitation
	err := dbConn.Where("family_id = ?", familyID).First(&inv).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return &inv, nil
}
