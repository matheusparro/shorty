package jwt

import (
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string `json:"uid"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwtlib.RegisteredClaims
}

func SignAccessToken(secret string, userID string, email string, role string, ttl time.Duration) (token string, expiresAt time.Time, err error) {
	now := time.Now().UTC()
	expiresAt = now.Add(ttl)

	claims := Claims{
		UserID: userID,
		Role:   role,
		Email:  email,
		RegisteredClaims: jwtlib.RegisteredClaims{
			IssuedAt:  jwtlib.NewNumericDate(now),
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
	}

	t := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	token, err = t.SignedString([]byte(secret))
	return
}

func ParseAccessToken(secret string, tokenStr string) (*Claims, error) {
	token, err := jwtlib.ParseWithClaims(tokenStr, &Claims{}, func(t *jwtlib.Token) (interface{}, error) {
		// garante o algoritmo esperado
		if _, ok := t.Method.(*jwtlib.SigningMethodHMAC); !ok {
			return nil, jwtlib.ErrTokenSignatureInvalid
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, jwtlib.ErrTokenInvalidClaims
	}
	return claims, nil
}

