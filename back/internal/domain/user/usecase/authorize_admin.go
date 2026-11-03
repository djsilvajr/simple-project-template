package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// AuthorizeAdmin permite a operação apenas para administradores.
func AuthorizeAdmin(actor model.Actor) error {
	if !actor.IsAdmin() {
		return usererrors.ErrForbidden
	}
	return nil
}
