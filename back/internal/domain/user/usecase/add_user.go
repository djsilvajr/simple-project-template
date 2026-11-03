package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// AddUser grava um novo usuário.
func AddUser(ctx context.Context, repo repository.Repository, user *model.User) error {
	return translateDBError(repo.Create(ctx, user))
}
