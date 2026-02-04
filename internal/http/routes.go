package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matheusparro/shorty/internal/config"
	"github.com/matheusparro/shorty/internal/handler"
	"github.com/matheusparro/shorty/internal/http/middleware"
	"github.com/matheusparro/shorty/internal/repository/postgres"
	"github.com/matheusparro/shorty/internal/service"
)

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool, cfg *config.Config) {
	// health
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// -----------------------
	// AUTH (já existia)
	// -----------------------
	userRepo := postgres.NewUserPG(db)
	refreshRepo := postgres.NewRefreshTokenPG(db)

	accessTTL := 15 * time.Minute
	refreshTTL := 15 * 24 * time.Hour

	authSvc := service.NewAuthService(
		userRepo,
		refreshRepo,
		cfg.JWTSecret,
		accessTTL,
		refreshTTL,
	)
	authHandler := handler.NewAuthHandler(authSvc)

	// -----------------------
	// SHORT URL (NEW)
	// -----------------------
	shortURLRepo := postgres.NewShortURLPG(db)                // NEW
	shortURLSvc := service.NewShortURLService(shortURLRepo)  // NEW
	shortURLHandler := handler.NewShortURLHandler(shortURLSvc, cfg.BaseURL) // NEW

	// api v1
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// auth
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// redirect público (NEW)
	v1.Get("/r/:code", shortURLHandler.Redirect) // NEW

	// rotas protegidas
	protected := v1.Group("/", middleware.JWTAuth(cfg.JWTSecret))

	protected.Get("/me", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		email, _ := c.Locals("email").(string)
		return c.JSON(fiber.Map{"userId": userID, "email": email})
	})

	// criar short url (NEW)
	protected.Post("/shorturls", shortURLHandler.Create) // NEW
}
