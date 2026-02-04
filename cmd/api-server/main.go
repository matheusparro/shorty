package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/matheusparro/shorty/internal/config"
	"github.com/matheusparro/shorty/internal/db"
	httpserver "github.com/matheusparro/shorty/internal/http"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	app := httpserver.NewApp()
	httpserver.RegisterRoutes(app, pool, &cfg)

	go func() {
		if err := app.Listen(":" + cfg.AppPort); err != nil {
			log.Println(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	_ = app.Shutdown()
}
