package usecase

import (
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// NormalizeLoginInput remove os espaços ao redor do login.
func NormalizeLoginInput(input model.LoginInput) model.LoginInput {
	input.Login = strings.TrimSpace(input.Login)
	return input
}
