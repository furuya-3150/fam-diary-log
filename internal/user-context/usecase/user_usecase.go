package usecase

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
)

// EditUserInput is the input model for editing a user (usecase層専用)
type EditUserInput struct {
	ID    string
	Name  string
	Email string
}

type UserUsecase interface {
	EditUser(ctx context.Context, input *EditUserInput) (*domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetFamilyMembers(ctx context.Context, familyID uuid.UUID, fields []string) ([]*domain.User, error)
}

type userUsecase struct {
	repo repository.UserRepository
	tm   db.TransactionManager
}

func NewUserUsecase(repo repository.UserRepository, tm db.TransactionManager) UserUsecase {
	return &userUsecase{
		repo: repo,
		tm:   tm,
	}
}

func (u *userUsecase) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to get user"}
	}
	if user == nil {
		return nil, &pkgerrors.ValidationError{Message: "user not found"}
	}
	return user, nil
}

func (u *userUsecase) EditUser(ctx context.Context, input *EditUserInput) (*domain.User, error) {
	userID, err := uuid.Parse(input.ID)
	if err != nil {
		return nil, &pkgerrors.ValidationError{Message: "invalid user id"}
	}
	user, err := u.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to get user"}
	}
	if user == nil {
		return nil, &pkgerrors.ValidationError{Message: "user not found"}
	}

	user.Name = input.Name
	user.Email = input.Email
	updated, err := u.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to update user"}
	}
	return updated, nil
}

// GetFamilyMembers gets all users in a family with optional field selection
func (u *userUsecase) GetFamilyMembers(ctx context.Context, familyID uuid.UUID, fields []string) ([]*domain.User, error) {
	// 許可されたフィールドのホワイトリスト
	allowedFields := map[string]bool{
		"id":          true,
		"email":       true,
		"name":        true,
		"provider":    true,
		"provider_id": true,
		"created_at":  true,
		"updated_at":  true,
	}

	// デフォルト値の設定（fieldsが空の場合は全フィールド）
	var validatedFields []string
	if len(fields) == 0 {
		validatedFields = []string{"id", "email", "name", "provider", "provider_id", "created_at", "updated_at"}
	} else {
		// バリデーション：許可されていないフィールドを除去
		for _, field := range fields {
			if allowedFields[field] {
				validatedFields = append(validatedFields, field)
			}
		}
		// すべてのフィールドが無効だった場合はデフォルトを使用
		if len(validatedFields) == 0 {
			validatedFields = []string{"id", "email", "name", "provider", "provider_id", "created_at", "updated_at"}
		}
	}

	users, err := u.repo.GetUsersByFamilyID(ctx, familyID, validatedFields)
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to get family members"}
	}
	return users, nil
}
