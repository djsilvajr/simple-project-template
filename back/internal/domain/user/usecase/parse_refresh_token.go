package usecase

import (
	"time"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
)

// ParseRefreshToken aceita um token válido ou expirado há no máximo grace.
func ParseRefreshToken(tokens *token.Manager, raw string, grace time.Duration) (*token.Claims, error) {
	claims, err := tokens.ParseForRefresh(raw, grace)
	if err != nil {
		return nil, usererrors.ErrSessionExpired
	}
	return claims, nil
}
