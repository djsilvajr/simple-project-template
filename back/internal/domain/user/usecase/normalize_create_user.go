package usecase

import (
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// NormalizeCreateUser limpa os campos do cadastro e aplica o papel padrão.
func NormalizeCreateUser(input model.CreateUserInput) model.CreateUserInput {
	input.Login = strings.TrimSpace(input.Login)
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Role == "" {
		input.Role = model.RoleUser
	}
	return input
}
