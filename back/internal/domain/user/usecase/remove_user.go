package usecase

import (
	"context"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// RemoveUser faz o soft delete do usuário.
func RemoveUser(ctx context.Context, repo repository.Repository, id uint) error {
	return repo.Delete(ctx, id)
}
