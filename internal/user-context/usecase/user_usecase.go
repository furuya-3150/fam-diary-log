package usecase

import (
	"context"
	"time"

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

// FamilyMemberInfo represents a family member with user info and role (DTO)
type FamilyMemberInfo struct {
	ID         uuid.UUID           `json:"id,omitempty"`
	Email      string              `json:"email,omitempty"`
	Name       string              `json:"name,omitempty"`
	Provider   domain.AuthProvider `json:"provider,omitempty"`
	ProviderID string              `json:"provider_id,omitempty"`
	CreatedAt  time.Time           `json:"created_at,omitempty"`
	UpdatedAt  time.Time           `json:"updated_at,omitempty"`
	Role       string              `json:"role,omitempty"`
}

type UserUsecase interface {
	EditUser(ctx context.Context, input *EditUserInput) (*domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetFamilyMembers(ctx context.Context, familyID uuid.UUID, fields []string) ([]*FamilyMemberInfo, error)
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
func (u *userUsecase) GetFamilyMembers(ctx context.Context, familyID uuid.UUID, fields []string) ([]*FamilyMemberInfo, error) {
	// 許可されたフィールドのホワイトリスト（テーブル別）
	allowedFields := map[string]map[string]bool{
		"users": {
			"id":          true,
			"email":       true,
			"name":        true,
			"provider":    true,
			"provider_id": true,
			"created_at":  true,
			"updated_at":  true,
		},
		"family_members": {
			"role": true,
		},
	}

	// デフォルト値の設定
	defaultUserFields := []string{"id", "email", "name", "provider", "provider_id", "created_at", "updated_at"}
	defaultFamilyMemberFields := []string{"role"}

	validatedUserFields := []string{}
	validatedFamilyMemberFields := []string{}

	if len(fields) == 0 {
		// フィールド指定なしの場合はデフォルトを使用
		validatedUserFields = defaultUserFields
		validatedFamilyMemberFields = defaultFamilyMemberFields
	} else {
		// バリデーション：許可されていないフィールドを除去
		for _, field := range fields {
			if allowedFields["users"][field] {
				validatedUserFields = append(validatedUserFields, field)
			} else if allowedFields["family_members"][field] {
				validatedFamilyMemberFields = append(validatedFamilyMemberFields, field)
			}
		}
		// すべてのフィールドが無効だった場合はデフォルトを使用
		if len(validatedUserFields) == 0 && len(validatedFamilyMemberFields) == 0 {
			validatedUserFields = defaultUserFields
			validatedFamilyMemberFields = defaultFamilyMemberFields
		}
	}

	usersWithRoles, err := u.repo.GetUsersByFamilyID(ctx, familyID, validatedUserFields, validatedFamilyMemberFields)
	if err != nil {
		return nil, &pkgerrors.InternalError{Message: "failed to get family members"}
	}

	// repository層のUserWithRoleをusecase層のFamilyMemberInfoに変換
	members := make([]*FamilyMemberInfo, len(usersWithRoles))
	userFieldsMap := make(map[string]bool)
	familyMemberFieldsMap := make(map[string]bool)
	for _, f := range validatedUserFields {
		userFieldsMap[f] = true
	}
	for _, f := range validatedFamilyMemberFields {
		familyMemberFieldsMap[f] = true
	}

	for i, uwr := range usersWithRoles {
		member := &FamilyMemberInfo{}
		if uwr.User != nil {
			if userFieldsMap["id"] {
				member.ID = uwr.User.ID
			}
			if userFieldsMap["email"] {
				member.Email = uwr.User.Email
			}
			if userFieldsMap["name"] {
				member.Name = uwr.User.Name
			}
			if userFieldsMap["provider"] {
				member.Provider = uwr.User.Provider
			}
			if userFieldsMap["provider_id"] {
				member.ProviderID = uwr.User.ProviderID
			}
			if userFieldsMap["created_at"] {
				member.CreatedAt = uwr.User.CreatedAt
			}
			if userFieldsMap["updated_at"] {
				member.UpdatedAt = uwr.User.UpdatedAt
			}
		}
		if familyMemberFieldsMap["role"] {
			member.Role = uwr.Role.StringJP()
		}
		members[i] = member
	}

	return members, nil
}
