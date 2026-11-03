package usecase

import (
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/infra/token"
)

// IssueToken gera um novo JWT para o usuário.
func IssueToken(tokens *token.Manager, user *model.User) (*model.LoginOutput, error) {
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
