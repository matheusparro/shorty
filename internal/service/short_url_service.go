// internal/service/shorturl_service.go
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/matheusparro/shorty/internal/domain"
	"github.com/matheusparro/shorty/internal/repository"
	"github.com/matheusparro/shorty/internal/service/input"
)

var ErrInvalidURL = errors.New("invalid url")

type ShortURLService struct {
	ShortURLRepository repository.ShortURLRepository
}

func NewShortURLService(shortURLRepo repository.ShortURLRepository) *ShortURLService {
	return &ShortURLService{ShortURLRepository: shortURLRepo}
}

func (s *ShortURLService) CreateShortURL(ctx context.Context, in input.CreateShortURLInput) (*domain.ShortURL, error) {
	if strings.TrimSpace(in.OriginalURL) == "" {
		return nil, ErrInvalidURL
	}
	if strings.TrimSpace(in.UserID) == "" {
		return nil, errors.New("user id is required")
	}

	shortCode := generateShortCode(7) // 7 chars (ajusta se quiser)

	entity, err := domain.NewShortURL(in.OriginalURL, shortCode, in.UserID, in.ExpiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.ShortURLRepository.Create(ctx, entity); err != nil {
		return nil, err
	}

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
	return s.ShortURLRepository.FindActiveURLForRedirect(ctx, shortCode)
}
