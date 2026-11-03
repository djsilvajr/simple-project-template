package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// AuthorizeRoleChange impede que quem não é admin altere um papel.
func AuthorizeRoleChange(actor model.Actor, role *model.Role) error {
	if role != nil && !actor.IsAdmin() {
		return usererrors.ErrForbidden
	}
	return nil
}
