package domain

import (
	"errors"
	"time"
)

type ShortURL struct {
	ID          string
	OriginalURL string
	ShortCode   string
	UserID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ExpiresAt   *time.Time
	VisitCount  int
	IsActive    bool
}

var ErrInvalidExpiration = errors.New("invalid expiration")
func NewShortURL(originalURL, shortCode, userID string, expiresAt *time.Time) (*ShortURL, error) {
	if originalURL == "" || shortCode == "" || userID == "" {
		return nil, ErrInvalidShortURL
	}

	now := time.Now().UTC()
	if expiresAt != nil && expiresAt.Before(now) {
		return nil, ErrInvalidExpiration
	}

	return &ShortURL{
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		UserID:      userID,
		CreatedAt:   now,
		UpdatedAt:   now,
		ExpiresAt:   expiresAt,
		VisitCount:  0,
		IsActive:    true,
	}, nil
}
