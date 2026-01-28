package usecase

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/google/uuid"
)

type FamilyUsecase interface {
	CreateFamily(ctx context.Context, name string, userID uuid.UUID) (*domain.Family, error)
}

type familyUsecase struct {
	fr  repository.FamilyRepository
	fmr repository.FamilyMemberRepository
	tm  db.TransactionManager
}

func NewFamilyUsecase(fr repository.FamilyRepository, fmr repository.FamilyMemberRepository, tm db.TransactionManager) FamilyUsecase {
	return &familyUsecase{
		fr:  fr,
		fmr: fmr,
		tm:  tm,
	}
}

func (u *familyUsecase) CreateFamily(ctx context.Context, name string, userID uuid.UUID) (*domain.Family, error) {
	already, err := u.fmr.IsUserAlreadyMember(ctx, userID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, &errors.ValidationError{Message: "you are already a member of a family"}
	}

	ctx, err = u.tm.BeginTx(ctx)
	family := &domain.Family{
		Name: name,
	}
	family, err = u.fr.CreateFamily(ctx, family)
	if err != nil {
		u.tm.RollbackTx(ctx)
		return nil, err
	}
	member := &domain.FamilyMember{
		FamilyID: family.ID,
		UserID:   userID,
		Role:     domain.RoleAdmin,
	}
	if err := u.fmr.AddFamilyMember(ctx, member); err != nil {
		u.tm.RollbackTx(ctx)
		return nil, err
	}
	u.tm.CommitTx(ctx)

	return family, nil
}
