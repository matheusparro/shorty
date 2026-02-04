package repository

import (
	"context"

	"github.com/matheusparro/shorty/internal/domain"
)

type ShortURLRepository interface {
	Create(ctx context.Context, shortURL *domain.ShortURL) error
	FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortURL, error)
	FindActiveForRedirect(ctx context.Context, shortCode string) (*domain.ShortURL, error)
	Inactivate(ctx context.Context, shortCode, userID string) error
	IncrementVisit(ctx context.Context, shortCode string) error
}