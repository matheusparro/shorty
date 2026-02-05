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

func RegisterRoutes(app *fiber.App, db *pgxpool.Pool, redisClient *cache.RedisClient, cfg *config.Config) {
	// health (já checa DB, agora checa Redis também)
	app.Get("/health", func(c *fiber.Ctx) error {
		status := fiber.Map{"status": "ok", "database": "connected"}
		
		if redisClient != nil {
			status["cache"] = "connected"
		} else {
			status["cache"] = "unavailable"
		}
		
		return c.JSON(status)
	})

	// -----------------------
	// AUTH
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
	// SHORT URL
	// -----------------------
	shortURLRepo := postgres.NewShortURLPG(db)
	
	// Passa o Redis para o serviço (pode ser nil)
	shortURLSvc := service.NewShortURLService(shortURLRepo, redisClient)
	
	shortURLHandler := handler.NewShortURLHandler(shortURLSvc, cfg.BaseURL)

	// api v1
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// auth
	auth := v1.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// redirect público
	v1.Get("/r/:code", shortURLHandler.Redirect)

	// rotas protegidas
	protected := v1.Group("/", middleware.JWTAuth(cfg.JWTSecret))

	protected.Get("/me", func(c *fiber.Ctx) error {
		userID, _ := c.Locals("userID").(string)
		email, _ := c.Locals("email").(string)
		return c.JSON(fiber.Map{"userId": userID, "email": email})
	})

	// criar short url
	protected.Post("/shorturls", shortURLHandler.Create)
}