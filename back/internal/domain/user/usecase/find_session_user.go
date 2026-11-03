package usecase

import (
	"context"
	"errors"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// FindSessionUser recarrega o dono da sessão. Se o usuário foi removido, a
// sessão é considerada expirada.
func FindSessionUser(ctx context.Context, repo repository.Repository, id uint) (*model.User, error) {
	user, err := repo.FindByID(ctx, id)
	if errors.Is(err, usererrors.ErrNotFound) {
		return nil, usererrors.ErrSessionExpired
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
