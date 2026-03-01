package domain

import "regexp"

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

func IsValidEmail(email string) bool {
	if len(email) < 3 || len(email) > 254 {
		return false
	}
	return emailRegex.MatchString(email)
}

func IsValidPlainPassword(password string) bool {
	return len(password) >= 8 && len(password) <= 72
}
