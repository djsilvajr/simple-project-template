package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// AuthorizeUserAccess permite o acesso ao admin ou ao próprio usuário.
func AuthorizeUserAccess(actor model.Actor, userID uint) error {
	if !actor.CanAccess(userID) {
		return usererrors.ErrForbidden
	}
	return nil
}
