package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// ValidateCreateUser valida todos os campos do cadastro.
func ValidateCreateUser(input model.CreateUserInput) error {
	vErr := usererrors.NewValidationError()
	validateLogin(vErr, input.Login)
	validateName(vErr, input.Name)
	validateEmail(vErr, input.Email)
	validatePassword(vErr, input.Password)
	validateRole(vErr, input.Role)
	return vErr.OrNil()
}
