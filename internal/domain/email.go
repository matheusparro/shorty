package domain

import (
	"log"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Email string

func NewEmail(v string) (Email, error) {
	log.Println("Validating email:", v)
	v = strings.TrimSpace(strings.ToLower(v))
	if len(v) < 3 || len(v) > 254 || !emailRegex.MatchString(v) {
			log.Println("Invalid email format:", v)
		return "", ErrInvalidEmail
	}
	log.Println("Valid email:", v)
	return Email(v), nil
}

func (e Email) String() string {
	return string(e)
}
