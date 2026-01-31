package jwtgen

import (
	"context"
	"testing"
	"time"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/domain"
	cfg "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestGenerateToken_ProducesValidJWT(t *testing.T) {
    // set predictable config
    cfg.Cfg.JWT.Secret = "test-secret"
    cfg.Cfg.JWT.ExpiresIn = time.Duration(3600) * time.Second
    cfg.Cfg.JWT.Issuer = "test-issuer"

    fixed := &clock.Fixed{Time: time.Unix(1000, 0)}
    tg := NewTokenGenerator(fixed)

    userID := uuid.New()
    familyID := uuid.New()

    tokenStr, expires, err := tg.GenerateToken(context.Background(), userID, familyID, domain.RoleMember)
    require.NoError(t, err)
    require.Equal(t, int64(3600), expires)

    // parse and validate signature
    parsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
        return []byte(cfg.Cfg.JWT.Secret), nil
    })
    require.NoError(t, err)
    require.True(t, parsed.Valid)

    claims := parsed.Claims.(jwt.MapClaims)
    require.Equal(t, userID.String(), claims["sub"])
    require.Equal(t, familyID.String(), claims["family_id"])
    require.Equal(t, cfg.Cfg.JWT.Issuer, claims["iss"])
    // iat and exp are numeric
    require.Equal(t, float64(fixed.Time.Unix()), claims["iat"].(float64))
}
