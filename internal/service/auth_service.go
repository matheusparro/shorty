package service

import (
	"context"
	"errors"

	"github.com/matheusparro/shorty/internal/domain"
	"github.com/matheusparro/shorty/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users repository.UserRepository
}

func NewAuthService(users repository.UserRepository) *AuthService {
	return &AuthService{users: users}
}

type RegisterResult struct {
	UserID string
	Email  string
}

type LoginResult struct {
	UserID string
	Email  string
}

func (s *AuthService) Register(ctx context.Context, emailRaw, password string) (*RegisterResult, error) {
	// regra de domínio: senha mínima (plain)
	if !domain.IsValidPlainPassword(password) {
		return nil, domain.ErrInvalidPassword
	}

	// domínio: email validado e normalizado
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		return nil, err
	}

	// infra/security: gera hash seguro
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	// domínio: cria entidade user (já validada)
	user, err := domain.NewUser(email, string(hash))
	if err != nil {
		return nil, err
	}

	// persistência: create no DB (tabela users, unique(email))
	if err := s.users.Create(ctx, user); err != nil {
		// se seu repo mapear unique violation -> domain.ErrEmailAlreadyUsed, melhor ainda
		if errors.Is(err, domain.ErrEmailAlreadyUsed) {
			return nil, domain.ErrEmailAlreadyUsed
		}
		return nil, err
	}

	return &RegisterResult{
		UserID: user.ID,
		Email:  user.Email.String(),
	}, nil
}

func (s *AuthService) Login(ctx context.Context, emailRaw, password string) (*LoginResult, error) {
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		// não vaza detalhe (padrão mercado)
		return nil, domain.ErrInvalidCredentials
	}

	user, err := s.users.FindByEmail(ctx, email.String())
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return &LoginResult{
		UserID: user.ID,
		Email:  user.Email.String(),
	}, nil
}
