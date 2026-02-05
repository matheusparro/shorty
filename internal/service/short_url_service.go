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

	"github.com/matheusparro/shorty/internal/cache"
	"github.com/matheusparro/shorty/internal/domain"
	"github.com/matheusparro/shorty/internal/repository"
	"github.com/matheusparro/shorty/internal/service/input"
)

var ErrInvalidURL = errors.New("invalid url")

type ShortURLService struct {
	ShortURLRepository repository.ShortURLRepository
	RedisClient        *cache.RedisClient
}

func NewShortURLService(shortURLRepo repository.ShortURLRepository, redisClient *cache.RedisClient) *ShortURLService {
	return &ShortURLService{
		ShortURLRepository: shortURLRepo,
		RedisClient:        redisClient,
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
	
	s.RedisClient.Client().SetEx(ctx, shortCode, entity.OriginalURL, time.Until(*entity.ExpiresAt))

	return entity, nil
}

func generateShortCode(n int) string {
	b := make([]byte, n+4)
	_, _ = rand.Read(b)
	code := base64.RawURLEncoding.EncodeToString(b)
	return code[:n]
}

func (s *ShortURLService) FindActiveForRedirect(
	ctx context.Context,
	shortCode string,
) (string, error) {
	
	if s.RedisClient != nil {
		ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		url, err := s.RedisClient.Client().Get(ctx, shortCode).Result()
		if err == nil {
			log.Default().Println("Cache hit for short code:", shortCode)
			return url, nil
		}
	}

	return s.ShortURLRepository.FindActiveURLForRedirect(ctx, shortCode)
}
