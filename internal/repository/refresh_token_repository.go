package repository

import (
	"context"
	"time"
)

type RefreshTokenRepository interface {
	Save(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error
	FindUserIDByHash(ctx context.Context, tokenHash string) (string, error)
	Revoke(ctx context.Context, tokenHash string) error
}
