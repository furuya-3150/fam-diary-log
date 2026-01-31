package usecase

import (
	"context"
	"regexp"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/random"
	"github.com/google/uuid"
)

type InviteMembersInput struct {
	FamilyID      uuid.UUID
	InviterUserID uuid.UUID
	Emails        []string
}

type FamilyUsecase interface {
	CreateFamily(ctx context.Context, name string, userID uuid.UUID) (*domain.Family, error)
	InviteMembers(ctx context.Context, in InviteMembersInput) error
	ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) error
	RespondToJoinRequest(ctx context.Context, requestID uuid.UUID, status domain.JoinRequestStatus, responderUserID uuid.UUID) error
}

type familyUsecase struct {
	fr  repository.FamilyRepository
	fmr repository.FamilyMemberRepository
	fiR repository.FamilyInvitationRepository
	fjr repository.FamilyJoinRequestRepository
	tm  db.TransactionManager
	clk clock.Clock
}

func NewFamilyUsecase(
	fr repository.FamilyRepository,
	fmr repository.FamilyMemberRepository,
	fiR repository.FamilyInvitationRepository,
	fjr repository.FamilyJoinRequestRepository,
	tm db.TransactionManager,
	clk clock.Clock) FamilyUsecase {
	return &familyUsecase{
		fr:  fr,
		fmr: fmr,
		fiR: fiR,
		fjr: fjr,
		tm:  tm,
		clk: clk,
	}
}

// メール形式バリデーション（簡易）
func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$`)
	return re.MatchString(email)
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

func (fu *familyUsecase) InviteMembers(ctx context.Context, input InviteMembersInput) error {
	targetDate := fu.clk.Now()
	// 有効期限は7日後
	expiresAt := targetDate.Add(7 * 24 * time.Hour)
	token, err := random.GenerateRandomBase64String(32)
	existing, err := fu.fiR.FindInvitationByFamilyID(ctx, input.FamilyID)
	if err != nil {
		return err
	}
	if err == nil && existing != nil {
		err := fu.fiR.UpdateInvitationTokenAndExpires(ctx, input.FamilyID, input.InviterUserID, token, expiresAt)
		if err != nil {
			return err
		}
		return nil
	}
	inv := &domain.FamilyInvitation{
		FamilyID:        input.FamilyID,
		InviterUserID:   input.InviterUserID,
		InvitationToken: token,
		ExpiresAt:       expiresAt,
	}
	if err := fu.fiR.CreateInvitation(ctx, inv); err != nil {
		return err
	}
	// TODO: メール送信キューイングへポスト
	// 招待メール送信
	return nil
}

func (fu *familyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) error {
	// トークンから招待を取得
	inv, err := fu.fiR.FindInvitationByToken(ctx, token)
	if err != nil {
		return err
	}
	if inv == nil {
		return &errors.NotFoundError{Message: "invitation not found"}
	}
	now := fu.clk.Now()
	if now.After(inv.ExpiresAt) {
		return &errors.ValidationError{Message: "invitation expired"}
	}

	// ユーザーがすでにメンバーか確認
	already, err := fu.fmr.IsUserAlreadyMember(ctx, userID)
	if err != nil {
		return err
	}
	if already {
		return &errors.ValidationError{Message: "you are already a member of a family"}
	}

	// すでに同一 family_id, user_id で pending の申請があるか確認
	if existing, err := fu.fjr.FindPendingRequest(ctx, inv.FamilyID, userID); err != nil {
		return err
	} else if existing != nil {
		return &errors.ValidationError{Message: "join request already pending"}
	}

	// 申請レコードを作成（status=1: 申請中）
	jr := &domain.FamilyJoinRequest{
		FamilyID: inv.FamilyID,
		UserID:   userID,
		Status:   domain.JoinRequestStatusPending,
	}
	if err := fu.fjr.CreateJoinRequest(ctx, jr); err != nil {
		return err
	}

	// TODO: メール送信キューイングへポスト
	return nil
}

func (fu *familyUsecase) RespondToJoinRequest(ctx context.Context, requestID uuid.UUID, status domain.JoinRequestStatus, responderUserID uuid.UUID) error {
	// join request が存在するか確認
	jr, err := fu.fjr.FindByID(ctx, requestID)
	if err != nil {
		return err
	}
	if jr == nil {
		return &errors.NotFoundError{Message: "join request not found"}
	}
	if jr.Status != domain.JoinRequestStatusPending || jr.RespondedAt != nil {
		return &errors.BadRequestError{}
	}
	now := fu.clk.Now()
	updates := map[string]interface{}{
		"status":           int(status),
		"responded_user_id": responderUserID,
		"responded_at":     now,
		"updated_at":       now,
	}
	ctx, err = fu.tm.BeginTx(ctx)
	if err != nil {
		return err
	}
	if err := fu.fjr.UpdateStatusByID(ctx, requestID, updates); err != nil {
		if err := fu.tm.RollbackTx(ctx); err != nil {
			return err
		}
		return err
	}
	if status == domain.JoinRequestStatusApproved {
		// 承認の場合、family member レコードを追加
		member := &domain.FamilyMember{
			FamilyID: jr.FamilyID,
			UserID:   jr.UserID,
			Role:     domain.RoleMember,
		}
		if err := fu.fmr.AddFamilyMember(ctx, member); err != nil {
			if err := fu.tm.RollbackTx(ctx); err != nil {
				return err
			}
			return err
		}
	}
	if err := fu.tm.CommitTx(ctx); err != nil {
		return err
	}
	// TODO: メール送信キューイングへポスト
	
	return nil
}
