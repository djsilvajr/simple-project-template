package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/service"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
)

type Login struct {
	service service.Service
	tokens  *token.Manager
}

func NewLogin(svc service.Service, tokens *token.Manager) *Login {
	return &Login{service: svc, tokens: tokens}
}

func (uc *Login) Execute(ctx context.Context, input model.LoginInput) (*model.LoginOutput, error) {
	user, err := uc.service.Authenticate(ctx, input.Login, input.Password)
	if err != nil {
		return nil, err
	}

	return issueToken(uc.tokens, user)
}

// issueToken gera um novo JWT para o usuário (usado no login e na renovação).
func issueToken(tokens *token.Manager, user *model.User) (*model.LoginOutput, error) {
	signed, expiresAt, err := tokens.Generate(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &model.LoginOutput{
		Token:     signed,
		ExpiresAt: expiresAt.Unix(),
		User:      user,
	}, nil
}
