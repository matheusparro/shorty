package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/matheusparro/shorty/internal/domain"
)
var ErrNotFound = errors.New("short url not found")

type ShortURLPG struct {
	db *pgxpool.Pool
}
func NewShortURLPG(db *pgxpool.Pool) *ShortURLPG {
	return &ShortURLPG{db: db}
}
func (r *ShortURLPG) Create(ctx context.Context, shortURL *domain.ShortURL) error {
	const q = `
		insert into short_urls (short_code, original_url, user_id, expires_at, is_active, visit_count)
		values ($1, $2, $3, $4, $5, $6)
		returning id, created_at, updated_at
	`

	var (
		id        string
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRow(ctx, q,
		shortURL.ShortCode,
		shortURL.OriginalURL,
		shortURL.UserID,
		shortURL.ExpiresAt,  // nil -> vira NULL no Postgres (pgx lida de boa)
		shortURL.IsActive,
		shortURL.VisitCount,
	).Scan(&id, &createdAt, &updatedAt)
	if err != nil {
		return err
	}

	shortURL.ID = id
	shortURL.CreatedAt = createdAt
	shortURL.UpdatedAt = updatedAt
	return nil
}

func (r *ShortURLPG) FindByShortCode(ctx context.Context, shortCode string) (*domain.ShortURL, error) {
	const q = `
		select id, short_code, original_url, user_id, created_at, updated_at, expires_at, visit_count, is_active
		from short_urls
		where short_code = $1
	`

	var su domain.ShortURL

	err := r.db.QueryRow(ctx, q, shortCode).Scan(
		&su.ID,
		&su.ShortCode,
		&su.OriginalURL,
		&su.UserID,
		&su.CreatedAt,
		&su.UpdatedAt,
		&su.ExpiresAt,   // *time.Time recebe NULL -> nil automaticamente
		&su.VisitCount,
		&su.IsActive,
	)
	if err != nil {
		return nil, err
	}

	return &su, nil
}

func (r *ShortURLPG) FindActiveURLForRedirect(ctx context.Context, shortCode string) (string, error) {
	const q = `
		SELECT original_url
		FROM short_urls
		WHERE short_code = $1
		  AND is_active = true
		  AND (expires_at IS NULL OR expires_at > now())
		LIMIT 1;
	`

	var originalURL string
	err := r.db.QueryRow(ctx, q, shortCode).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}

	return originalURL, nil
}

func (r *ShortURLPG) Inactivate(ctx context.Context, shortCode, userID string) error {
	const q = `
		update short_urls
		set is_active = false, updated_at = now()
		where short_code = $1
		  and user_id = $2
	`
	_, err := r.db.Exec(ctx, q, shortCode, userID)
	return err
}

func (r *ShortURLPG) IncrementVisit(ctx context.Context, shortCode string) error {
	const q = `
		update short_urls
		set visit_count = visit_count + 1, updated_at = now()
		where short_code = $1
	`
	_, err := r.db.Exec(ctx, q, shortCode)
	return err
}


