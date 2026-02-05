package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/matheusparro/shorty/internal/cache"
	"github.com/matheusparro/shorty/internal/config"
	"github.com/matheusparro/shorty/internal/db"
	httpserver "github.com/matheusparro/shorty/internal/http"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Conecta no PostgreSQL
	pool, err := db.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	// Conecta no Redis (OPCIONAL - não quebra se falhar)
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("⚠️  Redis unavailable (running without cache): %v", err)
		redisClient = nil // Sistema continua sem cache
	} else {
		log.Println("✅ Redis connected successfully")
		defer redisClient.Close()
	}

	// Inicia o servidor HTTP
	app := httpserver.NewApp()
	httpserver.RegisterRoutes(app, pool, redisClient, &cfg)

	go func() {
		log.Printf("🚀 Server running on port %s", cfg.AppPort)
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Println(err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down gracefully...")
	_ = app.Shutdown()
}