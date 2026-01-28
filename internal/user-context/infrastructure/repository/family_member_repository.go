package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
)

type FamilyMemberRepository interface {
	IsUserAlreadyMember(ctx context.Context, userID uuid.UUID) (bool, error)
	AddFamilyMember(ctx context.Context, member *domain.FamilyMember) error
}

type familyMemberRepository struct {
	dm *db.DBManager
}

func NewFamilyMemberRepository(dm *db.DBManager) FamilyMemberRepository {
	return &familyMemberRepository{dm: dm}
}

func (r *familyMemberRepository) IsUserAlreadyMember(ctx context.Context, userID uuid.UUID) (bool, error) {
	dbConn := r.dm.DB(ctx)
	var count int64
	err := dbConn.Model(&domain.FamilyMember{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *familyMemberRepository) AddFamilyMember(ctx context.Context, member *domain.FamilyMember) error {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Create(member).Error; err != nil {
		return err
	}
	return nil
}