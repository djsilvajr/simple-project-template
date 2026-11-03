package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// EnsureNotSelf impede que o ator remova o próprio usuário.
func EnsureNotSelf(actor model.Actor, userID uint) error {
	if actor.ID == userID {
		return usererrors.ErrCannotDeleteSelf
	}
	return nil
}
