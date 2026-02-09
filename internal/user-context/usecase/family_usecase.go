package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	jwtgen "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/jwt"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/ws"
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
	ApplyToFamily(ctx context.Context, token string, userID uuid.UUID) error
	RespondToJoinRequest(ctx context.Context, requestID uuid.UUID, status domain.JoinRequestStatus, responderUserID uuid.UUID) error
	JoinFamilyIfApproved(ctx context.Context, userID uuid.UUID) (string, error)
}

type familyUsecase struct {
	fr  repository.FamilyRepository
	fmr repository.FamilyMemberRepository
	fiR repository.FamilyInvitationRepository
	fjr repository.FamilyJoinRequestRepository
	ur  repository.UserRepository
	tm  db.TransactionManager
	clk clock.Clock
	tg  jwtgen.TokenGenerator
	pj  ws.Publisher
	mp  publisher.Publisher
}

func NewFamilyUsecase(
	fr repository.FamilyRepository,
	fmr repository.FamilyMemberRepository,
	fiR repository.FamilyInvitationRepository,
	fjr repository.FamilyJoinRequestRepository,
	ur repository.UserRepository,
	tm db.TransactionManager,
	clk clock.Clock,
	tg jwtgen.TokenGenerator,
	pj ws.Publisher,
	mp publisher.Publisher,
) FamilyUsecase {
	return &familyUsecase{
		fr:  fr,
		fmr: fmr,
		fiR: fiR,
		fjr: fjr,
		ur:  ur,
		tm:  tm,
		clk: clk,
		tg:  tg,
		pj:  pj,
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
		err := fu.fiR.UpdateInvitationTokenAndExpires(ctx, input.FamilyID, input.InviterUserID, token, expiresAt)
		if err != nil {
			return err
		}
		// return err
	} else {
		slog.Debug("InviteMembers: creating new invitation", "token", token, "expires_at", expiresAt)
		inv := &domain.FamilyInvitation{
			FamilyID:        input.FamilyID,
			InviterUserID:   input.InviterUserID,
			InvitationToken: token,
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

	slog.Info("Family invitation emails published", "to", input.Emails, "family_id", input.FamilyID, "inviter_user_id", input.InviterUserID)

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

	// 家族の管理者ユーザーを取得
	adminUsers, err := fu.ur.GetAdminUsersByFamilyID(ctx, inv.FamilyID)
	if err != nil {
		return err
	}
	if len(adminUsers) == 0 {
		slog.Warn("ApplyToFamily: no admin users found for family", "family_id", inv.FamilyID)
		return &errors.LogicError{Message: "no admin users found for family"}
	}
	adminMailAddresses := make([]string, 0, len(adminUsers))

	for _, u := range adminUsers {
		adminMailAddresses = append(adminMailAddresses, u.Email)
	}

	applyingUser, err := fu.ur.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}
	if applyingUser == nil {
		return &errors.NotFoundError{"not found applying user"}
	}

	// publish join request notification to mail queue
	defer fu.mp.Close()

	event := &domain.MailSendEvent{
		TemplateID: "family_request_v1",
		Locale:     "ja",
		To:         adminMailAddresses,
		Payload: map[string]interface{}{
			"requester_name": applyingUser.Name,
			"app_url":        config.Cfg.App.URL,
		},
	}

	if err := fu.mp.Publish(ctx, event); err != nil {
		return err
	}

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
	if jr.Status != domain.JoinRequestStatusPending || jr.RespondedAt != (time.Time{}) {
		return &errors.BadRequestError{}
	}
	now := fu.clk.Now()
	updates := map[string]interface{}{
		"status":            int(status),
		"responded_user_id": responderUserID,
		"responded_at":      now,
		"updated_at":        now,
	}
	if err := fu.fjr.UpdateStatusByID(ctx, requestID, updates); err != nil {
		return err
	}

	payload := map[string]interface{}{
		"type":       ws.PayloadTypeJoinRequestResponse,
		"status":     int(status),
		"family_id":  jr.FamilyID,
		"request_id": jr.ID,
	}
	// publish ws notification
	_ = fu.pj.Publish(ctx, jr.UserID, payload)
	// delete ws connections for the user
	fu.pj.CloseUserConnections(jr.UserID)

	return nil
}

func (fu *familyUsecase) JoinFamilyIfApproved(ctx context.Context, userID uuid.UUID) (string, error) {
	// find approved join request for this user
	jr, err := fu.fjr.FindApprovedByUser(ctx, userID)
	if err != nil {
		return "", err
	}
	if jr == nil {
		return "", &errors.NotFoundError{Message: "approved join request not found"}
	}

	// check already member
	already, err := fu.fmr.IsUserAlreadyMember(ctx, userID)
	if err != nil {
		return "", err
	}
	if already {
		return "", &errors.ValidationError{Message: "you are already a member of a family"}
	}

	member := &domain.FamilyMember{
		FamilyID: jr.FamilyID,
		UserID:   userID,
		Role:     domain.RoleMember,
	}
	if err := fu.fmr.AddFamilyMember(ctx, member); err != nil {
		return "", err
	}

	signed, err := fu.tg.GenerateToken(ctx, userID, jr.FamilyID, domain.RoleMember)
	if err != nil {
		return "", err
	}
	return signed, nil
}
