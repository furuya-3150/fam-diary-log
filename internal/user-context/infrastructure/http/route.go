package http

import (
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/handler"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/oauth"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	cfg := config.Load()

	dbManager := db.NewDBManager(cfg.DB.DatabaseURL)
	txManager := db.NewTransaction(dbManager) // TODO: Use txManager when needed

	// OAuth providers - Google only
	googleProvider := oauth.NewGoogleProviderWithOAuth2(
		cfg.OAuth.Google.ClientID,
		cfg.OAuth.Google.ClientSecret,
		cfg.OAuth.Google.RedirectURL,
	)

	// Auth
	authRepo := repository.NewUserRepository(dbManager)
	authUsecase := usecase.NewAuthUsecase(authRepo, googleProvider)
	authController := controller.NewAuthController(authUsecase)
	authHandler := handler.NewAuthHandler(authController)

	// User
	userUsecase := usecase.NewUserUsecase(authRepo, txManager)
	userController := controller.NewUserController(userUsecase)
	userHandler := handler.NewUserHandler(userController)

	e := echo.New()

	// Session middleware
	store := sessions.NewCookieStore([]byte(cfg.Session.Secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: true,
		Secure:   true, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	e.Use(session.Middleware(store))

	e.GET("healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	// Authentication routes - Google OAuth2 server-side flow only
	auth := e.Group("/auth")
	auth.GET("/google", authHandler.InitiateGoogleLogin)
	auth.GET("/google/callback", authHandler.GoogleCallback)

	// User routes
	e.PUT("/users/me", userHandler.EditProfile)

	return e
}
