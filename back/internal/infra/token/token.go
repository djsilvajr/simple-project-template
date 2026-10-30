package token

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("token inválido ou expirado")

type Claims struct {
	UserID uint   `json:"uid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Manager gera e valida tokens JWT (HS256).
type Manager struct {
	secret     []byte
	expiration time.Duration
	issuer     string
}

func NewManager(secret string, expiration time.Duration, issuer string) *Manager {
	return &Manager{
		secret:     []byte(secret),
		expiration: expiration,
		issuer:     issuer,
	}
}

func (m *Manager) Generate(userID uint, role string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.expiration)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			Issuer:    m.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token: erro ao assinar: %w", err)
	}

	return signed, expiresAt, nil
}

// Parse valida o token (assinatura, emissor e expiração).
func (m *Manager) Parse(raw string) (*Claims, error) {
	return m.parse(raw)
}

// ParseForRefresh valida o token aceitando que ele tenha expirado há no
// máximo `grace`. Usado apenas na renovação (POST /auth/refresh).
func (m *Manager) ParseForRefresh(raw string, grace time.Duration) (*Claims, error) {
	return m.parse(raw, jwt.WithLeeway(grace))
}

func (m *Manager) parse(raw string, extra ...jwt.ParserOption) (*Claims, error) {
	claims := &Claims{}

	opts := append([]jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	}, extra...)

	parsed, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (any, error) {
		return m.secret, nil
	}, opts...)
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
