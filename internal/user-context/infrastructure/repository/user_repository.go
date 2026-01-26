package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByProviderID(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
}

type userRepository struct {
	dm *db.DBManager
}

func NewUserRepository(dm *db.DBManager) UserRepository {
	return &userRepository{dm: dm}
}

func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	dbConn := r.dm.DB(ctx)
	var user domain.User
	if err := dbConn.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByProviderID(ctx context.Context, provider domain.AuthProvider, providerID string) (*domain.User, error) {
	dbConn := r.dm.DB(ctx)
	var user domain.User
	if err := dbConn.Where("provider = ? AND provider_id = ?", provider, providerID).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	dbConn := r.dm.DB(ctx)
	var user domain.User
	if err := dbConn.Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}