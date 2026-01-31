package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FamilyJoinRequestRepository interface {
	CreateJoinRequest(ctx context.Context, req *domain.FamilyJoinRequest) error
	FindPendingRequest(ctx context.Context, familyID, userID uuid.UUID) (*domain.FamilyJoinRequest, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.FamilyJoinRequest, error)
	UpdateStatusByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error
}

type familyJoinRequestRepository struct {
	dm *db.DBManager
}

func NewFamilyJoinRequestRepository(dm *db.DBManager) FamilyJoinRequestRepository {
	return &familyJoinRequestRepository{dm: dm}
}

func (r *familyJoinRequestRepository) CreateJoinRequest(ctx context.Context, req *domain.FamilyJoinRequest) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Create(req).Error
}

func (r *familyJoinRequestRepository) FindPendingRequest(ctx context.Context, familyID, userID uuid.UUID) (*domain.FamilyJoinRequest, error) {
	dbConn := r.dm.DB(ctx)
	var jr domain.FamilyJoinRequest
	err := dbConn.Where("family_id = ? AND user_id = ? AND status = ?", familyID, userID, int(domain.JoinRequestStatusPending)).First(&jr).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &jr, nil
}

func (r *familyJoinRequestRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.FamilyJoinRequest, error) {
	dbConn := r.dm.DB(ctx)
	var jr domain.FamilyJoinRequest
	err := dbConn.Where("id = ?", id).First(&jr).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &jr, nil
}

func (r *familyJoinRequestRepository) UpdateStatusByID(ctx context.Context, id uuid.UUID, updates map[string]interface{}) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Model(&domain.FamilyJoinRequest{}).Where("id = ?", id).Updates(updates).Error
}
