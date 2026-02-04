package service

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/matheusparro/shorty/internal/domain"
	"github.com/matheusparro/shorty/internal/repository"
	"github.com/matheusparro/shorty/internal/security/jwt"
	"github.com/matheusparro/shorty/internal/security/refresh"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      repository.UserRepository
	refresh    repository.RefreshTokenRepository
	jwtSecret  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(
	users repository.UserRepository,
	refresh repository.RefreshTokenRepository,
	jwtSecret string,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		users:      users,
		refresh:    refresh,
		jwtSecret:  jwtSecret,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}


type RegisterResult struct {
	UserID      string
	Email       string
	AccessToken string
	AccessExp   time.Time
}

type LoginResult struct {
	UserID       string
	Email        string
	Role         string
	AccessToken  string
	AccessExp    time.Time
	RefreshToken string
	RefreshExp   time.Time
}


func (s *AuthService) Register(ctx context.Context, emailRaw, password string) (*RegisterResult, error) {
	// regra de domínio: senha mínima (plain)
			log.Println("REGISTER OFICIAL")
	if !domain.IsValidPlainPassword(password) {
		log.Println("CHEGUEI NO erro 5")

		return nil, domain.ErrInvalidPassword
	}

	// domínio: email validado e normalizado
	email, err := domain.NewEmail(emailRaw)
	if err != nil {
		log.Println("CHEGUEI NO erro 6")

		return nil, err
	}

	// infra/security: gera hash seguro
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		log.Println("CHEGUEI NO erro 7")

		return nil, err
	}

	// domínio: cria entidade user (já validada)
	user, err := domain.NewUser(email, string(hash))
	if err != nil {
		log.Println("CHEGUEI NO erro 8")

		return nil, err
	}

	// persistência: create no DB (tabela users, unique(email))
	log.Println(user)
	if err := s.users.Create(ctx, user); err != nil {
		// se seu repo mapear unique violation -> domain.ErrEmailAlreadyUsed, melhor ainda
		if errors.Is(err, domain.ErrEmailAlreadyUsed) {
		log.Println("CHEGUEI NO erro 9")

			return nil, domain.ErrEmailAlreadyUsed
		}
		log.Println("CHEGUEI NO erro 10")
		return nil, err
	}

	token, exp, err := jwt.SignAccessToken(s.jwtSecret, user.ID, user.Role, user.Email.String(), s.accessTTL)
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:      user.ID,
		Email:       user.Email.String(),
		AccessToken: token,
		AccessExp:   exp,
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

	// 1) access token (JWT curto)
	accessToken, accessExp, err := jwt.SignAccessToken(
		s.jwtSecret,
		user.ID,
		user.Email.String(),
		user.Role,
		s.accessTTL,
	)
	if err != nil {
		return nil, err
	}

	// 2) refresh token (longo) + salvar HASH no banco
	refreshRaw, err := refresh.NewToken()
	if err != nil {
		return nil, err
	}

	refreshHash := refresh.Hash(refreshRaw, s.jwtSecret)
	refreshExp := time.Now().UTC().Add(s.refreshTTL)

	if err := s.refresh.Save(ctx, user.ID, refreshHash, refreshExp); err != nil {
		return nil, err
	}

	return &LoginResult{
		UserID:       user.ID,
		Email:        user.Email.String(),
		Role:         user.Role,
		AccessToken:  accessToken,
		AccessExp:    accessExp,
		RefreshToken: refreshRaw,  // handler pode colocar em cookie HttpOnly
		RefreshExp:   refreshExp,
	}, nil
}

