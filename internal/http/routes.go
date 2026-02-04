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

	userRepo := postgres.NewUserPG(db)
	refreshRepo := postgres.NewRefreshTokenPG(db)

	accessTTL := 15 * time.Minute
	refreshTTL := 15 * 24 * time.Hour // 15 dias (exemplo)

	authSvc := service.NewAuthService(
		userRepo,
		refreshRepo,
		cfg.JWTSecret,
		accessTTL,
		refreshTTL,
	)
	authHandler := handler.NewAuthHandler(authSvc)

	// api v1
	api := app.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	protected := v1.Group("/", middleware.JWTAuth(cfg.JWTSecret))

	protected.Get("/me", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		email, _ := c.Locals("email").(string)
		return c.JSON(fiber.Map{"userId": userID, "email": email})
	})
}
