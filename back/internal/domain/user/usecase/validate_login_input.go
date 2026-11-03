package usecase

import (
	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// ValidateLoginInput exige login e senha preenchidos.
func ValidateLoginInput(input model.LoginInput) error {
	vErr := usererrors.NewValidationError()
	if input.Login == "" {
		vErr.Add("login", "obrigatório")
	}
	if input.Password == "" {
		vErr.Add("password", "obrigatório")
	}
	return vErr.OrNil()
}
