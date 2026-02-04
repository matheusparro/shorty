package input

import "time"

type CreateShortURLInput struct {
	OriginalURL string
	UserID      string
	ExpiresAt   *time.Time
}