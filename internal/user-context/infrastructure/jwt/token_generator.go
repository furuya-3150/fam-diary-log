package jwtgen

import (
	"context"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	cfg "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateToken(ctx context.Context, userID uuid.UUID, familyID uuid.UUID, role domain.Role) (string, error)
}

type tokenGenerator struct {
	clk clock.Clock
}

func NewTokenGenerator(clk clock.Clock) TokenGenerator {
	return &tokenGenerator{clk: clk}
}

func (t *tokenGenerator) GenerateToken(ctx context.Context, userID uuid.UUID, familyID uuid.UUID, role domain.Role) (string, error) {
	c := cfg.Cfg
	now := t.clk.Now()
	expiresAt := now.Add(c.JWT.ExpiresIn)
	var claims jwt.MapClaims
	claims = jwt.MapClaims{
		"sub":       userID.String(),
		"family_id": familyID.String(),
		"iat":       now.Unix(),
		"exp":       expiresAt.Unix(),
		"iss":       c.JWT.Issuer,
	}
	if familyID != uuid.Nil {
		claims["family_id"] = familyID.String()
		claims["role"] = role.String()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(c.JWT.Secret))
	if err != nil {
		return "", err
	}
	return signed, nil
}
