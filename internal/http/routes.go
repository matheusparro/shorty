// internal/http/routes.go
package http

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matheusparro/shorty/internal/cache"
	"github.com/matheusparro/shorty/internal/config"
	"github.com/matheusparro/shorty/internal/handler"
	"github.com/matheusparro/shorty/internal/http/middleware"
	"github.com/matheusparro/shorty/internal/repository/postgres"
	"github.com/matheusparro/shorty/internal/service"
)

func RegisterRoutes(
	app *fiber.App,
	db *pgxpool.Pool,
	redisClient *cache.RedisClient,
	cfg *config.Config,
	publisher service.EventPublisher, // ✅ injetado pelo main
) {
	// HEALTH
	app.Get("/health", func(c *fiber.Ctx) error {
		status := fiber.Map{"status": "ok", "database": "connected"}

		if redisClient != nil {
			status["cache"] = "connected"
		} else {
			status["cache"] = "unavailable"
		}

		status["kafka_enabled"] = cfg.KafkaEnabled
		status["kafka_brokers"] = cfg.KafkaBrokers

		return c.JSON(status)
	})

	// AUTH
	userRepo := postgres.NewUserPG(db)
	refreshRepo := postgres.NewRefreshTokenPG(db)

	accessTTL := 15 * time.Minute
	refreshTTL := 15 * 24 * time.Hour

	authSvc := service.NewAuthService(userRepo, refreshRepo, cfg.JWTSecret, accessTTL, refreshTTL)
	authHandler := handler.NewAuthHandler(authSvc)

	// SHORT URL
	shortURLRepo := postgres.NewShortURLPG(db)
	shortURLSvc := service.NewShortURLService(shortURLRepo, redisClient, publisher) // ✅ publisher aqui
	shortURLHandler := handler.NewShortURLHandler(shortURLSvc, cfg.BaseURL)

	// ROUTES
	api := app.Group("/api")
	v1 := api.Group("/v1")

	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// redirect público
	app.Get("/:code", shortURLHandler.Redirect)

	// protegidas
	protected := v1.Group("/", middleware.JWTAuth(cfg.JWTSecret))
	protected.Get("/me", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		email, _ := c.Locals("email").(string)
		return c.JSON(fiber.Map{"userId": userID, "email": email})
	})

	protected.Post("/shorturls", shortURLHandler.Create)
}
