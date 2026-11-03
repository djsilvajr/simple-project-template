package usecase

import (
	"net/mail"
	"regexp"
	"unicode/utf8"

	usererrors "github.com/djsilvajr/simple-project-template/back/internal/domain/user/errors"
	"github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"
)

// Regras de cada campo, compartilhadas pelos use cases de validação.

const (
	minPasswordLength = 8
	maxPasswordLength = 128
)

var loginPattern = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,100}$`)

func validateLogin(vErr *usererrors.ValidationError, login string) {
	if !loginPattern.MatchString(login) {
		vErr.Add("login", "deve ter de 3 a 100 caracteres (letras, números, '.', '_' ou '-')")
	}
}

func validateName(vErr *usererrors.ValidationError, name string) {
	if name == "" {
		vErr.Add("name", "obrigatório")
	} else if utf8.RuneCountInString(name) > 100 {
		vErr.Add("name", "deve ter no máximo 100 caracteres")
	}
}

// Email é obrigatório: a coluna tem índice único e o tipo string grava ""
// quando vazio, o que faria dois usuários sem email colidirem.
func validateEmail(vErr *usererrors.ValidationError, email string) {
	if email == "" {
		vErr.Add("email", "obrigatório")
		return
	}
	if utf8.RuneCountInString(email) > 150 {
		vErr.Add("email", "deve ter no máximo 150 caracteres")
		return
	}
	if addr, err := mail.ParseAddress(email); err != nil || addr.Address != email {
		vErr.Add("email", "formato inválido")
	}
}

func validatePassword(vErr *usererrors.ValidationError, password string) {
	n := utf8.RuneCountInString(password)
	if n < minPasswordLength {
		vErr.Add("password", "deve ter no mínimo 8 caracteres")
	} else if n > maxPasswordLength {
		vErr.Add("password", "deve ter no máximo 128 caracteres")
	}
}

func validateRole(vErr *usererrors.ValidationError, role model.Role) {
	if !role.IsValid() {
		vErr.Add("role", "deve ser 'admin' ou 'user'")
	}
}
