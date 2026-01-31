package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/google/uuid"
)

type FamilyController interface {
	CreateFamily(ctx context.Context, req *dto.CreateFamilyRequest, userID uuid.UUID) (*dto.FamilyResponse, error)
	InviteMembers(ctx context.Context, req *dto.InviteMembersRequest) error
	ApplyToFamily(ctx context.Context, req *dto.ApplyRequest, userID uuid.UUID) error
	RespondToJoinRequest(ctx context.Context, req *dto.RespondJoinRequestRequest, userID uuid.UUID) error
}

type familyController struct {
	fu usecase.FamilyUsecase
}

func NewFamilyController(fu usecase.FamilyUsecase) FamilyController {
	return &familyController{fu: fu}
}

func (c *familyController) CreateFamily(ctx context.Context, req *dto.CreateFamilyRequest, userID uuid.UUID) (*dto.FamilyResponse, error) {
	family, err := c.fu.CreateFamily(ctx, req.Name, userID)
	if err != nil {
		return nil, err
	}
	return &dto.FamilyResponse{
		ID:        family.ID,
		Name:      family.Name,
		CreatedAt: family.CreatedAt,
		UpdatedAt: family.UpdatedAt,
	}, nil
}

func (c *familyController) InviteMembers(ctx context.Context, req *dto.InviteMembersRequest) error {
	in := usecase.InviteMembersInput{
		FamilyID:      req.FamilyID,
		InviterUserID: req.UserID,
		Emails:        req.Emails,
	}
	err := c.fu.InviteMembers(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

func (c *familyController) ApplyToFamily(ctx context.Context, req *dto.ApplyRequest, userID uuid.UUID) error {
	return c.fu.ApplyToFamily(ctx, req.Token, userID)
}

func (c *familyController) RespondToJoinRequest(ctx context.Context, req *dto.RespondJoinRequestRequest, userID uuid.UUID) error {
	// cast status to domain enum
	status := domain.JoinRequestStatus(req.Status)
	return c.fu.RespondToJoinRequest(ctx, req.ID, status, userID)
}
