package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// SaveUser grava as alterações de um usuário existente.
func SaveUser(ctx context.Context, repo repository.Repository, user *model.User) error {
	return translateDBError(repo.Update(ctx, user))
}
