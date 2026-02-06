// internal/worker/worker.go
package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matheusparro/shorty/internal/queue"
	"github.com/matheusparro/shorty/internal/repository/postgres"
	"github.com/matheusparro/shorty/internal/service"
)

type Deps struct {
	DB *pgxpool.Pool
}

var deps Deps

// InitDeps é chamado pelo main no modo worker
func InitDeps(d Deps) {
	deps = d
}

func Run(ctx context.Context, kafkaClient *queue.KafkaClient) error {
	if deps.DB == nil {
		return fmt.Errorf("worker deps not initialized (DB is nil)")
	}

	groupID := getEnv("WORKER_GROUP_ID", "shorty-workers")
	topics := getEnvSlice("WORKER_TOPICS", []string{
		queue.TopicURLClicks,
	})

	c, err := queue.NewConsumer(kafkaClient, groupID)
	if err != nil {
		return fmt.Errorf("new consumer: %w", err)
	}
	defer c.Close()

	shortURLRepo := postgres.NewShortURLPG(deps.DB)
	shortURLSvc := service.NewShortURLService(shortURLRepo, nil, nil)

	setShortURLService(shortURLSvc)

	applyRegistrations(c)

	log.Printf("📡 worker running group=%s topics=%v", groupID, topics)
	return c.Start(ctx, topics)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if v := os.Getenv(key); v != "" {
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return fallback
}
