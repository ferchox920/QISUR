package crypto

import (
	"context"
	"errors"
	"time"

	"catalog-api/internal/identity"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// BcryptHasher implementa identity.PasswordHasher con bcrypt.
type BcryptHasher struct {
	Cost int
}

func (h BcryptHasher) Hash(password string) (string, error) {
	cost := h.Cost
	if cost == 0 {
		cost = bcrypt.DefaultCost
	}
	out, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(out), err
}

func (h BcryptHasher) Compare(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// JWTProvider emite tokens JWT; el secreto/issuer/ttl viene por config.
type JWTProvider struct {
	Secret string
	Issuer string
	TTL    time.Duration
}

// AuthClaims extiende los claims estandar con metadata de rol.
type AuthClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func (p JWTProvider) Generate(ctx context.Context, user identity.User) (string, error) {
	if p.Secret == "" {
		return "", errors.New("jwt secret not configured")
	}

	claims := AuthClaims{
		Role: string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			Issuer:    p.Issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.TTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(p.Secret))
}

// Validate parsea y valida un token JWT y retorna sus claims.
func (p JWTProvider) Validate(token string) (AuthClaims, error) {
	var claims AuthClaims
	if p.Secret == "" {
		return claims, errors.New("jwt secret not configured")
	}
	parsed, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(p.Secret), nil
	})
	if err != nil {
		return claims, err
	}
	if !parsed.Valid {
		return claims, errors.New("invalid token")
	}
	return claims, nil
}
