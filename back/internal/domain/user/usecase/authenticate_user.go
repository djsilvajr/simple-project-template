package usecase

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/repository"
)

// dummyHash é usado apenas para igualar o tempo de resposta quando o login não existe.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// AuthenticateUser confere login e senha. Retorna sempre ErrInvalidCredentials
// (sem distinguir usuário inexistente de senha errada).
func AuthenticateUser(ctx context.Context, repo repository.Repository, secret []byte, input model.LoginInput) (*model.User, error) {
	user, err := repo.FindByLogin(ctx, input.Login)
	if errors.Is(err, usererrors.ErrNotFound) {
		// Executa um bcrypt mesmo assim para não vazar, pelo tempo de
		// resposta, se o login existe.
		_ = bcrypt.CompareHashAndPassword(dummyHash, pepper(input.Password, secret))
		return nil, usererrors.ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), pepper(input.Password, secret)) != nil {
		return nil, usererrors.ErrInvalidCredentials
	}

	return user, nil
}
