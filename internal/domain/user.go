package domain

import "time"

type User struct {
	ID           string
	Email        Email
	PasswordHash string
	Role         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(email Email, passwordHash string) (*User, error) {
	if len(passwordHash) < 20 {
		return nil, ErrInvalidPassword
	}

	now := time.Now().UTC()

	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "user", // ✅ default
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) Touch() {
	u.UpdatedAt = time.Now().UTC()
}
