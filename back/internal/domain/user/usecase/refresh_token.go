package usecase

import (
	"context"
	"errors"
	"time"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
)

type RefreshToken struct {
	service service.Service
	tokens  *token.Manager
	grace   time.Duration
}

func NewRefreshToken(svc service.Service, tokens *token.Manager, grace time.Duration) *RefreshToken {
	return &RefreshToken{service: svc, tokens: tokens, grace: grace}
}

// Execute troca um token válido, ou expirado há no máximo `grace`, por um
// novo. O usuário é recarregado do banco: se foi removido a renovação falha,
// e se o papel mudou o novo token já sai com o papel atualizado.
func (uc *RefreshToken) Execute(ctx context.Context, raw string) (*model.LoginOutput, error) {
	claims, err := uc.tokens.ParseForRefresh(raw, uc.grace)
	if err != nil {
		return nil, usererrors.ErrSessionExpired
	}

	user, err := uc.service.GetByID(ctx, claims.UserID)
	if errors.Is(err, usererrors.ErrNotFound) {
		return nil, usererrors.ErrSessionExpired
	}
	if err != nil {
		return nil, err
	}

	return issueToken(uc.tokens, user)
}
