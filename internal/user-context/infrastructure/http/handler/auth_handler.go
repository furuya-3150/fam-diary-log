package handler

import (
	"log/slog"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/pkg/errors"
	pkgerrors "github.com/furuya-3150/fam-diary-log/pkg/errors"
	"github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler interface {
	// OAuth2 server-side flow handlers
	InitiateGoogleLogin(c echo.Context) error
	GoogleCallback(c echo.Context) error
}

type authHandler struct {
	authController controller.AuthController
}

func NewAuthHandler(authController controller.AuthController) AuthHandler {
	return &authHandler{
		authController: authController,
	}
}

// InitiateGoogleLogin redirects user to Google's OAuth consent page
func (h *authHandler) InitiateGoogleLogin(c echo.Context) error {
	// Generate state and get OAuth URL from usecase
	authURL, state, err := h.authController.InitiateGoogleLogin()
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	// Save state in session for later verification
	sess, err := session.Get(config.SessionName, c)
	if err != nil {
		return errors.RespondWithError(c, &pkgerrors.InternalError{Message: "failed to get session"})
	}
	sess.Values[config.SessionKeyOAuthState] = state
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		return errors.RespondWithError(c, &pkgerrors.InternalError{Message: "failed to save session"})
	}

	// Redirect to Google's OAuth page
	return c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// GoogleCallback handles the OAuth callback from Google
func (h *authHandler) GoogleCallback(c echo.Context) error {
	code := c.QueryParam("code")
	if code == "" {
		return errors.RespondWithError(c, &pkgerrors.ValidationError{Message: "authorization code is required"})
	}

	// Get state from URL
	receivedState := c.QueryParam("state")
	if receivedState == "" {
		return errors.RespondWithError(c, &pkgerrors.ValidationError{Message: "state parameter is required"})
	}

	// Get state from session
	sess, err := session.Get(config.SessionName, c)
	if err != nil {
		return errors.RespondWithError(c, &pkgerrors.InternalError{Message: "failed to get session"})
	}

	savedState, ok := sess.Values[config.SessionKeyOAuthState].(string)
	if !ok || savedState == "" {
		return errors.RespondWithError(c, &pkgerrors.ValidationError{Message: "invalid session state"})
	}

	// Verify state (CSRF protection)
	if receivedState != savedState {
		return errors.RespondWithError(c, &pkgerrors.ValidationError{Message: "invalid state parameter - possible CSRF attack"})
	}

	// Clear state from session (one-time use)
	delete(sess.Values, config.SessionKeyOAuthState)
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		slog.Error("failed to clear OAuth state from session", "Error", err)
	}

	_, token, err := h.authController.HandleGoogleCallback(c.Request().Context(), code)
	if err != nil {
		return errors.RespondWithError(c, err)
	}

	// Set access token in HTTPOnly Cookie
	accessTokenCookie := &http.Cookie{
		Name:     auth.AuthCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,                          // JavaScriptからアクセス不可（XSS対策）
		Secure:   true,                          // HTTPS通信のみ
		SameSite: http.SameSiteStrictMode,       // CSRF対策
		MaxAge:   int(config.Cfg.JWT.ExpiresIn), // トークンの有効期限
	}
	c.SetCookie(accessTokenCookie)

	// セッションを削除
	sess.Options.MaxAge = -1
	if err := sess.Save(c.Request(), c.Response()); err != nil {
		slog.Error("failed to delete session", "Error", err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, config.Cfg.ClientApp.URL)
}
