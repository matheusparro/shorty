package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyUsed   = errors.New("email already in use")
	ErrUserNotFound       = errors.New("user not found")

	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrTokenRevoked = errors.New("token revoked")
	ErrInvalidShortURL = errors.New("invalid short URL")
	ErrShortURLNotFound = errors.New("short URL not found")
	ErrInvalidClientName = errors.New("invalid client name")
	ErrInvalidClientCity = errors.New("invalid client city")
	ErrInvalidClientAddress = errors.New("invalid client address")
	ErrInvalidClientPhone = errors.New("invalid client phone")

)
