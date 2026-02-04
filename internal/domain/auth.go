package domain

import "time"

// TokenPair representa o resultado de autenticação (login/register/refresh)
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	AccessExp    time.Time
	RefreshExp   time.Time
}

// AuthenticatedUser é um "view" do domínio útil pra auth (sem expor hash)
type AuthenticatedUser struct {
	UserID string
	Email  Email
}
