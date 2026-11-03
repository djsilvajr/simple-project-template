package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// ListUsers busca a página de usuários e o total de registros.
func ListUsers(ctx context.Context, repo repository.Repository, input model.ListUsersInput) ([]model.User, int64, error) {
	return repo.List(ctx, (input.Page-1)*input.PerPage, input.PerPage)
}
