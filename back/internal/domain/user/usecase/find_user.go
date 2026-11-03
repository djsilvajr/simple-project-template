package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// FindUser busca um usuário pelo ID.
func FindUser(ctx context.Context, repo repository.Repository, id uint) (*model.User, error) {
	return repo.FindByID(ctx, id)
}
