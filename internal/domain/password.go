package domain

func IsValidPlainPassword(p string) bool {
	return len(p) >= 8 && len(p) <= 72
}
