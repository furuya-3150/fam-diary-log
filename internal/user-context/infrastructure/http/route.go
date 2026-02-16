package http

import (
	"log/slog"
	"net/http"

	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/broker"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/config"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/controller"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/http/handler"
	jwtgen "github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/jwt"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/oauth"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/infrastructure/repository"
	"github.com/furuya-3150/fam-diary-log/internal/user-context/usecase"
	"github.com/furuya-3150/fam-diary-log/pkg/clock"
	"github.com/furuya-3150/fam-diary-log/pkg/db"
	middAuth "github.com/furuya-3150/fam-diary-log/pkg/middleware/auth"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

func NewRouter() *echo.Echo {
	cfg := config.Load()

	dbManager := db.NewDBManager(cfg.DB.DatabaseURL)
	txManager := db.NewTransaction(dbManager)

	// Family
	familyRepo := repository.NewFamilyRepository(dbManager)
	familyInvitationRepo := repository.NewFamilyInvitationRepository(dbManager)
	familyMemberRepo := repository.NewFamilyMemberRepository(dbManager)
	// token generator (infra) for signing JWTs
	tokenGenerator := jwtgen.NewTokenGenerator(&clock.Real{})

	// OAuth providers - Google only
	googleProvider := oauth.NewGoogleProviderWithOAuth2(
		cfg.OAuth.Google.ClientID,
		cfg.OAuth.Google.ClientSecret,
		cfg.OAuth.Google.RedirectURL,
	)

	// Auth
	authRepo := repository.NewUserRepository(dbManager)
	authUsecase := usecase.NewAuthUsecase(authRepo, familyMemberRepo, googleProvider, tokenGenerator)
	authController := controller.NewAuthController(authUsecase)
	authHandler := handler.NewAuthHandler(authController)

	// User
	userUsecase := usecase.NewUserUsecase(authRepo, txManager)
	userController := controller.NewUserController(userUsecase)
	userHandler := handler.NewUserHandler(userController)

	// mail broker publisher
	pub := broker.NewDiaryMailerPublisher(slog.Default())

	familyUsecase := usecase.NewFamilyUsecase(familyRepo, familyMemberRepo, familyInvitationRepo, authRepo, txManager, &clock.Real{}, tokenGenerator, pub)
	familyController := controller.NewFamilyController(familyUsecase)
	familyHandler := handler.NewFamilyHandler(familyController, familyUsecase)

	// Notification settings
	notificationRepo := repository.NewNotificationSettingRepository(dbManager)
	notificationUsecase := usecase.NewNotificationUsecase(notificationRepo)
	notificationHandler := handler.NewNotificationHandler(notificationUsecase)

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
	users := e.Group("/users")
	users.Use(middAuth.JWTAuthMiddleware(cfg.JWT.Secret), middAuth.RequireFamily())
	users.PUT("/me", userHandler.EditProfile)
	users.GET("/me", userHandler.GetProfile)

	// Family routes
	families := e.Group("/families")
	families.POST("", familyHandler.CreateFamily, middAuth.JWTAuthMiddleware(cfg.JWT.Secret))
	families.POST("/me/invitations", familyHandler.InviteMembers, middAuth.JWTAuthMiddleware(cfg.JWT.Secret), middAuth.RequireFamily())
	families.POST("/join-requests", familyHandler.ApplyToFamily, middAuth.JWTAuthMiddleware(cfg.JWT.Secret))

	// Notification settings routes
	notifications := e.Group("/families/me/settings")
	notifications.Use(middAuth.JWTAuthMiddleware(cfg.JWT.Secret), middAuth.RequireFamily())
	notifications.PUT("/notifications", notificationHandler.UpdateNotificationSetting)
	notifications.GET("/notifications", notificationHandler.GetNotificationSetting)

	return e
}
