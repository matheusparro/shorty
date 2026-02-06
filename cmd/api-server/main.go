// cmd/api-server/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/matheusparro/shorty/internal/cache"
	"github.com/matheusparro/shorty/internal/config"
	"github.com/matheusparro/shorty/internal/db"
	httpserver "github.com/matheusparro/shorty/internal/http"
	"github.com/matheusparro/shorty/internal/queue"
	"github.com/matheusparro/shorty/internal/service"
	"github.com/matheusparro/shorty/internal/worker"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	appMode := strings.ToLower(strings.TrimSpace(os.Getenv("APP_MODE")))
	if appMode == "" {
		appMode = "api"
	}
	log.Println("🚀 starting shorty mode:", appMode)

	// ctx de vida do processo (cancela no CTRL+C)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// DB
	dbCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	pool, err := db.NewPool(dbCtx, cfg.PostgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()
	log.Println("✅ postgres connected")

	// Redis (opcional)
	redisClient, err := cache.NewRedisClient(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("⚠️ redis unavailable: %v", err)
		redisClient = nil
	} else {
		log.Println("✅ redis connected")
		defer redisClient.Close()
	}

	// Kafka (opcional)
	var kafkaClient *queue.KafkaClient
	var publisher service.EventPublisher

	if cfg.KafkaEnabled && len(cfg.KafkaBrokers) > 0 {
		kc, err := queue.NewKafkaClient(cfg.KafkaBrokers)
		if err != nil {
			log.Printf("⚠️ kafka disabled: %v", err)
		} else {
			kafkaClient = kc
			log.Printf("✅ kafka client ready brokers=%v", cfg.KafkaBrokers)

			// Producer é útil no modo API (publicar clicks)
			producer, err := queue.NewProducer(kafkaClient)
			if err != nil {
				log.Printf("⚠️ kafka producer disabled: %v", err)
			} else {
				publisher = producer
				log.Println("✅ kafka producer enabled")
				defer producer.Close()
			}
		}
	} else {
		log.Printf("⚠️ kafka disabled (KafkaEnabled=%v brokers=%v)", cfg.KafkaEnabled, cfg.KafkaBrokers)
	}

	// -----------------------
	// MODE: API
	// -----------------------
	if appMode == "api" {
		app := httpserver.NewApp()
		httpserver.RegisterRoutes(app, pool, redisClient, &cfg, publisher)

		go func() {
			log.Printf("🌐 API running on :%s", cfg.AppPort)
			if err := app.Listen(":" + cfg.AppPort); err != nil {
				log.Println("http stopped:", err)
			}
		}()

		<-ctx.Done()
		log.Println("🛑 shutting down api...")
		_ = app.Shutdown()
		return
	}

	// -----------------------
	// MODE: WORKER
	// -----------------------
	if appMode == "worker" {
		if kafkaClient == nil {
			log.Fatal("kafka not available; worker cannot start")
		}

		worker.InitDeps(worker.Deps{DB: pool})

		go func() {
			log.Println("📡 worker started")
			if err := worker.Run(ctx, kafkaClient); err != nil {
				log.Println("worker stopped:", err)
				stop()
			}
		}()

		<-ctx.Done()
		log.Println("🛑 shutting down worker...")
		return
	}

	log.Fatalf("invalid APP_MODE=%s (use api|worker)", appMode)
}
