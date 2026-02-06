// internal/service/shorturl_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/matheusparro/shorty/internal/cache"
	"github.com/matheusparro/shorty/internal/domain"
	events "github.com/matheusparro/shorty/internal/events"
	"github.com/matheusparro/shorty/internal/queue"
	"github.com/matheusparro/shorty/internal/repository"
	"github.com/matheusparro/shorty/internal/service/input"
)

var ErrInvalidURL = errors.New("invalid url")

// EventPublisher é a porta (interface) — infra (Kafka) implementa isso.
type EventPublisher interface {
	PublishEvent(topic string, key string, event any) error
}

type ShortURLService struct {
	ShortURLRepository repository.ShortURLRepository
	RedisClient        *cache.RedisClient
	Publisher          EventPublisher
}

func NewShortURLService(
	shortURLRepo repository.ShortURLRepository,
	redisClient *cache.RedisClient,
	publisher EventPublisher,
) *ShortURLService {
	return &ShortURLService{
		ShortURLRepository: shortURLRepo,
		RedisClient:        redisClient,
		Publisher:          publisher,
	}
}

func (s *ShortURLService) CreateShortURL(ctx context.Context, in input.CreateShortURLInput) (*domain.ShortURL, error) {
	if strings.TrimSpace(in.OriginalURL) == "" {
		return nil, ErrInvalidURL
	}
	if strings.TrimSpace(in.UserID) == "" {
		return nil, errors.New("user id is required")
	}

	shortCode := generateShortCode(7)

	entity, err := domain.NewShortURL(in.OriginalURL, shortCode, in.UserID, in.ExpiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.ShortURLRepository.Create(ctx, entity); err != nil {
		return nil, err
	}

	// ✅ cache opcional e seguro (não assume ExpiresAt)
	if s.RedisClient != nil {
		if entity.ExpiresAt != nil {
			ttl := time.Until(*entity.ExpiresAt)
			if ttl > 0 {
				_ = s.RedisClient.Client().SetEx(ctx, shortCode, entity.OriginalURL, ttl).Err()
			}
		} else {
			_ = s.RedisClient.Client().Set(ctx, shortCode, entity.OriginalURL, 0).Err()
		}
	}

	// (Opcional futuro): publicar URLCreatedEvent aqui também.
	return entity, nil
}

func generateShortCode(n int) string {
	b := make([]byte, n+4)
	_, _ = rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)
	return code[:n]
}

func (s *ShortURLService) FindActiveForRedirect(ctx context.Context, in input.RedirectInput) (string, error) {
	shortCode := in.Code

	// 1) tenta cache
	if s.RedisClient != nil {
		ctxRedis, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		url, err := s.RedisClient.Client().Get(ctxRedis, shortCode).Result()
		if err == nil {
			log.Println("Cache hit for short code:", shortCode)

			// best-effort: publica click sem travar redirect
			s.publishClickBestEffort(in)

			return url, nil
		}
	}

	// 2) fallback banco
	url, err := s.ShortURLRepository.FindActiveURLForRedirect(ctx, shortCode)
	if err != nil {
		return "", err
	}

	s.publishClickBestEffort(in)
	return url, nil
}

func (s *ShortURLService) IncrementVisitByShortCode(ctx context.Context, shortCode string) error {
	 return s.ShortURLRepository.IncrementVisit(ctx, shortCode)
}

func (s *ShortURLService) publishClickBestEffort(in input.RedirectInput) {
	if s.Publisher == nil {
		return
	}

	ev := events.ClickEvent{
		EventMeta: events.EventMeta{
			EventID:    uuid.NewString(),
			Version:    1,
			OccurredAt: time.Now().UTC(),
			Source:     "shorty-api",
		},
		ShortCode: in.Code,
		IP:        in.IP,
		UserAgent: in.UserAgent,
		Referer:   in.Referer,
	}

	go func() {
		if err := s.Publisher.PublishEvent(queue.TopicURLClicks, in.Code, ev); err != nil {
			log.Printf("⚠️ failed to publish click event: %v", err)
		}
	}()
}

