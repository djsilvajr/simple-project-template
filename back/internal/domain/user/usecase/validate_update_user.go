package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// ValidateUpdateUser valida apenas os campos informados na atualização. Os
// valores são lidos do usuário, depois de passarem por ApplyUserChanges.
func ValidateUpdateUser(user *model.User, input model.UpdateUserInput) error {
	vErr := usererrors.NewValidationError()
	if input.Login != nil {
		validateLogin(vErr, user.Login)
	}
	if input.Name != nil {
		validateName(vErr, user.Name)
	}
	if input.Email != nil {
		validateEmail(vErr, user.Email)
	}
	if input.Role != nil {
		validateRole(vErr, user.Role)
	}
	if input.Password != nil {
		validatePassword(vErr, *input.Password)
	}
	return vErr.OrNil()
}
