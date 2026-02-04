package domain

import "time"

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewRefreshToken(userID, tokenHash string, expiresAt time.Time) (*RefreshToken, error) {
	if userID == "" || tokenHash == "" {
		return nil, ErrInvalidToken
	}
	if time.Now().UTC().After(expiresAt) {
		return nil, ErrTokenExpired
	}

	return &RefreshToken{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt.UTC(),
		CreatedAt: time.Now().UTC(),
	}, nil
}

func (t *RefreshToken) IsExpired(now time.Time) bool {
	return now.UTC().After(t.ExpiresAt)
}

func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t *RefreshToken) Revoke() {
	now := time.Now().UTC()
	t.RevokedAt = &now
}
