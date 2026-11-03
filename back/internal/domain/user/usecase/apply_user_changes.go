package usecase

import (
	"strings"

	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// ApplyUserChanges copia para o usuário os campos informados, já normalizados.
// A senha fica de fora porque precisa ser validada antes de virar hash.
func ApplyUserChanges(user *model.User, input model.UpdateUserInput) {
	if input.Login != nil {
		user.Login = strings.TrimSpace(*input.Login)
	}
	if input.Name != nil {
		user.Name = strings.TrimSpace(*input.Name)
	}
	if input.Email != nil {
		user.Email = strings.ToLower(strings.TrimSpace(*input.Email))
	}
	if input.Role != nil {
		user.Role = *input.Role
	}
}
