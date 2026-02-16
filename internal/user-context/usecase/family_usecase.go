package usecase

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	jwtgen "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/jwt"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/pkg/broker/publisher"
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
	CreateFamily(ctx context.Context, name string, userID uuid.UUID) (string, error)
	InviteMembers(ctx context.Context, in InviteMembersInput) error
	ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) (string, error)
}

type familyUsecase struct {
	fr  repository.FamilyRepository
	fmr repository.FamilyMemberRepository
	fiR repository.FamilyInvitationRepository
	ur  repository.UserRepository
	tm  db.TransactionManager
	clk clock.Clock
	tg  jwtgen.TokenGenerator
	mp  publisher.Publisher
}

func NewFamilyUsecase(
	fr repository.FamilyRepository,
	fmr repository.FamilyMemberRepository,
	fiR repository.FamilyInvitationRepository,
	ur repository.UserRepository,
	tm db.TransactionManager,
	clk clock.Clock,
	tg jwtgen.TokenGenerator,
	mp publisher.Publisher,
) FamilyUsecase {
	return &familyUsecase{
		fr:  fr,
		fmr: fmr,
		fiR: fiR,
		ur:  ur,
		tm:  tm,
		clk: clk,
		tg:  tg,
		mp:  mp,
	}
}

func (u *familyUsecase) CreateFamily(ctx context.Context, name string, userID uuid.UUID) (string, error) {
	already, err := u.fmr.IsUserAlreadyMember(ctx, userID)
	if err != nil {
		return "", err
	}
	if already {
		return "", &errors.ValidationError{Message: "you are already a member of a family"}
	}

	ctx, err = u.tm.BeginTx(ctx)
	if err != nil {
		return "", err
	}
	family := &domain.Family{
		Name: name,
	}
	family, err = u.fr.CreateFamily(ctx, family)
	if err != nil {
		u.tm.RollbackTx(ctx)
		return "", err
	}
	member := &domain.FamilyMember{
		FamilyID: family.ID,
		UserID:   userID,
		Role:     domain.RoleAdmin,
	}
	if err := u.fmr.AddFamilyMember(ctx, member); err != nil {
		u.tm.RollbackTx(ctx)
		return "", err
	}
	u.tm.CommitTx(ctx)
	signed, err := u.tg.GenerateToken(ctx, userID, family.ID, domain.RoleAdmin)
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (fu *familyUsecase) InviteMembers(ctx context.Context, input InviteMembersInput) error {
	slog.Debug("InviteMembers: start", "family_id", input.FamilyID, "inviter_user_id", input.InviterUserID, "emails", input.Emails)

	targetDate := fu.clk.Now()
	// 有効期限は7日後
	expiresAt := targetDate.Add(7 * 24 * time.Hour)
	token, err := random.GenerateRandomBase64String(32)
	existing, err := fu.fiR.FindInvitationByFamilyID(ctx, input.FamilyID)
	slog.Debug("InviteMembers: existing invitation fetched", "existing", existing, "error", err)
	if err != nil {
		return err
	}
	if existing != nil {
		// EmailsをJSONにシリアライズ
		emailsJSON, err := json.Marshal(input.Emails)
		if err != nil {
			return err
		}
		err = fu.fiR.UpdateInvitationTokenAndExpires(ctx, input.FamilyID, input.InviterUserID, token, expiresAt, emailsJSON)
		if err != nil {
			return err
		}
	} else {
		slog.Debug("InviteMembers: creating new invitation", "token", token, "expires_at", expiresAt)
		inv := &domain.FamilyInvitation{
			FamilyID:        input.FamilyID,
			InviterUserID:   input.InviterUserID,
			InvitationToken: token,
			InvitedEmails:   input.Emails,
			ExpiresAt:       expiresAt,
		}
		if err := fu.fiR.CreateInvitation(ctx, inv); err != nil {
			return err
		}
	}
	defer fu.mp.Close()

	inviter, err := fu.ur.GetUserByID(ctx, input.InviterUserID)
	if err != nil {
		return err
	}
	family, err := fu.fr.GetFamilyByID(ctx, input.FamilyID)
	if err != nil {
		return err
	}

	event := &domain.MailSendEvent{
		TemplateID: "family_invite_v1",
		To:         input.Emails,
		Locale:     "ja",
		Payload: map[string]interface{}{
			"inviter_name": inviter.Name,
			"family_name":  family.Name,
			"app_url":      config.Cfg.App.URL + "/auth/google?token=" + token,
		},
	}

	if err := fu.mp.Publish(ctx, event); err != nil {
		return err
	}

	slog.Debug("Family invitation emails published", "to", input.Emails, "family_id", input.FamilyID, "inviter_user_id", input.InviterUserID)

	return nil
}

func (fu *familyUsecase) ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) (string, error) {
	// トークンから招待を取得
	inv, err := fu.fiR.FindInvitationByToken(ctx, token)
	if err != nil {
		return "", err
	}
	if inv == nil {
		return "", &errors.BadRequestError{Message: "invalid invitation token"}
	}
	now := fu.clk.Now()
	if now.After(inv.ExpiresAt) {
		return "", &errors.ValidationError{Message: "invitation expired"}
	}

	// ユーザーがすでにメンバーか確認
	already, err := fu.fmr.IsUserAlreadyMember(ctx, userID)
	if err != nil {
		return "", err
	}
	if already {
		return "", &errors.ValidationError{Message: "you are already a member of a family"}
	}

	// ユーザー情報を取得してメールアドレスをチェック
	user, err := fu.ur.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", &errors.BadRequestError{Message: "invalid user"}
	}

	// メールアドレスが招待リストに含まれているかチェック
	if !fu.isEmailInvited(user.Email, inv.InvitedEmails) {
		slog.Warn("ApplyToFamily: user email not in invited list", "user_id", userID, "email", user.Email, "family_id", inv.FamilyID)
		return "", &errors.ValidationError{Message: "you are not invited to this family"}
	}

	// 招待されているユーザーは直接メンバーに追加
	slog.Debug("ApplyToFamily: adding user to family", "user_id", userID, "email", user.Email, "family_id", inv.FamilyID)

	member := &domain.FamilyMember{
		FamilyID: inv.FamilyID,
		UserID:   userID,
		Role:     domain.RoleMember,
	}
	if err := fu.fmr.AddFamilyMember(ctx, member); err != nil {
		return "", err
	}

	slog.Debug("ApplyToFamily: user successfully added to family", "user_id", userID, "family_id", inv.FamilyID)

	signed, err := fu.tg.GenerateToken(ctx, userID, inv.FamilyID, domain.RoleAdmin)
	if err != nil {
		return "", err
	}
	return signed, nil
}

// isEmailInvited checks if the given email is in the invited emails list
func (fu *familyUsecase) isEmailInvited(email string, invitedEmails []string) bool {
	for _, invitedEmail := range invitedEmails {
		if email == invitedEmail {
			return true
		}
	}
	return false
}
