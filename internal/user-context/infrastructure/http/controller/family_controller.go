package controller

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller/dto"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/google/uuid"
)

type FamilyController interface {
	CreateFamily(ctx context.Context, req *dto.CreateFamilyRequest, userID uuid.UUID) (*dto.FamilyResponse, error)
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
