package usecase

import "github.com/djsilvajr/simple-project-template/back/internal/domain/user/model"

const (
	defaultPerPage = 20
	maxPerPage     = 100
)

// NormalizePagination aplica os valores padrão e o limite de itens por página.
func NormalizePagination(input model.ListUsersInput) model.ListUsersInput {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PerPage < 1 {
		input.PerPage = defaultPerPage
	}
	if input.PerPage > maxPerPage {
		input.PerPage = maxPerPage
	}
	return input
}
