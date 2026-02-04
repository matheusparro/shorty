package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matheusparro/shorty/internal/domain"
)

type RefreshTokenPG struct {
	db *pgxpool.Pool
}

func NewRefreshTokenPG(db *pgxpool.Pool) *RefreshTokenPG {
	return &RefreshTokenPG{db: db}
}

func (r *RefreshTokenPG) Save(ctx context.Context, userID string, tokenHash string, expiresAt time.Time) error {
	const q = `
		insert into refresh_tokens (user_id, token_hash, expires_at)
		values ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, q, userID, tokenHash, expiresAt.UTC())
	return err
}

func (r *RefreshTokenPG) FindUserIDByHash(ctx context.Context, tokenHash string) (string, error) {
	const q = `
		select user_id
		from refresh_tokens
		where token_hash = $1
		  and revoked_at is null
		  and expires_at > now()
	`

	var userID string
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(&userID)

	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrInvalidToken
	}
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *RefreshTokenPG) Revoke(ctx context.Context, tokenHash string) error {
	const q = `
		update refresh_tokens
		set revoked_at = now()
		where token_hash = $1
		  and revoked_at is null
	`
	_, err := r.db.Exec(ctx, q, tokenHash)
	return err
}
