package usecase

import (
	"context"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// EnsureUserUnique garante que login e email não pertencem a outro usuário.
// exceptID ignora o próprio usuário em uma atualização (0 no cadastro).
func EnsureUserUnique(ctx context.Context, repo repository.Repository, login, email string, exceptID uint) error {
	exists, err := repo.ExistsByLogin(ctx, login, exceptID)
	if err != nil {
		return err
	}
	if exists {
		return usererrors.ErrLoginTaken
	}

	exists, err = repo.ExistsByEmail(ctx, email, exceptID)
	if err != nil {
		return err
	}
	if exists {
		return usererrors.ErrEmailTaken
	}

	return nil
}
