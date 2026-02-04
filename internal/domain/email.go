package domain

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Email string

func NewEmail(v string) (Email, error) {
	v = strings.TrimSpace(strings.ToLower(v))
	if len(v) < 3 || len(v) > 254 || !emailRegex.MatchString(v) {
		return "", ErrInvalidEmail
	}
	return Email(v), nil
}

func (e Email) String() string {
	return string(e)
}
