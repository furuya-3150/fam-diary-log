package auth

import (
	"context"
	"log"

	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/jwt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Context keys for storing authentication information
type contextKey string

const (
	AuthCookieName = "auth_token"
)

const (
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyFamilyID contextKey = "family_id"
	ContextKeyRole     contextKey = "role"
)

// Role represents user role in a family
type Role int

const (
	RoleUnknown Role = iota
	RoleAdmin
	RoleMember
)

func (r Role) String() string {
	switch r {
	case RoleAdmin:
		return "admin"
	case RoleMember:
		return "member"
	default:
		return "unknown"
	}
}

// ParseRole converts a string to Role
func ParseRole(s string) Role {
	switch s {
	case "admin":
		return RoleAdmin
	case "member":
		return RoleMember
	default:
		return RoleUnknown
	}
}

// JWTAuthMiddleware creates an Echo middleware that extracts JWT from cookie,
// validates it, and sets user_id, family_id, and role in the request context.
//
// Parameters:
//   - jwtSecret: Secret key for validating JWT
func JWTAuthMiddleware(jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Println("JWTAuthMiddleware called")
			// Extract JWT from cookie
			cookie, err := c.Cookie(AuthCookieName)
			if err != nil {
				return errors.RespondWithError(c, &errors.UnauthorizedError{Message: "missing authentication cookie"})
			}

			tokenString := cookie.Value
			if tokenString == "" {
				return errors.RespondWithError(c, &errors.UnauthorizedError{Message: "empty authentication token"})
			}

			// Validate and parse JWT
			claims, err := jwt.ValidateAndGetClaims(tokenString, jwtSecret)
			if err != nil {
				return errors.RespondWithError(c, &errors.UnauthorizedError{Message: "invalid or expired token"})
			}

			// Parse role from claims
			role := ParseRole(claims.Role)

			// Set values in request context
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyFamilyID, claims.FamilyID)
			ctx = context.WithValue(ctx, ContextKeyRole, role)

			// Update request with new context
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
		}
	}
}

// Helper functions to extract values from context

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(ContextKeyUserID).(uuid.UUID)
	return userID, ok
}

// GetFamilyIDFromContext extracts family ID from context
func GetFamilyIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	familyID, ok := ctx.Value(ContextKeyFamilyID).(uuid.UUID)
	return familyID, ok
}

// GetRoleFromContext extracts role from context
func GetRoleFromContext(ctx context.Context) (Role, bool) {
	role, ok := ctx.Value(ContextKeyRole).(Role)
	return role, ok
}

// RequireAuth is a helper middleware that enforces user authentication
// Use this after JWTAuthMiddleware to require user authentication on specific routes
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			_, hasUser := GetUserIDFromContext(ctx)

			if !hasUser {
				return errors.RespondWithError(c, &errors.UnauthorizedError{Message: "authentication required"})
			}

			return next(c)
		}
	}
}

// RequireFamily is a helper middleware that enforces family context
// Use this after JWTAuthMiddleware to require family membership on specific routes
func RequireFamily() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			familyID, hasFamily := GetFamilyIDFromContext(ctx)

			if !hasFamily || familyID == uuid.Nil {
				// return echo.NewHTTPError(http.StatusForbidden, errors.UnauthorizedError{Message: "family membership required"})
				log.Println("RequireFamily: family membership required")
				return errors.RespondWithError(c, &errors.ForbiddenError{Message: "family membership required"})
			}

			return next(c)
		}
	}
}

// RequireRole is a helper middleware that enforces a specific role
func RequireRole(requiredRole Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			role, ok := GetRoleFromContext(ctx)

			if !ok || role != requiredRole {
				return errors.RespondWithError(c, &errors.ForbiddenError{Message: "insufficient permissions"})
			}

			return next(c)
		}
	}
}
