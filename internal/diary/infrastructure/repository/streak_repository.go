package repository

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/diary/domain"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StreakRepository interface {
	CreateOrUpdate(ctx context.Context, streak *domain.Streak) (*domain.Streak, error)
	Get(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error)
}

type streakRepository struct {
	dm *db.DBManager
}

func NewStreakRepository(dm *db.DBManager) StreakRepository {
	return &streakRepository{
		dm: dm,
	}
}

func (sr *streakRepository) CreateOrUpdate(ctx context.Context, streak *domain.Streak) (*domain.Streak, error) {
	db := sr.dm.DB(ctx)

	// UPSERT: ユーザーとファミリーの組み合わせで既存データをチェック
	err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "family_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"current_streak", "last_post_date", "updated_at"}),
	}).Create(streak).Error
	if err != nil {
		return nil, err
	}

	return streak, nil
}

func (sr *streakRepository) Get(ctx context.Context, userID, familyID uuid.UUID) (*domain.Streak, error) {
	db := sr.dm.DB(ctx)
	var streak *domain.Streak

	err := db.Where("user_id = ? AND family_id = ?", userID, familyID).First(&streak).Error
	if err != nil {
		// レコードが見つからない場合は (nil, nil) を返す
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return streak, nil
}
