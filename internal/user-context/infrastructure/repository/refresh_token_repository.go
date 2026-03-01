package repository

import (
	"context"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error)
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context, before time.Time) error
}

type refreshTokenRepository struct {
	dm *db.DBManager
}

func NewRefreshTokenRepository(dm *db.DBManager) RefreshTokenRepository {
	return &refreshTokenRepository{dm: dm}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) (*domain.RefreshToken, error) {
	dbConn := r.dm.DB(ctx)
	if err := dbConn.Create(token).Error; err != nil {
		return nil, err
	}
	return token, nil
}

func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	dbConn := r.dm.DB(ctx)
	var rt domain.RefreshToken
	err := dbConn.Where("token = ?", token).First(&rt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &rt, nil
}

func (r *refreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Model(&domain.RefreshToken{}).
		Where("id = ?", id).
		Update("revoked", true).Error
}

func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Model(&domain.RefreshToken{}).
		Where("user_id = ? AND revoked = false", userID).
		Update("revoked", true).Error
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	dbConn := r.dm.DB(ctx)
	return dbConn.Where("expires_at < ?", before).Delete(&domain.RefreshToken{}).Error
}