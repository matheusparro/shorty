package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matheusparro/shorty/internal/domain"
)

type UserPG struct {
	db *pgxpool.Pool
}

func NewUserPG(db *pgxpool.Pool) *UserPG {
	return &UserPG{db: db}
}

func (r *UserPG) Create(ctx context.Context, user *domain.User) error {
	// Timestamps ficam no DB; mas se você quiser, pode setar no domain também.
	const q = `
		insert into users (email, password_hash)
		values ($1, $2)
		returning id, created_at, updated_at
	`

	var id string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRow(ctx, q, user.Email.String(), user.PasswordHash).
		Scan(&id, &createdAt, &updatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyUsed
		}
		return err
	}

	user.ID = id
	user.CreatedAt = createdAt
	user.UpdatedAt = updatedAt
	return nil
}

func (r *UserPG) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
		select id, email, password_hash, created_at, updated_at
		from users
		where email = $1
	`

	var (
		id           string
		emailStr     string
		passwordHash string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRow(ctx, q, email).
		Scan(&id, &emailStr, &passwordHash, &createdAt, &updatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	em, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		ID:           id,
		Email:        em,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}
	return u, nil
}

func (r *UserPG) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
		select id, email, password_hash, created_at, updated_at
		from users
		where id = $1
	`

	var (
		userID       string
		emailStr     string
		passwordHash string
		createdAt    time.Time
		updatedAt    time.Time
	)

	err := r.db.QueryRow(ctx, q, id).
		Scan(&userID, &emailStr, &passwordHash, &createdAt, &updatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	em, err := domain.NewEmail(emailStr)
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           userID,
		Email:        em,
		PasswordHash: passwordHash,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
