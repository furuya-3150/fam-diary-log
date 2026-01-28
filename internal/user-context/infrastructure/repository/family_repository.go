package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
)

type FamilyRepository interface {
	CreateFamily(ctx context.Context, family *domain.Family) (*domain.Family, error)
}

type familyRepository struct {
	dm *db.DBManager
}

func NewFamilyRepository(dm *db.DBManager) FamilyRepository {
	return &familyRepository{dm: dm}
}

func (r *familyRepository) CreateFamily(ctx context.Context, family *domain.Family) (*domain.Family, error) {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Create(family).Error; err != nil {
		return family, err
	}
	return family, nil
}
