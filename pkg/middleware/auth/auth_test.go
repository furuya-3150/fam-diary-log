package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-key"

func createTestJWT(userID, familyID *uuid.UUID, role string, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      userID.String(),
		"role":     role,
		"email":    "test@example.com",
		"provider": "test",
		"iat":      now.Unix(),
		"exp":      now.Add(expiry).Unix(),
		"iss":      "test-issuer",
	}
	if familyID != nil {
		claims["family_id"] = familyID.String()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestJWTAuthMiddleware_ValidToken(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenString, err := createTestJWT(&userID, &familyID, "admin", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Test handler that verifies context values
	handler := func(c echo.Context) error {
		ctx := c.Request().Context()

		extractedUserID, ok := GetUserIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, userID, extractedUserID)

		extractedFamilyID, ok := GetFamilyIDFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, familyID, extractedFamilyID)

		extractedRole, ok := GetRoleFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, RoleAdmin, extractedRole)

		return c.String(http.StatusOK, "success")
	}

	middleware := JWTAuthMiddleware(testSecret)
	h := middleware(handler)
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTAuthMiddleware_MissingCookie_Required(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middleware := JWTAuthMiddleware(testSecret)
	h := middleware(handler)
	err := h(c)
	log.Println("err:", err)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp, "message")
	assert.NotEmpty(t, resp["message"])
}

func TestJWTAuthMiddleware_InvalidToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: "invalid.token.here",
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middleware := JWTAuthMiddleware(testSecret)
	h := middleware(handler)
	err := h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp, "message")
	assert.NotEmpty(t, resp["message"])
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	// Create token that expired 1 hour ago
	tokenString, err := createTestJWT(&userID, &familyID, "member", testSecret, -time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	middleware := JWTAuthMiddleware(testSecret)
	h := middleware(handler)
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp, "message")
	assert.NotEmpty(t, resp["message"])
}

func TestJWTAuthMiddleware_MemberRole(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenString, err := createTestJWT(&userID, &familyID, "member", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		ctx := c.Request().Context()
		role, ok := GetRoleFromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, RoleMember, role)
		return c.String(http.StatusOK, "success")
	}

	middleware := JWTAuthMiddleware(testSecret)
	h := middleware(handler)
	err = h(c)

	assert.NoError(t, err)
}

func TestRequireAuth_WithAuth(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenString, err := createTestJWT(&userID, &familyID, "admin", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Apply both middlewares
	authMW := JWTAuthMiddleware(testSecret)
	requireMW := RequireAuth()
	h := authMW(requireMW(handler))
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_WithoutAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Apply both middlewares (optional auth + require)
	authMW := JWTAuthMiddleware(testSecret)
	requireMW := RequireAuth()
	h := authMW(requireMW(handler))
	err := h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// レスポンスボディのJSONを確認
	var resp map[string]interface{}
	jsonErr := json.Unmarshal(rec.Body.Bytes(), &resp)
	assert.NoError(t, jsonErr)
	assert.Contains(t, resp, "message")
	assert.NotEmpty(t, resp["message"])
}

func TestRequireRole_AdminRole(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenString, err := createTestJWT(&userID, &familyID, "admin", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_token",
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	authMW := JWTAuthMiddleware(testSecret)
	roleMW := RequireRole(RoleAdmin)
	h := authMW(roleMW(handler))
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestParseRole(t *testing.T) {
	assert.Equal(t, RoleAdmin, ParseRole("admin"))
	assert.Equal(t, RoleMember, ParseRole("member"))
	assert.Equal(t, RoleUnknown, ParseRole("invalid"))
	assert.Equal(t, RoleUnknown, ParseRole(""))
}

func TestRoleString(t *testing.T) {
	assert.Equal(t, "admin", RoleAdmin.String())
	assert.Equal(t, "member", RoleMember.String())
	assert.Equal(t, "unknown", RoleUnknown.String())
}

func TestRequireFamily_WithFamily(t *testing.T) {
	userID := uuid.New()
	familyID := uuid.New()
	tokenString, err := createTestJWT(&userID, &familyID, "admin", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Apply both middlewares
	authMW := JWTAuthMiddleware(testSecret)
	familyMW := RequireFamily()
	h := authMW(familyMW(handler))
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireFamily_WithoutFamily(t *testing.T) {
	userID := uuid.New()
	// Create token with nil family_id
	tokenString, err := createTestJWT(&userID, nil, "admin", testSecret, time.Hour)
	require.NoError(t, err)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{
		Name:  AuthCookieName,
		Value: tokenString,
	})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	}

	// Apply both middlewares
	authMW := JWTAuthMiddleware(testSecret)
	familyMW := RequireFamily()
	h := authMW(familyMW(handler))
	err = h(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
